package sqlite

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
	"github.com/deleema/homelabwatch/internal/servicedefs"
)

type discoveredServiceRecord struct {
	ID                  string
	DeviceID            string
	DeviceName          string
	MergeKey            string
	Name                string
	ServiceType         string
	ConfidenceScore     int
	ServiceDefinitionID string
	AddressSource       domain.ServiceAddressSource
	HostValue           string
	Scheme              string
	Port                int
	Path                string
	URL                 string
	Icon                string
	State               domain.DiscoveryState
	IgnoreFingerprint   string
	AutomationMode      domain.BookmarkAutomationPolicy
	HealthConfigMode    domain.HealthConfigMode
	Status              domain.HealthStatus
	AcceptedServiceID   string
	AcceptedBookmarkID  string
	LastCheckedAt       time.Time
	LastFingerprintedAt time.Time
	Details             map[string]any
	FirstSeenAt         time.Time
	LastSeenAt          time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type discoveryFingerprint struct {
	ServiceType   string
	Name          string
	Confidence    int
	AddressSource domain.ServiceAddressSource
	HostValue     string
	Icon          string
	Details       map[string]any
	Hash          string
}

func (s *Store) GetDiscoverySettings(ctx context.Context) (domain.DiscoverySettings, error) {
	settings, err := s.GetAppSettings(ctx)
	if err != nil {
		return domain.DiscoverySettings{}, err
	}
	return domain.DiscoverySettings{
		BookmarkPolicy:            settings.BookmarkPolicy,
		AutoBookmarkSources:       append([]domain.ServiceSource(nil), settings.AutoBookmarkSources...),
		AutoBookmarkMinConfidence: settings.AutoBookmarkMinConfidence,
	}, nil
}

func (s *Store) SaveDiscoverySettings(ctx context.Context, input domain.DiscoverySettings) (domain.DiscoverySettings, error) {
	if input.BookmarkPolicy == "" {
		input.BookmarkPolicy = domain.BookmarkAutomationManual
	}
	if input.BookmarkPolicy != domain.BookmarkAutomationManual && input.BookmarkPolicy != domain.BookmarkAutomationAutoHighConfidence {
		return domain.DiscoverySettings{}, errors.New("bookmark policy must be manual or auto_high_confidence")
	}
	if len(input.AutoBookmarkSources) == 0 {
		input.AutoBookmarkSources = []domain.ServiceSource{
			domain.ServiceSourceDocker,
			domain.ServiceSourceLAN,
			domain.ServiceSourceMDNS,
		}
	}
	if input.AutoBookmarkMinConfidence <= 0 {
		input.AutoBookmarkMinConfidence = 90
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.DiscoverySettings{}, err
	}
	defer tx.Rollback()

	now := nowString()
	settings := map[string]string{
		"bookmark_policy":              string(input.BookmarkPolicy),
		"auto_bookmark_sources":        string(mustJSON(input.AutoBookmarkSources)),
		"auto_bookmark_min_confidence": strconv.Itoa(input.AutoBookmarkMinConfidence),
	}
	for key, value := range settings {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO app_settings(key, value, updated_at)
			VALUES(?, ?, ?)
			ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
		`, key, value, now); err != nil {
			return domain.DiscoverySettings{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return domain.DiscoverySettings{}, err
	}
	return s.GetDiscoverySettings(ctx)
}

func (s *Store) UpsertDiscoveredServiceObservation(ctx context.Context, observation domain.ServiceObservation, deviceID string) (domain.DiscoveredService, error) {
	outcome, err := s.UpsertDiscoveredServiceObservationWithOutcome(ctx, observation, deviceID)
	return outcome.DiscoveredService, err
}

func (s *Store) UpsertDiscoveredServiceObservationWithOutcome(ctx context.Context, observation domain.ServiceObservation, deviceID string) (domain.DiscoveredServiceObservationOutcome, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.DiscoveredServiceObservationOutcome{}, err
	}
	defer tx.Rollback()

	now := observation.LastSeenAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	fingerprint := fingerprintObservation(observation, deviceID)
	mergeKey := discoveredMergeKey(deviceID, fingerprint.HostValue, observation.Port, observation.Path, fingerprint.ServiceType, observation.Name)

	record, found, err := s.lookupDiscoveredServiceByMergeKeyTx(ctx, tx, mergeKey)
	if err != nil {
		return domain.DiscoveredServiceObservationOutcome{}, err
	}
	if !found {
		record = discoveredServiceRecord{
			ID:                  newID("dsvc"),
			DeviceID:            deviceID,
			MergeKey:            mergeKey,
			Name:                fingerprint.Name,
			ServiceType:         fingerprint.ServiceType,
			ConfidenceScore:     fingerprint.Confidence,
			ServiceDefinitionID: "",
			AddressSource:       fingerprint.AddressSource,
			HostValue:           fingerprint.HostValue,
			Scheme:              firstNonEmpty(observation.Scheme, schemeFromObservation(observation, fingerprint)),
			Port:                observation.Port,
			Path:                normalizePath(observation.Path),
			URL:                 firstNonEmpty(observation.URL, buildServiceURL(firstNonEmpty(observation.Scheme, schemeFromObservation(observation, fingerprint)), firstNonEmpty(observation.Host, fingerprint.HostValue), observation.Port, observation.Path)),
			Icon:                firstNonEmpty(observation.Icon, fingerprint.Icon),
			State:               domain.DiscoveryStatePending,
			AutomationMode:      domain.BookmarkAutomationManual,
			HealthConfigMode:    domain.HealthConfigModeAuto,
			Status:              domain.HealthStatusUnknown,
			Details:             fingerprint.Details,
			FirstSeenAt:         now,
			LastSeenAt:          now,
			CreatedAt:           now,
			UpdatedAt:           now,
		}
	} else {
		record.DeviceID = firstNonEmpty(deviceID, record.DeviceID)
		record.Name = firstNonEmpty(fingerprint.Name, record.Name)
		record.ServiceType = firstNonEmpty(fingerprint.ServiceType, record.ServiceType)
		record.ConfidenceScore = max(record.ConfidenceScore, fingerprint.Confidence)
		record.AddressSource = mergeDiscoveryAddressSource(record.AddressSource, fingerprint.AddressSource)
		if record.AddressSource == domain.ServiceAddressDevicePrimary {
			record.HostValue = firstNonEmpty(record.HostValue, fingerprint.HostValue)
		} else {
			record.HostValue = firstNonEmpty(fingerprint.HostValue, record.HostValue)
		}
		record.Scheme = firstNonEmpty(observation.Scheme, record.Scheme, schemeFromObservation(observation, fingerprint))
		if observation.Port > 0 {
			record.Port = observation.Port
		}
		record.Path = firstNonEmpty(normalizePath(observation.Path), record.Path)
		record.URL = firstNonEmpty(observation.URL, record.URL)
		record.Icon = firstNonEmpty(observation.Icon, fingerprint.Icon, record.Icon)
		record.LastSeenAt = now
		record.UpdatedAt = now
		record.Details = mergeStringMaps(record.Details, fingerprint.Details)
		if record.State == domain.DiscoveryStateIgnored && record.IgnoreFingerprint != "" && record.IgnoreFingerprint != fingerprint.Hash {
			record.State = domain.DiscoveryStatePending
		}
	}
	record.LastFingerprintedAt = now

	settings, err := s.discoverySettingsTx(ctx, tx)
	if err != nil {
		return domain.DiscoveredServiceObservationOutcome{}, err
	}
	record.AutomationMode = settings.BookmarkPolicy
	if record.HealthConfigMode == "" {
		record.HealthConfigMode = domain.HealthConfigModeAuto
	}

	if err := s.saveDiscoveredServiceTx(ctx, tx, record); err != nil {
		return domain.DiscoveredServiceObservationOutcome{}, err
	}
	if err := s.upsertDiscoveredEvidenceTx(ctx, tx, record.ID, observation, fingerprint, now); err != nil {
		return domain.DiscoveredServiceObservationOutcome{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.DiscoveredServiceObservationOutcome{}, err
	}
	discovered, err := s.GetDiscoveredService(ctx, record.ID)
	if err != nil {
		return domain.DiscoveredServiceObservationOutcome{}, err
	}
	return domain.DiscoveredServiceObservationOutcome{DiscoveredService: discovered, Created: !found}, nil
}

func (s *Store) ListDiscoveredServices(ctx context.Context) ([]domain.DiscoveredService, error) {
	rows, err := s.reader().QueryContext(ctx, `
		SELECT
			ds.id,
			COALESCE(ds.device_id, ''),
			COALESCE(d.display_name, d.hostname, ''),
			ds.merge_key,
			ds.name,
			ds.service_type,
			ds.confidence_score,
			COALESCE(ds.service_definition_id, ''),
			ds.address_source,
			ds.host_value,
			ds.scheme,
			ds.port,
			ds.path,
			ds.url,
			ds.icon,
			ds.state,
			ds.ignore_fingerprint,
			ds.automation_mode,
			COALESCE(ds.health_config_mode, 'auto'),
			ds.status,
			COALESCE(ds.accepted_service_id, ''),
			COALESCE(ds.accepted_bookmark_id, ''),
			COALESCE(ds.last_checked_at, ''),
			COALESCE(ds.last_fingerprinted_at, ''),
			ds.details_json,
			ds.first_seen_at,
			ds.last_seen_at,
			ds.created_at,
			ds.updated_at
		FROM discovered_services ds
		LEFT JOIN devices d ON d.id = ds.device_id
		ORDER BY
			CASE ds.state
				WHEN 'pending' THEN 0
				WHEN 'accepted' THEN 1
				ELSE 2
			END,
			ds.confidence_score DESC,
			ds.last_seen_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]discoveredServiceRecord, 0)
	for rows.Next() {
		record, err := scanDiscoveredService(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return s.buildDiscoveredServices(ctx, records)
}

func (s *Store) ListRecentDiscoveredServices(ctx context.Context, limit int) ([]domain.DiscoveredService, error) {
	if limit <= 0 {
		limit = 6
	}
	items, err := s.ListDiscoveredServices(ctx)
	if err != nil {
		return nil, err
	}
	filtered := make([]domain.DiscoveredService, 0, len(items))
	for _, item := range items {
		if item.State == domain.DiscoveryStatePending {
			filtered = append(filtered, item)
		}
	}
	slices.SortFunc(filtered, func(left, right domain.DiscoveredService) int {
		if left.LastSeenAt.Equal(right.LastSeenAt) {
			return strings.Compare(left.ID, right.ID)
		}
		if left.LastSeenAt.After(right.LastSeenAt) {
			return -1
		}
		return 1
	})
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}
	return filtered, nil
}

func (s *Store) GetDiscoveredService(ctx context.Context, id string) (domain.DiscoveredService, error) {
	row := s.reader().QueryRowContext(ctx, `
		SELECT
			ds.id,
			COALESCE(ds.device_id, ''),
			COALESCE(d.display_name, d.hostname, ''),
			ds.merge_key,
			ds.name,
			ds.service_type,
			ds.confidence_score,
			COALESCE(ds.service_definition_id, ''),
			ds.address_source,
			ds.host_value,
			ds.scheme,
			ds.port,
			ds.path,
			ds.url,
			ds.icon,
			ds.state,
			ds.ignore_fingerprint,
			ds.automation_mode,
			COALESCE(ds.health_config_mode, 'auto'),
			ds.status,
			COALESCE(ds.accepted_service_id, ''),
			COALESCE(ds.accepted_bookmark_id, ''),
			COALESCE(ds.last_checked_at, ''),
			COALESCE(ds.last_fingerprinted_at, ''),
			ds.details_json,
			ds.first_seen_at,
			ds.last_seen_at,
			ds.created_at,
			ds.updated_at
		FROM discovered_services ds
		LEFT JOIN devices d ON d.id = ds.device_id
		WHERE ds.id = ?
	`, id)
	record, err := scanDiscoveredService(row)
	if err != nil {
		return domain.DiscoveredService{}, err
	}
	items, err := s.buildDiscoveredServices(ctx, []discoveredServiceRecord{record})
	if err != nil {
		return domain.DiscoveredService{}, err
	}
	if len(items) == 0 {
		return domain.DiscoveredService{}, sql.ErrNoRows
	}
	return items[0], nil
}

func (s *Store) IgnoreDiscoveredService(ctx context.Context, id string) (domain.DiscoveredService, error) {
	item, err := s.GetDiscoveredService(ctx, id)
	if err != nil {
		return domain.DiscoveredService{}, err
	}
	fingerprint := computeDiscoveredFingerprintHash(item)
	if _, err := s.db.ExecContext(ctx, `
		UPDATE discovered_services
		SET state = ?, ignore_fingerprint = ?, updated_at = ?
		WHERE id = ?
	`, domain.DiscoveryStateIgnored, fingerprint, nowString(), id); err != nil {
		return domain.DiscoveredService{}, err
	}
	return s.GetDiscoveredService(ctx, id)
}

func (s *Store) RestoreDiscoveredService(ctx context.Context, id string) (domain.DiscoveredService, error) {
	if _, err := s.db.ExecContext(ctx, `
		UPDATE discovered_services
		SET state = ?, ignore_fingerprint = '', updated_at = ?
		WHERE id = ?
	`, domain.DiscoveryStatePending, nowString(), id); err != nil {
		return domain.DiscoveredService{}, err
	}
	return s.GetDiscoveredService(ctx, id)
}

func (s *Store) MarkDiscoveredServiceAccepted(ctx context.Context, id, serviceID, bookmarkID string) (domain.DiscoveredService, error) {
	if _, err := s.db.ExecContext(ctx, `
		UPDATE discovered_services
		SET state = ?, accepted_service_id = ?, accepted_bookmark_id = ?, updated_at = ?
		WHERE id = ?
	`, domain.DiscoveryStateAccepted, nullableString(serviceID), nullableString(bookmarkID), nowString(), id); err != nil {
		return domain.DiscoveredService{}, err
	}
	return s.GetDiscoveredService(ctx, id)
}

func (s *Store) SaveDiscoveredServiceBookmarkLink(ctx context.Context, bookmarkID, discoveredServiceID string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE bookmarks SET discovered_service_id = ?, updated_at = ? WHERE id = ?`, nullableString(discoveredServiceID), nowString(), bookmarkID)
	return err
}

func (s *Store) UpdateDiscoveredServiceHealth(ctx context.Context, id string, status domain.HealthStatus, checkedAt time.Time) error {
	if checkedAt.IsZero() {
		checkedAt = time.Now().UTC()
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE discovered_services
		SET status = ?, last_checked_at = ?, updated_at = ?
		WHERE id = ?
	`, status, checkedAt.Format(time.RFC3339Nano), nowString(), id)
	return err
}

func (s *Store) ListDiscoveredServicesDueForHealth(ctx context.Context, interval time.Duration) ([]domain.DiscoveredService, error) {
	if interval <= 0 {
		interval = time.Minute
	}
	cutoff := time.Now().UTC().Add(-interval).Format(time.RFC3339Nano)
	rows, err := s.reader().QueryContext(ctx, `
		SELECT
			ds.id,
			COALESCE(ds.device_id, ''),
			COALESCE(d.display_name, d.hostname, ''),
			ds.merge_key,
			ds.name,
			ds.service_type,
			ds.confidence_score,
			COALESCE(ds.service_definition_id, ''),
			ds.address_source,
			ds.host_value,
			ds.scheme,
			ds.port,
			ds.path,
			ds.url,
			ds.icon,
			ds.state,
			ds.ignore_fingerprint,
			ds.automation_mode,
			COALESCE(ds.health_config_mode, 'auto'),
			ds.status,
			COALESCE(ds.accepted_service_id, ''),
			COALESCE(ds.accepted_bookmark_id, ''),
			COALESCE(ds.last_checked_at, ''),
			COALESCE(ds.last_fingerprinted_at, ''),
			ds.details_json,
			ds.first_seen_at,
			ds.last_seen_at,
			ds.created_at,
			ds.updated_at
		FROM discovered_services ds
		LEFT JOIN devices d ON d.id = ds.device_id
		WHERE ds.state = ?
		  AND (COALESCE(ds.last_checked_at, '') = '' OR ds.last_checked_at <= ?)
		ORDER BY ds.last_seen_at DESC
	`, domain.DiscoveryStatePending, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]discoveredServiceRecord, 0)
	for rows.Next() {
		record, err := scanDiscoveredService(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return s.buildDiscoveredServices(ctx, records)
}

func (s *Store) RefingerprintDiscoveredServices(ctx context.Context) error {
	items, err := s.ListDiscoveredServices(ctx)
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.State == domain.DiscoveryStateAccepted {
			continue
		}
		confidence := 40
		serviceType := item.ServiceType
		name := item.Name
		for _, evidence := range item.Evidence {
			fp := fingerprintObservation(domain.ServiceObservation{
				Name:            evidence.Name,
				Source:          evidence.Source,
				SourceRef:       evidence.SourceRef,
				ServiceTypeHint: evidence.ServiceTypeHint,
				AddressSource:   item.AddressSource,
				HostValue:       item.HostValue,
				Host:            evidence.Host,
				Port:            evidence.Port,
				Path:            evidence.Path,
				URL:             evidence.URL,
				Details:         evidence.Details,
			}, item.DeviceID)
			if fp.Confidence >= confidence {
				confidence = fp.Confidence
				serviceType = firstNonEmpty(fp.ServiceType, serviceType)
				name = firstNonEmpty(fp.Name, name)
			}
		}
		if _, err := s.db.ExecContext(ctx, `
			UPDATE discovered_services
			SET service_type = ?, name = ?, confidence_score = ?, last_fingerprinted_at = ?, updated_at = ?
			WHERE id = ?
		`, serviceType, name, confidence, nowString(), nowString(), item.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ApplyDiscoveredServiceDefinition(ctx context.Context, id string, match *domain.ServiceDefinitionMatch, confidence int, fingerprintedAt time.Time) error {
	if fingerprintedAt.IsZero() {
		fingerprintedAt = time.Now().UTC()
	}
	item, err := s.GetDiscoveredService(ctx, id)
	if err != nil {
		return err
	}
	serviceType := item.ServiceType
	name := item.Name
	icon := item.Icon
	definitionID := ""
	if match != nil {
		definitionID = match.Definition.ID
		serviceType = firstNonEmpty(match.Definition.Key, serviceType)
		name = firstNonEmpty(match.Definition.Name, name)
		icon = firstNonEmpty(match.Definition.Icon, icon)
		if confidence < match.Score {
			confidence = match.Score
		}
	}
	if confidence <= 0 {
		confidence = item.ConfidenceScore
	}
	if _, err := s.db.ExecContext(ctx, `
		UPDATE discovered_services
		SET service_definition_id = ?, service_type = ?, name = ?, icon = ?, confidence_score = ?, last_fingerprinted_at = ?, updated_at = ?
		WHERE id = ?
	`, nullableString(definitionID), serviceType, name, nullableString(icon), confidence, fingerprintedAt.Format(time.RFC3339Nano), nowString(), id); err != nil {
		return err
	}
	return s.SyncDiscoveredServiceHealthChecks(ctx, id)
}

func (s *Store) SyncDiscoveredServiceHealthChecks(ctx context.Context, id string) error {
	item, err := s.GetDiscoveredService(ctx, id)
	if err != nil {
		return err
	}
	if item.HealthConfigMode == domain.HealthConfigModeCustom {
		return nil
	}
	if item.ServiceDefinitionID != "" {
		definitions, err := s.ListServiceDefinitions(ctx)
		if err != nil {
			return err
		}
		for _, definition := range definitions {
			if definition.ID != item.ServiceDefinitionID {
				continue
			}
			checks := servicedefs.InstantiateChecks(domain.HealthCheckSubjectDiscoveredService, item.ID, item.AddressSource, item.HostValue, item.Host, item.Scheme, item.Port, item.Path, definition)
			return s.ReplaceManagedChecks(ctx, domain.HealthCheckSubjectDiscoveredService, item.ID, checks, definition.ID, domain.HealthConfigModeAuto)
		}
	}
	check := domain.ServiceCheck{
		SubjectType:     domain.HealthCheckSubjectDiscoveredService,
		SubjectID:       item.ID,
		ServiceID:       item.ID,
		AddressSource:   item.AddressSource,
		HostValue:       item.HostValue,
		Host:            item.Host,
		Protocol:        item.Scheme,
		Port:            item.Port,
		Enabled:         true,
		IntervalSeconds: 60,
		TimeoutSeconds:  10,
		SortOrder:       0,
		ConfigSource:    domain.HealthCheckConfigSourceFallback,
	}
	switch {
	case item.Host != "" && item.Port > 0:
		check.Name = "TCP connectivity"
		check.Type = domain.CheckTypeTCP
	default:
		check.Name = "Ping reachability"
		check.Type = domain.CheckTypePing
	}
	return s.ReplaceManagedChecks(ctx, domain.HealthCheckSubjectDiscoveredService, item.ID, []domain.ServiceCheck{check}, "", domain.HealthConfigModeAuto)
}

func (s *Store) buildDiscoveredServices(ctx context.Context, records []discoveredServiceRecord) ([]domain.DiscoveredService, error) {
	if len(records) == 0 {
		return []domain.DiscoveredService{}, nil
	}
	primaryAddresses, err := s.loadPrimaryDeviceAddresses(ctx)
	if err != nil {
		return nil, err
	}
	evidenceByService, err := s.loadDiscoveryEvidence(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]domain.DiscoveredService, 0, len(records))
	for _, record := range records {
		item := domain.DiscoveredService{
			ID:                  record.ID,
			DeviceID:            record.DeviceID,
			DeviceName:          record.DeviceName,
			MergeKey:            record.MergeKey,
			Name:                record.Name,
			ServiceType:         record.ServiceType,
			ConfidenceScore:     record.ConfidenceScore,
			ServiceDefinitionID: record.ServiceDefinitionID,
			AddressSource:       record.AddressSource,
			HostValue:           record.HostValue,
			Scheme:              record.Scheme,
			Port:                record.Port,
			Path:                record.Path,
			Icon:                record.Icon,
			State:               record.State,
			IgnoreFingerprint:   record.IgnoreFingerprint,
			AutomationMode:      record.AutomationMode,
			HealthConfigMode:    record.HealthConfigMode,
			Status:              record.Status,
			AcceptedServiceID:   record.AcceptedServiceID,
			AcceptedBookmarkID:  record.AcceptedBookmarkID,
			FirstSeenAt:         record.FirstSeenAt,
			LastSeenAt:          record.LastSeenAt,
			LastCheckedAt:       record.LastCheckedAt,
			LastFingerprintedAt: record.LastFingerprintedAt,
			CreatedAt:           record.CreatedAt,
			UpdatedAt:           record.UpdatedAt,
			Details:             record.Details,
			Evidence:            evidenceByService[record.ID],
		}
		item.Host = resolveAddressSourceHost(record.AddressSource, record.HostValue, primaryAddresses[record.DeviceID])
		item.URL = buildServiceURL(record.Scheme, item.Host, record.Port, record.Path)
		if item.URL == "" {
			item.URL = record.URL
		}
		for _, evidence := range item.Evidence {
			if !slices.Contains(item.SourceTypes, evidence.Source) {
				item.SourceTypes = append(item.SourceTypes, evidence.Source)
			}
		}
		slices.Sort(item.SourceTypes)
		items = append(items, item)
	}
	return items, nil
}

func (s *Store) discoverySettingsTx(ctx context.Context, tx *sql.Tx) (domain.DiscoverySettings, error) {
	rows, err := tx.QueryContext(ctx, `SELECT key, value FROM app_settings WHERE key IN ('bookmark_policy', 'auto_bookmark_sources', 'auto_bookmark_min_confidence')`)
	if err != nil {
		return domain.DiscoverySettings{}, err
	}
	defer rows.Close()

	settings := domain.DiscoverySettings{
		BookmarkPolicy:            domain.BookmarkAutomationManual,
		AutoBookmarkSources:       []domain.ServiceSource{domain.ServiceSourceDocker, domain.ServiceSourceLAN, domain.ServiceSourceMDNS},
		AutoBookmarkMinConfidence: 90,
	}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return domain.DiscoverySettings{}, err
		}
		switch key {
		case "bookmark_policy":
			settings.BookmarkPolicy = domain.BookmarkAutomationPolicy(value)
		case "auto_bookmark_sources":
			_ = json.Unmarshal([]byte(value), &settings.AutoBookmarkSources)
		case "auto_bookmark_min_confidence":
			if parsed, err := strconv.Atoi(value); err == nil {
				settings.AutoBookmarkMinConfidence = parsed
			}
		}
	}
	return settings, rows.Err()
}

func (s *Store) lookupDiscoveredServiceByMergeKeyTx(ctx context.Context, tx *sql.Tx, mergeKey string) (discoveredServiceRecord, bool, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT
			ds.id,
			COALESCE(ds.device_id, ''),
			COALESCE(d.display_name, d.hostname, ''),
			ds.merge_key,
			ds.name,
			ds.service_type,
			ds.confidence_score,
			COALESCE(ds.service_definition_id, ''),
			ds.address_source,
			ds.host_value,
			ds.scheme,
			ds.port,
			ds.path,
			ds.url,
			ds.icon,
			ds.state,
			ds.ignore_fingerprint,
			ds.automation_mode,
			COALESCE(ds.health_config_mode, 'auto'),
			ds.status,
			COALESCE(ds.accepted_service_id, ''),
			COALESCE(ds.accepted_bookmark_id, ''),
			COALESCE(ds.last_checked_at, ''),
			COALESCE(ds.last_fingerprinted_at, ''),
			ds.details_json,
			ds.first_seen_at,
			ds.last_seen_at,
			ds.created_at,
			ds.updated_at
		FROM discovered_services ds
		LEFT JOIN devices d ON d.id = ds.device_id
		WHERE ds.merge_key = ?
	`, mergeKey)
	record, err := scanDiscoveredService(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return discoveredServiceRecord{}, false, nil
		}
		return discoveredServiceRecord{}, false, err
	}
	return record, true, nil
}

func (s *Store) saveDiscoveredServiceTx(ctx context.Context, tx *sql.Tx, record discoveredServiceRecord) error {
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now().UTC()
	}
	if record.UpdatedAt.IsZero() {
		record.UpdatedAt = record.CreatedAt
	}
	if record.FirstSeenAt.IsZero() {
		record.FirstSeenAt = record.CreatedAt
	}
	if record.LastSeenAt.IsZero() {
		record.LastSeenAt = record.UpdatedAt
	}
	if record.AddressSource == "" {
		record.AddressSource = domain.ServiceAddressLiteralHost
	}
	if record.AutomationMode == "" {
		record.AutomationMode = domain.BookmarkAutomationManual
	}
	if record.HealthConfigMode == "" {
		record.HealthConfigMode = domain.HealthConfigModeAuto
	}
	if record.State == "" {
		record.State = domain.DiscoveryStatePending
	}
	if record.Status == "" {
		record.Status = domain.HealthStatusUnknown
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO discovered_services(
			id, device_id, merge_key, name, service_type, confidence_score, service_definition_id, address_source, host_value,
			scheme, port, path, url, icon, state, ignore_fingerprint, automation_mode, health_config_mode, status,
			last_checked_at, last_fingerprinted_at, accepted_service_id, accepted_bookmark_id, details_json,
			first_seen_at, last_seen_at, created_at, updated_at
		) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			device_id = excluded.device_id,
			merge_key = excluded.merge_key,
			name = excluded.name,
			service_type = excluded.service_type,
			confidence_score = excluded.confidence_score,
			service_definition_id = excluded.service_definition_id,
			address_source = excluded.address_source,
			host_value = excluded.host_value,
			scheme = excluded.scheme,
			port = excluded.port,
			path = excluded.path,
			url = excluded.url,
			icon = excluded.icon,
			state = excluded.state,
			ignore_fingerprint = excluded.ignore_fingerprint,
			automation_mode = excluded.automation_mode,
			health_config_mode = excluded.health_config_mode,
			status = excluded.status,
			last_checked_at = excluded.last_checked_at,
			last_fingerprinted_at = excluded.last_fingerprinted_at,
			accepted_service_id = COALESCE(excluded.accepted_service_id, discovered_services.accepted_service_id),
			accepted_bookmark_id = COALESCE(excluded.accepted_bookmark_id, discovered_services.accepted_bookmark_id),
			details_json = excluded.details_json,
			first_seen_at = MIN(discovered_services.first_seen_at, excluded.first_seen_at),
			last_seen_at = excluded.last_seen_at,
			updated_at = excluded.updated_at
	`, record.ID, nullableString(record.DeviceID), record.MergeKey, record.Name, record.ServiceType, record.ConfidenceScore, nullableString(record.ServiceDefinitionID), record.AddressSource, record.HostValue, record.Scheme, record.Port, record.Path, record.URL, nullableString(record.Icon), record.State, record.IgnoreFingerprint, record.AutomationMode, record.HealthConfigMode, record.Status, nullableTime(record.LastCheckedAt), nullableTime(record.LastFingerprintedAt), nullableString(record.AcceptedServiceID), nullableString(record.AcceptedBookmarkID), string(mustJSON(record.Details)), record.FirstSeenAt.Format(time.RFC3339Nano), record.LastSeenAt.Format(time.RFC3339Nano), record.CreatedAt.Format(time.RFC3339Nano), record.UpdatedAt.Format(time.RFC3339Nano))
	return err
}

func (s *Store) upsertDiscoveredEvidenceTx(ctx context.Context, tx *sql.Tx, discoveredServiceID string, observation domain.ServiceObservation, fingerprint discoveryFingerprint, seenAt time.Time) error {
	path := normalizePath(observation.Path)
	if path == "" {
		path = normalizePath(extractURLPath(observation.URL))
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO discovered_service_evidence(
			id, discovered_service_id, source_type, source_ref, service_type_hint, name, host, port, path,
			url, fingerprint_hash, details_json, first_seen_at, last_seen_at
		) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(source_type, source_ref) DO UPDATE SET
			discovered_service_id = excluded.discovered_service_id,
			service_type_hint = excluded.service_type_hint,
			name = excluded.name,
			host = excluded.host,
			port = excluded.port,
			path = excluded.path,
			url = excluded.url,
			fingerprint_hash = excluded.fingerprint_hash,
			details_json = excluded.details_json,
			last_seen_at = excluded.last_seen_at
	`, newID("devi"), discoveredServiceID, observation.Source, observation.SourceRef, firstNonEmpty(observation.ServiceTypeHint, fingerprint.ServiceType), firstNonEmpty(observation.Name, fingerprint.Name), firstNonEmpty(observation.Host, fingerprint.HostValue), observation.Port, path, observation.URL, fingerprint.Hash, string(mustJSON(sanitizeDiscoveryDetails(observation.Details))), seenAt.Format(time.RFC3339Nano), seenAt.Format(time.RFC3339Nano))
	return err
}

func (s *Store) loadDiscoveryEvidence(ctx context.Context) (map[string][]domain.DiscoveryEvidence, error) {
	rows, err := s.reader().QueryContext(ctx, `
		SELECT
			id,
			discovered_service_id,
			source_type,
			source_ref,
			service_type_hint,
			name,
			host,
			port,
			path,
			url,
			fingerprint_hash,
			details_json,
			first_seen_at,
			last_seen_at
		FROM discovered_service_evidence
		ORDER BY last_seen_at DESC, source_type
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := map[string][]domain.DiscoveryEvidence{}
	for rows.Next() {
		var item domain.DiscoveryEvidence
		var sourceType, detailsJSON, firstSeenAt, lastSeenAt string
		if err := rows.Scan(&item.ID, &item.DiscoveredServiceID, &sourceType, &item.SourceRef, &item.ServiceTypeHint, &item.Name, &item.Host, &item.Port, &item.Path, &item.URL, &item.FingerprintHash, &detailsJSON, &firstSeenAt, &lastSeenAt); err != nil {
			return nil, err
		}
		item.Source = domain.ServiceSource(sourceType)
		item.FirstSeenAt = parseTime(firstSeenAt)
		item.LastSeenAt = parseTime(lastSeenAt)
		_ = json.Unmarshal([]byte(detailsJSON), &item.Details)
		items[item.DiscoveredServiceID] = append(items[item.DiscoveredServiceID], item)
	}
	return items, rows.Err()
}

func scanDiscoveredService(scanner interface{ Scan(dest ...any) error }) (discoveredServiceRecord, error) {
	var record discoveredServiceRecord
	var detailsJSON string
	var lastCheckedAt, lastFingerprintedAt, firstSeenAt, lastSeenAt, createdAt, updatedAt string
	if err := scanner.Scan(
		&record.ID,
		&record.DeviceID,
		&record.DeviceName,
		&record.MergeKey,
		&record.Name,
		&record.ServiceType,
		&record.ConfidenceScore,
		&record.ServiceDefinitionID,
		&record.AddressSource,
		&record.HostValue,
		&record.Scheme,
		&record.Port,
		&record.Path,
		&record.URL,
		&record.Icon,
		&record.State,
		&record.IgnoreFingerprint,
		&record.AutomationMode,
		&record.HealthConfigMode,
		&record.Status,
		&record.AcceptedServiceID,
		&record.AcceptedBookmarkID,
		&lastCheckedAt,
		&lastFingerprintedAt,
		&detailsJSON,
		&firstSeenAt,
		&lastSeenAt,
		&createdAt,
		&updatedAt,
	); err != nil {
		return discoveredServiceRecord{}, err
	}
	record.LastCheckedAt = parseTime(lastCheckedAt)
	record.LastFingerprintedAt = parseTime(lastFingerprintedAt)
	record.FirstSeenAt = parseTime(firstSeenAt)
	record.LastSeenAt = parseTime(lastSeenAt)
	record.CreatedAt = parseTime(createdAt)
	record.UpdatedAt = parseTime(updatedAt)
	_ = json.Unmarshal([]byte(detailsJSON), &record.Details)
	return record, nil
}

func fingerprintObservation(observation domain.ServiceObservation, deviceID string) discoveryFingerprint {
	details := sanitizeDiscoveryDetails(observation.Details)
	serviceType := firstNonEmpty(strings.TrimSpace(observation.ServiceTypeHint), serviceTypeFromObservation(observation))
	name := strings.TrimSpace(observation.Name)
	confidence := 35

	if image := strings.ToLower(strings.TrimSpace(stringMapValue(details, "image"))); image != "" {
		if matchType, ok := serviceTypeByImage(image); ok {
			serviceType = firstNonEmpty(serviceType, matchType)
			confidence = max(confidence, 95)
		}
	}
	if hinted, ok := details["mdnsService"].(string); ok {
		if matchType, ok := serviceTypeByMDNS(strings.ToLower(strings.TrimSpace(hinted))); ok {
			serviceType = firstNonEmpty(serviceType, matchType)
			confidence = max(confidence, 88)
		}
	}
	if serviceType != "" {
		confidence = max(confidence, 75)
	}
	if name == "" {
		name = displayNameForServiceType(serviceType)
	}
	if name == "" {
		name = firstNonEmpty(serviceNameFromSource(observation), observation.Host, observation.SourceRef)
	}

	addressSource := observation.AddressSource
	hostValue := strings.TrimSpace(observation.HostValue)
	switch {
	case addressSource != "":
	case strings.HasSuffix(strings.ToLower(strings.TrimSpace(observation.Host)), ".local"):
		addressSource = domain.ServiceAddressMDNSHostname
	case deviceID != "" && isIPAddress(observation.Host):
		addressSource = domain.ServiceAddressDevicePrimary
	default:
		addressSource = domain.ServiceAddressLiteralHost
	}
	if hostValue == "" {
		switch addressSource {
		case domain.ServiceAddressMDNSHostname:
			hostValue = strings.ToLower(strings.TrimSpace(observation.Host))
		default:
			hostValue = strings.TrimSpace(observation.Host)
		}
	}
	if hostValue == "" {
		hostValue = hostValueFromURL(observation.URL)
	}
	if mappedType, ok := serviceTypeByPort(observation.Port); ok {
		if serviceType == "" || serviceType == "web" {
			serviceType = mappedType
		}
		confidence = max(confidence, 78)
	}
	if serviceType == "" && observation.Port > 0 {
		serviceType = genericServiceTypeForPort(observation.Port)
	}
	if displayName := displayNameForServiceType(serviceType); displayName != "" {
		name = displayName
	}
	if serviceType == "" {
		serviceType = "unknown"
		confidence = max(confidence, 40)
	}

	sum := sha256.Sum256([]byte(strings.Join([]string{
		string(observation.Source),
		strings.TrimSpace(observation.SourceRef),
		serviceType,
		hostValue,
		strconv.Itoa(observation.Port),
		normalizePath(observation.Path),
		stringMapValue(details, "image"),
		stringMapValue(details, "mdnsService"),
	}, "|")))
	return discoveryFingerprint{
		ServiceType:   serviceType,
		Name:          name,
		Confidence:    confidence,
		AddressSource: addressSource,
		HostValue:     hostValue,
		Icon:          serviceType,
		Details:       details,
		Hash:          hex.EncodeToString(sum[:]),
	}
}

func discoveredMergeKey(deviceID, hostValue string, port int, path, serviceType, name string) string {
	path = normalizePath(path)
	typeKey := slugify(firstNonEmpty(serviceType, name, "service"))
	if deviceID != "" {
		return fmt.Sprintf("device:%s|%d|%s|%s", deviceID, port, path, typeKey)
	}
	return fmt.Sprintf("host:%s|%d|%s|%s", strings.ToLower(strings.TrimSpace(hostValue)), port, path, typeKey)
}

func resolveAddressSourceHost(addressSource domain.ServiceAddressSource, hostValue, devicePrimaryAddress string) string {
	switch addressSource {
	case domain.ServiceAddressDevicePrimary:
		return firstNonEmpty(devicePrimaryAddress, hostValue)
	case domain.ServiceAddressMDNSHostname:
		return firstNonEmpty(hostValue, devicePrimaryAddress)
	default:
		return firstNonEmpty(hostValue, devicePrimaryAddress)
	}
}

func computeDiscoveredFingerprintHash(item domain.DiscoveredService) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		item.MergeKey,
		item.ServiceType,
		item.HostValue,
		strconv.Itoa(item.Port),
		item.Path,
	}, "|")))
	return hex.EncodeToString(sum[:])
}

func sanitizeDiscoveryDetails(details map[string]any) map[string]any {
	if len(details) == 0 {
		return map[string]any{}
	}
	sanitized := map[string]any{}
	for key, value := range details {
		if looksSensitiveKey(key) {
			continue
		}
		switch typed := value.(type) {
		case map[string]any:
			sanitized[key] = sanitizeDiscoveryDetails(typed)
		case map[string]string:
			nested := map[string]any{}
			for nestedKey, nestedValue := range typed {
				if looksSensitiveKey(nestedKey) {
					continue
				}
				nested[nestedKey] = nestedValue
			}
			sanitized[key] = nested
		default:
			sanitized[key] = value
		}
	}
	return sanitized
}

func looksSensitiveKey(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	return strings.Contains(key, "token") || strings.Contains(key, "secret") || strings.Contains(key, "password") || strings.Contains(key, "key")
}

func mergeStringMaps(base, update map[string]any) map[string]any {
	if len(base) == 0 && len(update) == 0 {
		return map[string]any{}
	}
	merged := map[string]any{}
	maps.Copy(merged, base)
	maps.Copy(merged, update)
	return merged
}

func serviceTypeByImage(image string) (string, bool) {
	switch {
	case strings.Contains(image, "grafana/grafana"):
		return "grafana", true
	case strings.Contains(image, "homeassistant/home-assistant"):
		return "home-assistant", true
	case strings.Contains(image, "prom/prometheus"):
		return "prometheus", true
	case strings.Contains(image, "plexinc/pms-docker"), strings.Contains(image, "linuxserver/plex"):
		return "plex", true
	case strings.Contains(image, "portainer/portainer"):
		return "portainer", true
	case strings.Contains(image, "nextcloud"):
		return "nextcloud", true
	default:
		return "", false
	}
}

func serviceTypeByPort(port int) (string, bool) {
	switch port {
	case 3000:
		return "grafana", true
	case 8123:
		return "home-assistant", true
	case 9090:
		return "prometheus", true
	case 32400:
		return "plex", true
	case 9000, 9443:
		return "portainer", true
	default:
		return "", false
	}
}

func genericServiceTypeForPort(port int) string {
	switch port {
	case 80, 443, 8080, 8443:
		return "web"
	default:
		return ""
	}
}

func serviceTypeByMDNS(service string) (string, bool) {
	switch service {
	case "_home-assistant._tcp":
		return "home-assistant", true
	case "_plexmediasvr._tcp":
		return "plex", true
	case "_http._tcp":
		return "web", true
	case "_https._tcp":
		return "web", true
	default:
		return "", false
	}
}

func serviceTypeFromObservation(observation domain.ServiceObservation) string {
	if observation.ServiceTypeHint != "" {
		return strings.TrimSpace(observation.ServiceTypeHint)
	}
	if typed, ok := serviceTypeByPort(observation.Port); ok {
		return typed
	}
	if hinted, ok := observation.Details["hint"].(string); ok && hinted == "web" {
		return genericServiceTypeForPort(observation.Port)
	}
	return ""
}

func serviceNameFromSource(observation domain.ServiceObservation) string {
	if displayName := displayNameForServiceType(serviceTypeFromObservation(observation)); displayName != "" {
		return displayName
	}
	if strings.TrimSpace(observation.Name) != "" {
		return observation.Name
	}
	return ""
}

func displayNameForServiceType(serviceType string) string {
	switch serviceType {
	case "grafana":
		return "Grafana"
	case "home-assistant":
		return "Home Assistant"
	case "prometheus":
		return "Prometheus"
	case "plex":
		return "Plex"
	case "portainer":
		return "Portainer"
	case "nextcloud":
		return "Nextcloud"
	case "web":
		return "Web Service"
	default:
		return ""
	}
}

func schemeFromObservation(observation domain.ServiceObservation, fingerprint discoveryFingerprint) string {
	if observation.Scheme != "" {
		return observation.Scheme
	}
	if strings.HasPrefix(observation.URL, "https://") {
		return "https"
	}
	if strings.HasPrefix(observation.URL, "http://") {
		return "http"
	}
	switch {
	case observation.Port == 443 || observation.Port == 8443 || observation.Port == 9443:
		return "https"
	case fingerprint.ServiceType == "web":
		return "http"
	default:
		return "http"
	}
}

func firstNonEmptyAddressSource(values ...domain.ServiceAddressSource) domain.ServiceAddressSource {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func mergeDiscoveryAddressSource(current, next domain.ServiceAddressSource) domain.ServiceAddressSource {
	switch {
	case current == domain.ServiceAddressDevicePrimary || next == domain.ServiceAddressDevicePrimary:
		return domain.ServiceAddressDevicePrimary
	case current == domain.ServiceAddressMDNSHostname || next == domain.ServiceAddressMDNSHostname:
		return domain.ServiceAddressMDNSHostname
	default:
		return firstNonEmptyAddressSource(next, current, domain.ServiceAddressLiteralHost)
	}
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || path == "/" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func hostValueFromURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}

func extractURLPath(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsed.Path
}

func stringMapValue(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	if raw, ok := values[key]; ok {
		if typed, ok := raw.(string); ok {
			return typed
		}
	}
	return ""
}
