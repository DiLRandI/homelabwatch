package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
)

func (s *Store) ListNotificationChannels(ctx context.Context) ([]domain.NotificationChannel, error) {
	rows, err := s.reader().QueryContext(ctx, `SELECT id, name, type, enabled, config_json, created_at, updated_at FROM notification_channels ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.NotificationChannel
	for rows.Next() {
		item, err := scanNotificationChannel(rows.Scan)
		if err != nil {
			return nil, err
		}
		item.Config = redactNotificationConfig(item.Type, item.Config)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) GetNotificationChannel(ctx context.Context, id string) (domain.NotificationChannel, error) {
	item, err := s.getNotificationChannel(ctx, s.reader(), id)
	if err != nil {
		return domain.NotificationChannel{}, err
	}
	item.Config = redactNotificationConfig(item.Type, item.Config)
	return item, nil
}

func (s *Store) GetNotificationChannelForSend(ctx context.Context, id string) (domain.NotificationChannel, error) {
	return s.getNotificationChannel(ctx, s.reader(), id)
}

func (s *Store) SaveNotificationChannel(ctx context.Context, input domain.NotificationChannel) (domain.NotificationChannel, error) {
	existing := domain.NotificationChannel{}
	if input.ID != "" {
		found, err := s.getNotificationChannel(ctx, s.reader(), input.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return domain.NotificationChannel{}, err
		}
		existing = found
	}
	if existing.ID != "" {
		input = mergeNotificationChannelPatch(existing, input)
	}
	if input.ID == "" {
		input.ID = newID("nch")
		input.Enabled = true
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return domain.NotificationChannel{}, errors.New("channel name is required")
	}
	if input.Type == "" {
		input.Type = existing.Type
	}
	if input.Type != domain.NotificationChannelWebhook && input.Type != domain.NotificationChannelNtfy {
		return domain.NotificationChannel{}, errors.New("channel type must be webhook or ntfy")
	}
	config, err := normalizeNotificationChannelConfig(input.Type, input.Config, existing.Config)
	if err != nil {
		return domain.NotificationChannel{}, err
	}
	now := time.Now().UTC()
	if input.CreatedAt.IsZero() {
		input.CreatedAt = now
	}
	input.UpdatedAt = now
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO notification_channels(id, name, type, enabled, config_json, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name, type = excluded.type, enabled = excluded.enabled, config_json = excluded.config_json, updated_at = excluded.updated_at
	`, input.ID, input.Name, input.Type, boolInt(input.Enabled), string(mustJSON(config)), input.CreatedAt.Format(time.RFC3339Nano), input.UpdatedAt.Format(time.RFC3339Nano))
	if err != nil {
		return domain.NotificationChannel{}, err
	}
	return s.GetNotificationChannel(ctx, input.ID)
}

func (s *Store) DeleteNotificationChannel(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM notification_channels WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) ListNotificationRules(ctx context.Context) ([]domain.NotificationRule, error) {
	rows, err := s.reader().QueryContext(ctx, `SELECT id, name, event_type, enabled, filters_json, created_at, updated_at FROM notification_rules ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.NotificationRule
	for rows.Next() {
		item, err := scanNotificationRule(rows.Scan)
		if err != nil {
			return nil, err
		}
		if err := s.attachNotificationRuleChannels(ctx, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) ListEnabledNotificationRules(ctx context.Context, eventType domain.NotificationEventType) ([]domain.NotificationRule, error) {
	rows, err := s.reader().QueryContext(ctx, `SELECT id, name, event_type, enabled, filters_json, created_at, updated_at FROM notification_rules WHERE enabled = 1 AND event_type = ? ORDER BY name`, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.NotificationRule
	for rows.Next() {
		item, err := scanNotificationRule(rows.Scan)
		if err != nil {
			return nil, err
		}
		if err := s.attachNotificationRuleChannels(ctx, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) SaveNotificationRule(ctx context.Context, input domain.NotificationRule) (domain.NotificationRule, error) {
	existing := domain.NotificationRule{}
	if input.ID != "" {
		found, err := s.GetNotificationRule(ctx, input.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return domain.NotificationRule{}, err
		}
		existing = found
	}
	if input.ID == "" {
		input.ID = newID("nrl")
		input.Enabled = true
	}
	if existing.ID != "" && input.CreatedAt.IsZero() {
		input.CreatedAt = existing.CreatedAt
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return domain.NotificationRule{}, errors.New("rule name is required")
	}
	if !validNotificationEventType(input.EventType) {
		return domain.NotificationRule{}, errors.New("unsupported notification event type")
	}
	if len(input.ChannelIDs) == 0 {
		return domain.NotificationRule{}, errors.New("rule requires at least one channel")
	}
	if input.Filters == nil {
		input.Filters = map[string]any{}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.NotificationRule{}, err
	}
	defer tx.Rollback()
	for _, channelID := range input.ChannelIDs {
		var exists int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(1) FROM notification_channels WHERE id = ?`, channelID).Scan(&exists); err != nil {
			return domain.NotificationRule{}, err
		}
		if exists == 0 {
			return domain.NotificationRule{}, fmt.Errorf("notification channel %s not found", channelID)
		}
	}
	now := time.Now().UTC()
	if input.CreatedAt.IsZero() {
		input.CreatedAt = now
	}
	input.UpdatedAt = now
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO notification_rules(id, name, event_type, enabled, filters_json, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name, event_type = excluded.event_type, enabled = excluded.enabled, filters_json = excluded.filters_json, updated_at = excluded.updated_at
	`, input.ID, input.Name, input.EventType, boolInt(input.Enabled), string(mustJSON(input.Filters)), input.CreatedAt.Format(time.RFC3339Nano), input.UpdatedAt.Format(time.RFC3339Nano)); err != nil {
		return domain.NotificationRule{}, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM notification_rule_channels WHERE rule_id = ?`, input.ID); err != nil {
		return domain.NotificationRule{}, err
	}
	for _, channelID := range input.ChannelIDs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO notification_rule_channels(rule_id, channel_id, created_at) VALUES(?, ?, ?)`, input.ID, channelID, now.Format(time.RFC3339Nano)); err != nil {
			return domain.NotificationRule{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return domain.NotificationRule{}, err
	}
	return s.GetNotificationRule(ctx, input.ID)
}

func (s *Store) GetNotificationRule(ctx context.Context, id string) (domain.NotificationRule, error) {
	row := s.reader().QueryRowContext(ctx, `SELECT id, name, event_type, enabled, filters_json, created_at, updated_at FROM notification_rules WHERE id = ?`, id)
	item, err := scanNotificationRule(row.Scan)
	if err != nil {
		return domain.NotificationRule{}, err
	}
	if err := s.attachNotificationRuleChannels(ctx, &item); err != nil {
		return domain.NotificationRule{}, err
	}
	return item, nil
}

func (s *Store) DeleteNotificationRule(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM notification_rules WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) CreateNotificationDelivery(ctx context.Context, delivery domain.NotificationDelivery) (domain.NotificationDelivery, error) {
	if delivery.ID == "" {
		delivery.ID = newID("ndl")
	}
	if delivery.Status == "" {
		delivery.Status = domain.NotificationDeliveryPending
	}
	if delivery.AttemptedAt.IsZero() {
		delivery.AttemptedAt = time.Now().UTC()
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO notification_deliveries(id, rule_id, channel_id, event_type, status, message, attempted_at) VALUES(?, ?, ?, ?, ?, ?, ?)`, delivery.ID, nullableString(delivery.RuleID), nullableString(delivery.ChannelID), delivery.EventType, delivery.Status, delivery.Message, delivery.AttemptedAt.Format(time.RFC3339Nano))
	if err != nil {
		return domain.NotificationDelivery{}, err
	}
	return s.GetNotificationDelivery(ctx, delivery.ID)
}

func (s *Store) UpdateNotificationDelivery(ctx context.Context, delivery domain.NotificationDelivery) (domain.NotificationDelivery, error) {
	if delivery.AttemptedAt.IsZero() {
		delivery.AttemptedAt = time.Now().UTC()
	}
	_, err := s.db.ExecContext(ctx, `UPDATE notification_deliveries SET status = ?, message = ?, attempted_at = ? WHERE id = ?`, delivery.Status, delivery.Message, delivery.AttemptedAt.Format(time.RFC3339Nano), delivery.ID)
	if err != nil {
		return domain.NotificationDelivery{}, err
	}
	return s.GetNotificationDelivery(ctx, delivery.ID)
}

func (s *Store) GetNotificationDelivery(ctx context.Context, id string) (domain.NotificationDelivery, error) {
	rows, err := s.reader().QueryContext(ctx, notificationDeliverySelect()+` WHERE d.id = ?`, id)
	if err != nil {
		return domain.NotificationDelivery{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return domain.NotificationDelivery{}, sql.ErrNoRows
	}
	return scanNotificationDelivery(rows.Scan)
}

func (s *Store) ListNotificationDeliveries(ctx context.Context, limit int) ([]domain.NotificationDelivery, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	rows, err := s.reader().QueryContext(ctx, notificationDeliverySelect()+` ORDER BY d.attempted_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.NotificationDelivery
	for rows.Next() {
		item, err := scanNotificationDelivery(rows.Scan)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) getNotificationChannel(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, id string) (domain.NotificationChannel, error) {
	row := queryer.QueryRowContext(ctx, `SELECT id, name, type, enabled, config_json, created_at, updated_at FROM notification_channels WHERE id = ?`, id)
	return scanNotificationChannel(row.Scan)
}

func (s *Store) attachNotificationRuleChannels(ctx context.Context, rule *domain.NotificationRule) error {
	rows, err := s.reader().QueryContext(ctx, `
		SELECT c.id, c.name, c.type, c.enabled, c.config_json, c.created_at, c.updated_at
		FROM notification_rule_channels rc
		JOIN notification_channels c ON c.id = rc.channel_id
		WHERE rc.rule_id = ?
		ORDER BY c.name
	`, rule.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		channel, err := scanNotificationChannel(rows.Scan)
		if err != nil {
			return err
		}
		channel.Config = redactNotificationConfig(channel.Type, channel.Config)
		rule.ChannelIDs = append(rule.ChannelIDs, channel.ID)
		rule.Channels = append(rule.Channels, channel)
	}
	return rows.Err()
}

func normalizeNotificationChannelConfig(channelType domain.NotificationChannelType, config, existing map[string]any) (map[string]any, error) {
	next := map[string]any{}
	for key, value := range config {
		if str, ok := value.(string); ok && (strings.TrimSpace(str) == "" || str == domain.RedactedSecret) {
			if existing != nil {
				if prior, found := existing[key]; found {
					next[key] = prior
				}
			}
			continue
		}
		next[key] = value
	}
	switch channelType {
	case domain.NotificationChannelWebhook:
		rawURL := stringConfig(next, "url")
		if rawURL == "" {
			return nil, errors.New("webhook url is required")
		}
		if _, err := url.ParseRequestURI(rawURL); err != nil {
			return nil, errors.New("webhook url must be valid")
		}
		if _, ok := next["timeoutSeconds"]; !ok {
			next["timeoutSeconds"] = float64(10)
		}
	case domain.NotificationChannelNtfy:
		serverURL := strings.TrimRight(stringConfig(next, "serverUrl"), "/")
		if serverURL == "" {
			return nil, errors.New("ntfy serverUrl is required")
		}
		if _, err := url.ParseRequestURI(serverURL); err != nil {
			return nil, errors.New("ntfy serverUrl must be valid")
		}
		topic := strings.Trim(strings.TrimSpace(stringConfig(next, "topic")), "/")
		if topic == "" {
			return nil, errors.New("ntfy topic is required")
		}
		next["serverUrl"] = serverURL
		next["topic"] = topic
		if stringConfig(next, "priority") == "" {
			next["priority"] = "default"
		}
	}
	return next, nil
}

func mergeNotificationChannelPatch(existing, input domain.NotificationChannel) domain.NotificationChannel {
	if strings.TrimSpace(input.Name) == "" {
		input.Name = existing.Name
	}
	if input.Type == "" {
		input.Type = existing.Type
	}
	if input.Config == nil {
		input.Config = existing.Config
	}
	input.CreatedAt = existing.CreatedAt
	return input
}

func redactNotificationConfig(channelType domain.NotificationChannelType, config map[string]any) map[string]any {
	next := map[string]any{}
	for key, value := range config {
		next[key] = value
	}
	switch channelType {
	case domain.NotificationChannelWebhook:
		if stringConfig(next, "url") != "" {
			next["url"] = domain.RedactedSecret
		}
	case domain.NotificationChannelNtfy:
		if stringConfig(next, "token") != "" {
			next["token"] = domain.RedactedSecret
		}
	}
	return next
}

func scanNotificationChannel(scan func(dest ...any) error) (domain.NotificationChannel, error) {
	var item domain.NotificationChannel
	var enabled int
	var configJSON, createdAt, updatedAt string
	if err := scan(&item.ID, &item.Name, &item.Type, &enabled, &configJSON, &createdAt, &updatedAt); err != nil {
		return domain.NotificationChannel{}, err
	}
	item.Enabled = enabled == 1
	item.CreatedAt = parseTime(createdAt)
	item.UpdatedAt = parseTime(updatedAt)
	item.Config = map[string]any{}
	_ = json.Unmarshal([]byte(configJSON), &item.Config)
	return item, nil
}

func scanNotificationRule(scan func(dest ...any) error) (domain.NotificationRule, error) {
	var item domain.NotificationRule
	var enabled int
	var filtersJSON, createdAt, updatedAt string
	if err := scan(&item.ID, &item.Name, &item.EventType, &enabled, &filtersJSON, &createdAt, &updatedAt); err != nil {
		return domain.NotificationRule{}, err
	}
	item.Enabled = enabled == 1
	item.CreatedAt = parseTime(createdAt)
	item.UpdatedAt = parseTime(updatedAt)
	item.Filters = map[string]any{}
	_ = json.Unmarshal([]byte(filtersJSON), &item.Filters)
	return item, nil
}

func notificationDeliverySelect() string {
	return `SELECT d.id, COALESCE(d.rule_id, ''), COALESCE(r.name, ''), COALESCE(d.channel_id, ''), COALESCE(c.name, ''), d.event_type, d.status, d.message, d.attempted_at
		FROM notification_deliveries d
		LEFT JOIN notification_rules r ON r.id = d.rule_id
		LEFT JOIN notification_channels c ON c.id = d.channel_id`
}

func scanNotificationDelivery(scan func(dest ...any) error) (domain.NotificationDelivery, error) {
	var item domain.NotificationDelivery
	var attemptedAt string
	if err := scan(&item.ID, &item.RuleID, &item.RuleName, &item.ChannelID, &item.ChannelName, &item.EventType, &item.Status, &item.Message, &attemptedAt); err != nil {
		return domain.NotificationDelivery{}, err
	}
	item.AttemptedAt = parseTime(attemptedAt)
	return item, nil
}

func validNotificationEventType(value domain.NotificationEventType) bool {
	switch value {
	case domain.NotificationEventServiceHealthChanged,
		domain.NotificationEventCheckFailed,
		domain.NotificationEventCheckRecovered,
		domain.NotificationEventDiscoveredServiceCreated,
		domain.NotificationEventDeviceCreated,
		domain.NotificationEventWorkerFailed:
		return true
	default:
		return false
	}
}

func stringConfig(config map[string]any, key string) string {
	if config == nil {
		return ""
	}
	value, ok := config[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}
