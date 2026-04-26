CREATE TABLE IF NOT EXISTS notification_channels (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('webhook', 'ntfy')),
    enabled INTEGER NOT NULL DEFAULT 1,
    config_json TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS notification_rules (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    event_type TEXT NOT NULL CHECK(event_type IN (
        'service_health_changed',
        'check_failed',
        'check_recovered',
        'discovered_service_created',
        'device_created',
        'worker_failed'
    )),
    enabled INTEGER NOT NULL DEFAULT 1,
    filters_json TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS notification_rule_channels (
    rule_id TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY(rule_id, channel_id),
    FOREIGN KEY(rule_id) REFERENCES notification_rules(id) ON DELETE CASCADE,
    FOREIGN KEY(channel_id) REFERENCES notification_channels(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS notification_deliveries (
    id TEXT PRIMARY KEY,
    rule_id TEXT,
    channel_id TEXT,
    event_type TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('pending', 'sent', 'failed')),
    message TEXT NOT NULL DEFAULT '',
    attempted_at TEXT NOT NULL
);

ALTER TABLE job_state ADD COLUMN consecutive_failures INTEGER NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_notification_rules_event_enabled ON notification_rules(event_type, enabled);
CREATE INDEX IF NOT EXISTS idx_notification_rule_channels_channel ON notification_rule_channels(channel_id);
CREATE INDEX IF NOT EXISTS idx_notification_deliveries_attempted ON notification_deliveries(attempted_at DESC);
