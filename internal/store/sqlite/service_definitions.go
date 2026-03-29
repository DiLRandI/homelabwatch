package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
	"github.com/deleema/homelabwatch/internal/servicedefs"
)

func (s *Store) seedBuiltInServiceDefinitions(context.Context) error {
	return nil
}

func (s *Store) ListServiceDefinitions(ctx context.Context) ([]domain.ServiceDefinition, error) {
	custom, err := s.listCustomServiceDefinitions(ctx)
	if err != nil {
		return nil, err
	}
	return servicedefs.MergeDefinitions(custom), nil
}

func (s *Store) SaveServiceDefinition(ctx context.Context, input domain.ServiceDefinitionInput) (domain.ServiceDefinition, error) {
	if strings.HasPrefix(strings.TrimSpace(input.ID), "builtin_") {
		return domain.ServiceDefinition{}, errors.New("built-in service definitions are read-only")
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return domain.ServiceDefinition{}, errors.New("service definition name is required")
	}
	if len(input.CheckTemplates) == 0 {
		return domain.ServiceDefinition{}, errors.New("service definition must include at least one check template")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.ServiceDefinition{}, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = newID("sdef")
	}
	key := strings.TrimSpace(input.Key)
	if key == "" {
		key = slugify(name)
	}
	enabled := input.Enabled
	if !input.Enabled {
		enabled = false
	}
	createdAt := now
	if existing, err := s.getCustomServiceDefinitionTx(ctx, tx, id); err == nil {
		createdAt = existing.CreatedAt
	} else if !errors.Is(err, sql.ErrNoRows) {
		return domain.ServiceDefinition{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO service_definitions(id, key, name, icon, priority, built_in, enabled, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?, 0, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			key = excluded.key,
			name = excluded.name,
			icon = excluded.icon,
			priority = excluded.priority,
			enabled = excluded.enabled,
			updated_at = excluded.updated_at
	`, id, key, name, strings.TrimSpace(input.Icon), input.Priority, boolInt(enabled), createdAt.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano)); err != nil {
		return domain.ServiceDefinition{}, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM service_definition_matchers WHERE service_definition_id = ?`, id); err != nil {
		return domain.ServiceDefinition{}, err
	}
	for index, matcher := range input.Matchers {
		matcherID := strings.TrimSpace(matcher.ID)
		if matcherID == "" {
			matcherID = newID("sdm")
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO service_definition_matchers(id, service_definition_id, type, operator, value, extra, weight, sort_order, created_at, updated_at)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, matcherID, id, strings.TrimSpace(matcher.Type), firstNonEmpty(strings.TrimSpace(matcher.Operator), "exact"), strings.TrimSpace(matcher.Value), strings.TrimSpace(matcher.Extra), matcher.Weight, index, now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano)); err != nil {
			return domain.ServiceDefinition{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM service_definition_check_templates WHERE service_definition_id = ?`, id); err != nil {
		return domain.ServiceDefinition{}, err
	}
	for index, template := range input.CheckTemplates {
		templateID := strings.TrimSpace(template.ID)
		if templateID == "" {
			templateID = newID("sdc")
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO service_definition_check_templates(
				id, service_definition_id, name, kind, protocol, address_source, host_value, port, path, method,
				interval_seconds, timeout_seconds, expected_status_min, expected_status_max, enabled, sort_order,
				config_source, created_at, updated_at
			) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, templateID, id, strings.TrimSpace(template.Name), template.Type, strings.TrimSpace(template.Protocol), strings.TrimSpace(string(template.AddressSource)), strings.TrimSpace(template.HostValue), template.Port, strings.TrimSpace(template.Path), firstNonEmpty(strings.TrimSpace(template.Method), "GET"), definitionDefaultInt(template.IntervalSeconds, 60), definitionDefaultInt(template.TimeoutSeconds, 10), definitionDefaultStatusMin(template.ExpectedStatusMin, template.Type), definitionDefaultStatusMax(template.ExpectedStatusMax, template.Type), boolInt(template.Enabled), index, firstNonEmpty(string(template.ConfigSource), string(domain.HealthCheckConfigSourceDefinition)), now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano)); err != nil {
			return domain.ServiceDefinition{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return domain.ServiceDefinition{}, err
	}
	definitions, err := s.ListServiceDefinitions(ctx)
	if err != nil {
		return domain.ServiceDefinition{}, err
	}
	for _, item := range definitions {
		if item.ID == id {
			return item, nil
		}
	}
	return domain.ServiceDefinition{}, sql.ErrNoRows
}

func (s *Store) DeleteServiceDefinition(ctx context.Context, id string) error {
	if strings.HasPrefix(strings.TrimSpace(id), "builtin_") {
		return errors.New("built-in service definitions are read-only")
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM service_definitions WHERE id = ?`, id)
	return err
}

func (s *Store) listCustomServiceDefinitions(ctx context.Context) ([]domain.ServiceDefinition, error) {
	rows, err := s.reader().QueryContext(ctx, `
		SELECT id, key, name, icon, priority, built_in, enabled, created_at, updated_at
		FROM service_definitions
		ORDER BY priority DESC, name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.ServiceDefinition{}
	for rows.Next() {
		var item domain.ServiceDefinition
		var builtIn, enabled int
		var createdAt, updatedAt string
		if err := rows.Scan(&item.ID, &item.Key, &item.Name, &item.Icon, &item.Priority, &builtIn, &enabled, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		item.BuiltIn = builtIn == 1
		item.Enabled = enabled == 1
		item.CreatedAt = parseTime(createdAt)
		item.UpdatedAt = parseTime(updatedAt)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	matchersByDefinition, err := s.loadServiceDefinitionMatchers(ctx)
	if err != nil {
		return nil, err
	}
	checksByDefinition, err := s.loadServiceDefinitionChecks(ctx)
	if err != nil {
		return nil, err
	}
	for index := range items {
		items[index].Matchers = matchersByDefinition[items[index].ID]
		items[index].CheckTemplates = checksByDefinition[items[index].ID]
	}
	return items, nil
}

func (s *Store) getCustomServiceDefinitionTx(ctx context.Context, tx *sql.Tx, id string) (domain.ServiceDefinition, error) {
	var item domain.ServiceDefinition
	var builtIn, enabled int
	var createdAt, updatedAt string
	err := tx.QueryRowContext(ctx, `
		SELECT id, key, name, icon, priority, built_in, enabled, created_at, updated_at
		FROM service_definitions
		WHERE id = ?
	`, id).Scan(&item.ID, &item.Key, &item.Name, &item.Icon, &item.Priority, &builtIn, &enabled, &createdAt, &updatedAt)
	if err != nil {
		return domain.ServiceDefinition{}, err
	}
	item.BuiltIn = builtIn == 1
	item.Enabled = enabled == 1
	item.CreatedAt = parseTime(createdAt)
	item.UpdatedAt = parseTime(updatedAt)
	return item, nil
}

func (s *Store) loadServiceDefinitionMatchers(ctx context.Context) (map[string][]domain.ServiceDefinitionMatcher, error) {
	rows, err := s.reader().QueryContext(ctx, `
		SELECT id, service_definition_id, type, operator, value, extra, weight, sort_order, created_at, updated_at
		FROM service_definition_matchers
		ORDER BY service_definition_id, sort_order
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := map[string][]domain.ServiceDefinitionMatcher{}
	for rows.Next() {
		var item domain.ServiceDefinitionMatcher
		var definitionID, createdAt, updatedAt string
		if err := rows.Scan(&item.ID, &definitionID, &item.Type, &item.Operator, &item.Value, &item.Extra, &item.Weight, &item.SortOrder, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		item.CreatedAt = parseTime(createdAt)
		item.UpdatedAt = parseTime(updatedAt)
		items[definitionID] = append(items[definitionID], item)
	}
	return items, rows.Err()
}

func (s *Store) loadServiceDefinitionChecks(ctx context.Context) (map[string][]domain.ServiceDefinitionCheckTemplate, error) {
	rows, err := s.reader().QueryContext(ctx, `
		SELECT id, service_definition_id, name, kind, protocol, address_source, host_value, port, path, method, interval_seconds, timeout_seconds, expected_status_min, expected_status_max, enabled, sort_order, config_source
		FROM service_definition_check_templates
		ORDER BY service_definition_id, sort_order
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := map[string][]domain.ServiceDefinitionCheckTemplate{}
	for rows.Next() {
		var item domain.ServiceDefinitionCheckTemplate
		var definitionID, addressSource, configSource string
		var enabled int
		if err := rows.Scan(&item.ID, &definitionID, &item.Name, &item.Type, &item.Protocol, &addressSource, &item.HostValue, &item.Port, &item.Path, &item.Method, &item.IntervalSeconds, &item.TimeoutSeconds, &item.ExpectedStatusMin, &item.ExpectedStatusMax, &enabled, &item.SortOrder, &configSource); err != nil {
			return nil, err
		}
		item.AddressSource = domain.ServiceAddressSource(addressSource)
		item.ConfigSource = domain.HealthCheckConfigSource(configSource)
		item.Enabled = enabled == 1
		items[definitionID] = append(items[definitionID], item)
	}
	for key := range items {
		sort.SliceStable(items[key], func(i, j int) bool {
			if items[key][i].SortOrder == items[key][j].SortOrder {
				return items[key][i].Name < items[key][j].Name
			}
			return items[key][i].SortOrder < items[key][j].SortOrder
		})
	}
	return items, rows.Err()
}

func definitionDefaultInt(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func definitionDefaultStatusMin(value int, checkType domain.CheckType) int {
	if value > 0 {
		return value
	}
	if checkType == domain.CheckTypeHTTP {
		return 200
	}
	return 0
}

func definitionDefaultStatusMax(value int, checkType domain.CheckType) int {
	if value > 0 {
		return value
	}
	if checkType == domain.CheckTypeHTTP {
		return 399
	}
	return 0
}
