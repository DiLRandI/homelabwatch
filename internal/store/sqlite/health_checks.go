package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
	"github.com/deleema/homelabwatch/internal/servicedefs"
)

func (s *Store) ListServiceChecks(ctx context.Context, serviceID string) ([]domain.ServiceCheck, error) {
	return s.listChecksBySubject(ctx, domain.HealthCheckSubjectService, serviceID)
}

func (s *Store) ListDiscoveredServiceChecks(ctx context.Context, discoveredServiceID string) ([]domain.ServiceCheck, error) {
	return s.listChecksBySubject(ctx, domain.HealthCheckSubjectDiscoveredService, discoveredServiceID)
}

func (s *Store) listChecksBySubject(ctx context.Context, subjectType domain.HealthCheckSubjectType, subjectID string) ([]domain.ServiceCheck, error) {
	primaryAddresses, err := s.loadPrimaryDeviceAddresses(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			h.id,
			h.subject_type,
			h.subject_id,
			h.name,
			h.kind,
			COALESCE(h.protocol, ''),
			COALESCE(h.address_source, 'literal_host'),
			COALESCE(h.host_value, ''),
			h.port,
			COALESCE(h.path, ''),
			COALESCE(h.method, ''),
			COALESCE(h.target, ''),
			h.interval_seconds,
			h.timeout_seconds,
			h.expected_status_min,
			h.expected_status_max,
			h.enabled,
			COALESCE(h.sort_order, 0),
			COALESCE(h.config_source, 'fallback'),
			COALESCE(h.service_definition_id, ''),
			h.created_at,
			h.updated_at,
			COALESCE(r.id, ''),
			COALESCE(r.status, ''),
			COALESCE(r.latency_ms, 0),
			COALESCE(r.http_status_code, 0),
			COALESCE(r.response_size_bytes, 0),
			COALESCE(r.message, ''),
			COALESCE(r.checked_at, ''),
			COALESCE(svc.device_id, dsvc.device_id, '')
		FROM health_checks h
		LEFT JOIN services svc ON h.subject_type = 'service' AND svc.id = h.subject_id
		LEFT JOIN discovered_services dsvc ON h.subject_type = 'discovered_service' AND dsvc.id = h.subject_id
		LEFT JOIN (
			SELECT hcr1.*
			FROM health_check_results hcr1
			JOIN (
				SELECT health_check_id, MAX(checked_at) AS checked_at
				FROM health_check_results
				GROUP BY health_check_id
			) latest ON latest.health_check_id = hcr1.health_check_id AND latest.checked_at = hcr1.checked_at
		) r ON r.health_check_id = h.id
		WHERE h.subject_type = ? AND h.subject_id = ?
		ORDER BY h.sort_order, h.name
	`, subjectType, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.ServiceCheck{}
	for rows.Next() {
		item, deviceID, err := scanHealthCheck(rows)
		if err != nil {
			return nil, err
		}
		item.Host = resolveAddressSourceHost(item.AddressSource, item.HostValue, primaryAddresses[deviceID])
		item.Target = servicedefs.ResolveCheckTarget(item)
		if item.SubjectType == domain.HealthCheckSubjectService {
			item.ServiceID = item.SubjectID
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) SaveServiceCheck(ctx context.Context, check domain.ServiceCheck) (domain.ServiceCheck, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.ServiceCheck{}, err
	}
	defer tx.Rollback()

	if check.SubjectType == "" {
		check.SubjectType = domain.HealthCheckSubjectService
	}
	if check.SubjectID == "" {
		check.SubjectID = check.ServiceID
	}
	if check.SubjectID == "" && check.ID != "" {
		existing, _, err := s.getHealthCheckTx(ctx, tx, check.ID)
		if err != nil {
			return domain.ServiceCheck{}, err
		}
		check.SubjectType = existing.SubjectType
		check.SubjectID = existing.SubjectID
		check.ServiceID = existing.ServiceID
		if check.CreatedAt.IsZero() {
			check.CreatedAt = existing.CreatedAt
		}
	}
	if check.SubjectType != domain.HealthCheckSubjectService {
		return domain.ServiceCheck{}, errors.New("service check subject must be service")
	}
	if check.SubjectID == "" {
		return domain.ServiceCheck{}, errors.New("service id is required")
	}
	if err := s.saveHealthCheckTx(ctx, tx, &check, true); err != nil {
		return domain.ServiceCheck{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.ServiceCheck{}, err
	}
	items, err := s.ListServiceChecks(ctx, check.SubjectID)
	if err != nil {
		return domain.ServiceCheck{}, err
	}
	for _, item := range items {
		if item.ID == check.ID {
			return item, nil
		}
	}
	return domain.ServiceCheck{}, sql.ErrNoRows
}

func (s *Store) DeleteServiceCheck(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	check, _, err := s.getHealthCheckTx(ctx, tx, id)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM health_checks WHERE id = ?`, id); err != nil {
		return err
	}
	if check.SubjectType == domain.HealthCheckSubjectService {
		if _, err := tx.ExecContext(ctx, `UPDATE services SET health_config_mode = ?, updated_at = ? WHERE id = ?`, domain.HealthConfigModeCustom, nowString(), check.SubjectID); err != nil {
			return err
		}
		status, err := s.rollupServiceStatusTx(ctx, tx, check.SubjectID)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE services SET status = ?, updated_at = ? WHERE id = ?`, status, nowString(), check.SubjectID); err != nil {
			return err
		}
	}
	if check.SubjectType == domain.HealthCheckSubjectDiscoveredService {
		status, err := s.rollupDiscoveredServiceStatusTx(ctx, tx, check.SubjectID)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE discovered_services SET status = ?, updated_at = ? WHERE id = ?`, status, nowString(), check.SubjectID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) SaveCheckResult(ctx context.Context, result domain.CheckResult) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	check, intervalSeconds, err := s.getHealthCheckTx(ctx, tx, result.CheckID)
	if err != nil {
		return err
	}
	if result.ID == "" {
		result.ID = newID("hcr")
	}
	if result.CheckedAt.IsZero() {
		result.CheckedAt = time.Now().UTC()
	}
	result.SubjectType = check.SubjectType
	result.SubjectID = check.SubjectID
	result.ServiceID = check.SubjectID
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO health_check_results(
			id, health_check_id, subject_type, subject_id, status, latency_ms, message, checked_at, http_status_code, response_size_bytes
		) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, result.ID, result.CheckID, result.SubjectType, result.SubjectID, result.Status, result.LatencyMS, result.Message, result.CheckedAt.Format(time.RFC3339Nano), result.HTTPStatusCode, result.ResponseSizeBytes); err != nil {
		return err
	}

	consecutiveFailures := 0
	var currentFailures int
	var currentLastStatus string
	if err := tx.QueryRowContext(ctx, `SELECT consecutive_failures, COALESCE(last_status, 'unknown') FROM health_checks WHERE id = ?`, check.ID).Scan(&currentFailures, &currentLastStatus); err != nil {
		return err
	}
	if result.Status != domain.HealthStatusHealthy {
		if currentLastStatus != string(domain.HealthStatusHealthy) {
			consecutiveFailures = currentFailures
		}
		consecutiveFailures++
	}
	nextRunAt := result.CheckedAt.Add(time.Duration(defaultCheckInterval(intervalSeconds)) * time.Second)
	if _, err := tx.ExecContext(ctx, `
		UPDATE health_checks
		SET target = ?, last_run_at = ?, last_status = ?, consecutive_failures = ?, next_run_at = ?, updated_at = ?
		WHERE id = ?
	`, servicedefs.ResolveCheckTarget(check), result.CheckedAt.Format(time.RFC3339Nano), result.Status, consecutiveFailures, nextRunAt.Format(time.RFC3339Nano), nowString(), check.ID); err != nil {
		return err
	}

	switch check.SubjectType {
	case domain.HealthCheckSubjectService:
		var previousStatus string
		if err := tx.QueryRowContext(ctx, `SELECT status FROM services WHERE id = ?`, check.SubjectID).Scan(&previousStatus); err != nil {
			return err
		}
		nextStatus, err := s.rollupServiceStatusTx(ctx, tx, check.SubjectID)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE services SET status = ?, last_checked_at = ?, updated_at = ? WHERE id = ?`, nextStatus, result.CheckedAt.Format(time.RFC3339Nano), nowString(), check.SubjectID); err != nil {
			return err
		}
		if previousStatus != string(nextStatus) {
			if err := s.insertServiceEventTx(ctx, tx, check.SubjectID, "health_changed", nextStatus, firstNonEmpty(result.Message, fmt.Sprintf("service status changed to %s", nextStatus))); err != nil {
				return err
			}
		}
	case domain.HealthCheckSubjectDiscoveredService:
		nextStatus, err := s.rollupDiscoveredServiceStatusTx(ctx, tx, check.SubjectID)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE discovered_services SET status = ?, last_checked_at = ?, updated_at = ? WHERE id = ?`, nextStatus, result.CheckedAt.Format(time.RFC3339Nano), nowString(), check.SubjectID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) GetChecksDue(ctx context.Context) ([]domain.MonitorCheck, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT h.id, h.subject_id
		FROM health_checks h
		WHERE h.enabled = 1
		  AND h.subject_type = 'service'
		  AND (COALESCE(h.next_run_at, '') = '' OR h.next_run_at <= ?)
		ORDER BY h.next_run_at, h.sort_order, h.name
	`, nowString())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.MonitorCheck{}
	for rows.Next() {
		var checkID, serviceID string
		if err := rows.Scan(&checkID, &serviceID); err != nil {
			return nil, err
		}
		service, err := s.GetService(ctx, serviceID)
		if err != nil {
			return nil, err
		}
		var check domain.ServiceCheck
		for _, item := range service.Checks {
			if item.ID == checkID {
				check = item
				break
			}
		}
		if check.ID == "" {
			return nil, fmt.Errorf("health check %s not found on service %s", checkID, serviceID)
		}
		items = append(items, domain.MonitorCheck{Check: check, Service: service})
	}
	return items, rows.Err()
}

func (s *Store) GetDiscoveredChecksDue(ctx context.Context) ([]domain.ServiceCheck, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT h.id, h.subject_id
		FROM health_checks h
		WHERE h.enabled = 1
		  AND h.subject_type = 'discovered_service'
		  AND (COALESCE(h.next_run_at, '') = '' OR h.next_run_at <= ?)
		ORDER BY h.next_run_at, h.sort_order, h.name
	`, nowString())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.ServiceCheck{}
	for rows.Next() {
		var checkID, discoveredID string
		if err := rows.Scan(&checkID, &discoveredID); err != nil {
			return nil, err
		}
		checks, err := s.ListDiscoveredServiceChecks(ctx, discoveredID)
		if err != nil {
			return nil, err
		}
		for _, check := range checks {
			if check.ID == checkID && check.Enabled {
				items = append(items, check)
				break
			}
		}
	}
	return items, rows.Err()
}

func (s *Store) Cleanup(ctx context.Context, retain time.Duration) error {
	cutoff := time.Now().UTC().Add(-retain).Format(time.RFC3339Nano)
	if _, err := s.db.ExecContext(ctx, "DELETE FROM health_check_results WHERE checked_at < ?", cutoff); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, "DELETE FROM check_results WHERE checked_at < ?", cutoff); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, "DELETE FROM service_events WHERE created_at < ?", cutoff); err != nil {
		return err
	}
	return nil
}

func (s *Store) EnsureDefaultChecksForService(ctx context.Context, service domain.Service) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := s.ensureDefaultCheckTx(ctx, tx, service); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) SyncServiceHealthChecks(ctx context.Context, serviceID string) error {
	service, err := s.GetService(ctx, serviceID)
	if err != nil {
		return err
	}
	if service.HealthConfigMode == domain.HealthConfigModeCustom {
		return nil
	}
	if service.ServiceDefinitionID != "" {
		definitions, err := s.ListServiceDefinitions(ctx)
		if err != nil {
			return err
		}
		for _, definition := range definitions {
			if definition.ID != service.ServiceDefinitionID {
				continue
			}
			checks := servicedefs.InstantiateChecks(domain.HealthCheckSubjectService, service.ID, service.AddressSource, service.HostValue, service.Host, service.Scheme, service.Port, service.Path, definition)
			return s.ReplaceManagedChecks(ctx, domain.HealthCheckSubjectService, service.ID, checks, definition.ID, domain.HealthConfigModeAuto)
		}
	}
	check := domain.ServiceCheck{
		SubjectType:     domain.HealthCheckSubjectService,
		SubjectID:       service.ID,
		ServiceID:       service.ID,
		AddressSource:   service.AddressSource,
		HostValue:       service.HostValue,
		Host:            service.Host,
		Protocol:        service.Scheme,
		Port:            service.Port,
		Path:            service.Path,
		Enabled:         true,
		IntervalSeconds: 60,
		TimeoutSeconds:  10,
		SortOrder:       0,
		ConfigSource:    domain.HealthCheckConfigSourceFallback,
	}
	switch {
	case service.Path != "" && (service.Scheme == "http" || service.Scheme == "https"):
		check.Name = "HTTP endpoint"
		check.Type = domain.CheckTypeHTTP
		check.Method = "GET"
		check.ExpectedStatusMin = 200
		check.ExpectedStatusMax = 399
	case service.Host != "" && service.Port > 0:
		check.Name = "TCP connectivity"
		check.Type = domain.CheckTypeTCP
	default:
		check.Name = "Ping reachability"
		check.Type = domain.CheckTypePing
	}
	return s.ReplaceManagedChecks(ctx, domain.HealthCheckSubjectService, service.ID, []domain.ServiceCheck{check}, "", domain.HealthConfigModeAuto)
}

func (s *Store) ReplaceManagedChecks(ctx context.Context, subjectType domain.HealthCheckSubjectType, subjectID string, checks []domain.ServiceCheck, serviceDefinitionID string, configMode domain.HealthConfigMode) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM health_checks WHERE subject_type = ? AND subject_id = ? AND config_source <> ?`, subjectType, subjectID, domain.HealthCheckConfigSourceUser); err != nil {
		return err
	}
	for index := range checks {
		checks[index].SubjectType = subjectType
		checks[index].SubjectID = subjectID
		checks[index].ServiceID = subjectID
		checks[index].ServiceDefinitionID = serviceDefinitionID
		if err := s.saveHealthCheckTx(ctx, tx, &checks[index], false); err != nil {
			return err
		}
	}
	switch subjectType {
	case domain.HealthCheckSubjectService:
		_, err = tx.ExecContext(ctx, `UPDATE services SET service_definition_id = ?, health_config_mode = ?, updated_at = ? WHERE id = ?`, nullableString(serviceDefinitionID), configMode, nowString(), subjectID)
	case domain.HealthCheckSubjectDiscoveredService:
		_, err = tx.ExecContext(ctx, `UPDATE discovered_services SET service_definition_id = ?, health_config_mode = ?, updated_at = ? WHERE id = ?`, nullableString(serviceDefinitionID), configMode, nowString(), subjectID)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) CopyDiscoveredChecksToService(ctx context.Context, discoveredServiceID, serviceID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var definitionID, healthConfigMode string
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(service_definition_id, ''), COALESCE(health_config_mode, 'auto') FROM discovered_services WHERE id = ?`, discoveredServiceID).Scan(&definitionID, &healthConfigMode); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM health_checks WHERE subject_type = ? AND subject_id = ? AND config_source <> ?`, domain.HealthCheckSubjectService, serviceID, domain.HealthCheckConfigSourceUser); err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT id, subject_type, subject_id, name, kind, COALESCE(protocol, ''), COALESCE(address_source, 'literal_host'), COALESCE(host_value, ''), port, COALESCE(path, ''), COALESCE(method, ''), COALESCE(target, ''), interval_seconds, timeout_seconds, expected_status_min, expected_status_max, enabled, COALESCE(sort_order, 0), COALESCE(config_source, 'fallback'), COALESCE(service_definition_id, ''), created_at, updated_at
		FROM health_checks
		WHERE subject_type = ? AND subject_id = ?
		ORDER BY sort_order, name
	`, domain.HealthCheckSubjectDiscoveredService, discoveredServiceID)
	if err != nil {
		return err
	}
	defer rows.Close()

	oldToNew := map[string]string{}
	for rows.Next() {
		check, err := scanHealthCheckRow(rows)
		if err != nil {
			return err
		}
		oldToNew[check.ID] = newID("hck")
		check.ID = oldToNew[check.ID]
		check.SubjectType = domain.HealthCheckSubjectService
		check.SubjectID = serviceID
		check.ServiceID = serviceID
		if err := s.saveHealthCheckTx(ctx, tx, &check, false); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for oldID, newIDValue := range oldToNew {
		results, err := tx.QueryContext(ctx, `
			SELECT id, health_check_id, subject_type, subject_id, status, latency_ms, COALESCE(message, ''), checked_at, COALESCE(http_status_code, 0), COALESCE(response_size_bytes, 0)
			FROM health_check_results
			WHERE health_check_id = ?
			ORDER BY checked_at DESC
			LIMIT 100
		`, oldID)
		if err != nil {
			return err
		}
		for results.Next() {
			var item domain.CheckResult
			var checkedAt string
			if err := results.Scan(&item.ID, &item.CheckID, &item.SubjectType, &item.SubjectID, &item.Status, &item.LatencyMS, &item.Message, &checkedAt, &item.HTTPStatusCode, &item.ResponseSizeBytes); err != nil {
				_ = results.Close()
				return err
			}
			item.ID = newID("hcr")
			item.CheckID = newIDValue
			item.SubjectType = domain.HealthCheckSubjectService
			item.SubjectID = serviceID
			item.ServiceID = serviceID
			item.CheckedAt = parseTime(checkedAt)
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO health_check_results(id, health_check_id, subject_type, subject_id, status, latency_ms, message, checked_at, http_status_code, response_size_bytes)
				VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, item.ID, item.CheckID, item.SubjectType, item.SubjectID, item.Status, item.LatencyMS, item.Message, item.CheckedAt.Format(time.RFC3339Nano), item.HTTPStatusCode, item.ResponseSizeBytes); err != nil {
				_ = results.Close()
				return err
			}
		}
		if err := results.Close(); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE services SET service_definition_id = ?, health_config_mode = ?, updated_at = ? WHERE id = ?`, nullableString(definitionID), healthConfigMode, nowString(), serviceID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) saveHealthCheckTx(ctx context.Context, tx *sql.Tx, check *domain.ServiceCheck, markCustom bool) error {
	now := time.Now().UTC()
	if check.ID == "" {
		check.ID = newID("hck")
		check.CreatedAt = now
	}
	if check.CreatedAt.IsZero() {
		existing, _, err := s.getHealthCheckTx(ctx, tx, check.ID)
		if err == nil {
			check.CreatedAt = existing.CreatedAt
		} else {
			check.CreatedAt = now
		}
	}
	if check.SubjectType == "" {
		check.SubjectType = domain.HealthCheckSubjectService
	}
	if check.SubjectID == "" {
		check.SubjectID = check.ServiceID
	}
	if check.ServiceID == "" {
		check.ServiceID = check.SubjectID
	}
	if check.Name == "" {
		check.Name = defaultCheckName(*check)
	}
	if check.IntervalSeconds <= 0 {
		check.IntervalSeconds = 60
	}
	if check.TimeoutSeconds <= 0 {
		check.TimeoutSeconds = 10
	}
	if check.Type == domain.CheckTypeHTTP {
		if check.Method == "" {
			check.Method = "GET"
		}
		if check.ExpectedStatusMin == 0 {
			check.ExpectedStatusMin = 200
		}
		if check.ExpectedStatusMax == 0 {
			check.ExpectedStatusMax = 399
		}
	}
	if check.ConfigSource == "" {
		if markCustom {
			check.ConfigSource = domain.HealthCheckConfigSourceUser
		} else {
			check.ConfigSource = domain.HealthCheckConfigSourceFallback
		}
	}
	if check.AddressSource == "" {
		check.AddressSource = domain.ServiceAddressLiteralHost
	}
	if check.HostValue == "" {
		check.HostValue = firstNonEmpty(check.HostValue, check.Host)
	}
	check.Target = servicedefs.ResolveCheckTarget(*check)
	check.UpdatedAt = now
	nextRunAt := now
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO health_checks(
			id, subject_type, subject_id, kind, target, interval_seconds, timeout_seconds, expected_status_min, expected_status_max,
			enabled, next_run_at, last_run_at, last_status, consecutive_failures, created_at, updated_at, name, sort_order,
			protocol, address_source, host_value, port, path, method, config_source, service_definition_id
		) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, 'unknown', 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			subject_type = excluded.subject_type,
			subject_id = excluded.subject_id,
			kind = excluded.kind,
			target = excluded.target,
			interval_seconds = excluded.interval_seconds,
			timeout_seconds = excluded.timeout_seconds,
			expected_status_min = excluded.expected_status_min,
			expected_status_max = excluded.expected_status_max,
			enabled = excluded.enabled,
			next_run_at = excluded.next_run_at,
			updated_at = excluded.updated_at,
			name = excluded.name,
			sort_order = excluded.sort_order,
			protocol = excluded.protocol,
			address_source = excluded.address_source,
			host_value = excluded.host_value,
			port = excluded.port,
			path = excluded.path,
			method = excluded.method,
			config_source = excluded.config_source,
			service_definition_id = excluded.service_definition_id
	`, check.ID, check.SubjectType, check.SubjectID, check.Type, check.Target, check.IntervalSeconds, check.TimeoutSeconds, check.ExpectedStatusMin, check.ExpectedStatusMax, boolInt(check.Enabled), nextRunAt.Format(time.RFC3339Nano), check.CreatedAt.Format(time.RFC3339Nano), check.UpdatedAt.Format(time.RFC3339Nano), check.Name, check.SortOrder, check.Protocol, check.AddressSource, check.HostValue, check.Port, check.Path, check.Method, check.ConfigSource, nullableString(check.ServiceDefinitionID)); err != nil {
		return err
	}
	if markCustom {
		switch check.SubjectType {
		case domain.HealthCheckSubjectService:
			if _, err := tx.ExecContext(ctx, `UPDATE services SET health_config_mode = ?, updated_at = ? WHERE id = ?`, domain.HealthConfigModeCustom, nowString(), check.SubjectID); err != nil {
				return err
			}
		case domain.HealthCheckSubjectDiscoveredService:
			if _, err := tx.ExecContext(ctx, `UPDATE discovered_services SET health_config_mode = ?, updated_at = ? WHERE id = ?`, domain.HealthConfigModeCustom, nowString(), check.SubjectID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Store) getHealthCheckTx(ctx context.Context, tx *sql.Tx, id string) (domain.ServiceCheck, int, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT
			id, subject_type, subject_id, name, kind, COALESCE(protocol, ''), COALESCE(address_source, 'literal_host'),
			COALESCE(host_value, ''), port, COALESCE(path, ''), COALESCE(method, ''), COALESCE(target, ''),
			interval_seconds, timeout_seconds, expected_status_min, expected_status_max, enabled, COALESCE(sort_order, 0),
			COALESCE(config_source, 'fallback'), COALESCE(service_definition_id, ''), created_at, updated_at
		FROM health_checks
		WHERE id = ?
	`, id)
	check, err := scanHealthCheckRow(row)
	if err != nil {
		return domain.ServiceCheck{}, 0, err
	}
	return check, check.IntervalSeconds, nil
}

func (s *Store) rollupServiceStatusTx(ctx context.Context, tx *sql.Tx, serviceID string) (domain.HealthStatus, error) {
	return s.rollupSubjectStatusTx(ctx, tx, domain.HealthCheckSubjectService, serviceID)
}

func (s *Store) rollupDiscoveredServiceStatusTx(ctx context.Context, tx *sql.Tx, discoveredServiceID string) (domain.HealthStatus, error) {
	return s.rollupSubjectStatusTx(ctx, tx, domain.HealthCheckSubjectDiscoveredService, discoveredServiceID)
}

func (s *Store) rollupSubjectStatusTx(ctx context.Context, tx *sql.Tx, subjectType domain.HealthCheckSubjectType, subjectID string) (domain.HealthStatus, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT COALESCE(r.status, 'unknown')
		FROM health_checks h
		LEFT JOIN (
			SELECT hcr1.*
			FROM health_check_results hcr1
			JOIN (
				SELECT health_check_id, MAX(checked_at) AS checked_at
				FROM health_check_results
				GROUP BY health_check_id
			) latest ON latest.health_check_id = hcr1.health_check_id AND latest.checked_at = hcr1.checked_at
		) r ON r.health_check_id = h.id
		WHERE h.subject_type = ? AND h.subject_id = ? AND h.enabled = 1
	`, subjectType, subjectID)
	if err != nil {
		return domain.HealthStatusUnknown, err
	}
	defer rows.Close()
	statuses := []domain.HealthStatus{}
	for rows.Next() {
		var status string
		if err := rows.Scan(&status); err != nil {
			return domain.HealthStatusUnknown, err
		}
		statuses = append(statuses, domain.HealthStatus(status))
	}
	if len(statuses) == 0 {
		return domain.HealthStatusUnknown, nil
	}
	allHealthy := true
	allUnhealthy := true
	for _, status := range statuses {
		switch status {
		case domain.HealthStatusHealthy:
			allUnhealthy = false
		case domain.HealthStatusUnhealthy:
			allHealthy = false
		default:
			allHealthy = false
			allUnhealthy = false
		}
	}
	if allHealthy {
		return domain.HealthStatusHealthy, nil
	}
	if allUnhealthy {
		return domain.HealthStatusUnhealthy, nil
	}
	return domain.HealthStatusDegraded, nil
}

func (s *Store) ensureDefaultCheckTx(ctx context.Context, tx *sql.Tx, service domain.Service) error {
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(1) FROM health_checks WHERE subject_type = ? AND subject_id = ?`, domain.HealthCheckSubjectService, service.ID).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	check := domain.ServiceCheck{
		SubjectType:     domain.HealthCheckSubjectService,
		SubjectID:       service.ID,
		ServiceID:       service.ID,
		AddressSource:   service.AddressSource,
		HostValue:       service.HostValue,
		Host:            service.Host,
		Protocol:        service.Scheme,
		Port:            service.Port,
		Path:            service.Path,
		Enabled:         true,
		IntervalSeconds: 60,
		TimeoutSeconds:  10,
		SortOrder:       0,
	}
	switch {
	case service.Path != "" && (service.Scheme == "http" || service.Scheme == "https"):
		check.Name = "HTTP endpoint"
		check.Type = domain.CheckTypeHTTP
		check.Method = "GET"
		check.ExpectedStatusMin = 200
		check.ExpectedStatusMax = 399
		check.ConfigSource = domain.HealthCheckConfigSourceFallback
	case service.Host != "" && service.Port > 0:
		check.Name = "TCP connectivity"
		check.Type = domain.CheckTypeTCP
		check.ConfigSource = domain.HealthCheckConfigSourceFallback
	default:
		check.Name = "Ping reachability"
		check.Type = domain.CheckTypePing
		check.ConfigSource = domain.HealthCheckConfigSourceFallback
	}
	return s.saveHealthCheckTx(ctx, tx, &check, false)
}

func scanHealthCheck(scanner interface{ Scan(dest ...any) error }) (domain.ServiceCheck, string, error) {
	var (
		check                                                  domain.ServiceCheck
		subjectType, addressSource, configSource               string
		enabled                                                int
		createdAt, updatedAt                                   string
		resultID, resultStatus, resultMessage, resultCheckedAt string
		resultLatency                                          int64
		resultHTTPStatusCode                                   int
		resultResponseSizeBytes                                int64
		deviceID                                               string
	)
	if err := scanner.Scan(
		&check.ID,
		&subjectType,
		&check.SubjectID,
		&check.Name,
		&check.Type,
		&check.Protocol,
		&addressSource,
		&check.HostValue,
		&check.Port,
		&check.Path,
		&check.Method,
		&check.Target,
		&check.IntervalSeconds,
		&check.TimeoutSeconds,
		&check.ExpectedStatusMin,
		&check.ExpectedStatusMax,
		&enabled,
		&check.SortOrder,
		&configSource,
		&check.ServiceDefinitionID,
		&createdAt,
		&updatedAt,
		&resultID,
		&resultStatus,
		&resultLatency,
		&resultHTTPStatusCode,
		&resultResponseSizeBytes,
		&resultMessage,
		&resultCheckedAt,
		&deviceID,
	); err != nil {
		return domain.ServiceCheck{}, "", err
	}
	check.SubjectType = domain.HealthCheckSubjectType(subjectType)
	check.AddressSource = domain.ServiceAddressSource(addressSource)
	check.ConfigSource = domain.HealthCheckConfigSource(configSource)
	check.Enabled = enabled == 1
	check.CreatedAt = parseTime(createdAt)
	check.UpdatedAt = parseTime(updatedAt)
	check.ServiceID = check.SubjectID
	if resultID != "" {
		check.LastResult = &domain.CheckResult{
			ID:                resultID,
			CheckID:           check.ID,
			ServiceID:         check.SubjectID,
			SubjectType:       check.SubjectType,
			SubjectID:         check.SubjectID,
			Status:            domain.HealthStatus(resultStatus),
			LatencyMS:         resultLatency,
			HTTPStatusCode:    resultHTTPStatusCode,
			ResponseSizeBytes: resultResponseSizeBytes,
			Message:           resultMessage,
			CheckedAt:         parseTime(resultCheckedAt),
		}
	}
	return check, deviceID, nil
}

func scanHealthCheckRow(scanner interface{ Scan(dest ...any) error }) (domain.ServiceCheck, error) {
	var (
		check                                    domain.ServiceCheck
		subjectType, addressSource, configSource string
		enabled                                  int
		createdAt, updatedAt                     string
	)
	if err := scanner.Scan(
		&check.ID,
		&subjectType,
		&check.SubjectID,
		&check.Name,
		&check.Type,
		&check.Protocol,
		&addressSource,
		&check.HostValue,
		&check.Port,
		&check.Path,
		&check.Method,
		&check.Target,
		&check.IntervalSeconds,
		&check.TimeoutSeconds,
		&check.ExpectedStatusMin,
		&check.ExpectedStatusMax,
		&enabled,
		&check.SortOrder,
		&configSource,
		&check.ServiceDefinitionID,
		&createdAt,
		&updatedAt,
	); err != nil {
		return domain.ServiceCheck{}, err
	}
	check.SubjectType = domain.HealthCheckSubjectType(subjectType)
	check.ServiceID = check.SubjectID
	check.AddressSource = domain.ServiceAddressSource(addressSource)
	check.ConfigSource = domain.HealthCheckConfigSource(configSource)
	check.Enabled = enabled == 1
	check.CreatedAt = parseTime(createdAt)
	check.UpdatedAt = parseTime(updatedAt)
	return check, nil
}

func defaultCheckName(check domain.ServiceCheck) string {
	switch check.Type {
	case domain.CheckTypeHTTP:
		return "HTTP endpoint"
	case domain.CheckTypeTCP:
		return "TCP connectivity"
	case domain.CheckTypePing:
		return "Ping reachability"
	default:
		return "Health check"
	}
}

func defaultCheckInterval(value int) int {
	if value > 0 {
		return value
	}
	return 60
}
