ALTER TABLE services ADD COLUMN service_definition_id TEXT;
ALTER TABLE services ADD COLUMN health_config_mode TEXT NOT NULL DEFAULT 'auto';
ALTER TABLE services ADD COLUMN fingerprinted_at TEXT;

ALTER TABLE discovered_services ADD COLUMN service_definition_id TEXT;
ALTER TABLE discovered_services ADD COLUMN health_config_mode TEXT NOT NULL DEFAULT 'auto';

ALTER TABLE health_checks ADD COLUMN name TEXT NOT NULL DEFAULT '';
ALTER TABLE health_checks ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0;
ALTER TABLE health_checks ADD COLUMN protocol TEXT NOT NULL DEFAULT '';
ALTER TABLE health_checks ADD COLUMN address_source TEXT NOT NULL DEFAULT 'literal_host';
ALTER TABLE health_checks ADD COLUMN host_value TEXT NOT NULL DEFAULT '';
ALTER TABLE health_checks ADD COLUMN port INTEGER NOT NULL DEFAULT 0;
ALTER TABLE health_checks ADD COLUMN path TEXT NOT NULL DEFAULT '';
ALTER TABLE health_checks ADD COLUMN method TEXT NOT NULL DEFAULT 'GET';
ALTER TABLE health_checks ADD COLUMN config_source TEXT NOT NULL DEFAULT 'fallback';
ALTER TABLE health_checks ADD COLUMN service_definition_id TEXT;

ALTER TABLE health_check_results ADD COLUMN http_status_code INTEGER NOT NULL DEFAULT 0;
ALTER TABLE health_check_results ADD COLUMN response_size_bytes INTEGER NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS service_definitions (
    id TEXT PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    icon TEXT NOT NULL DEFAULT '',
    priority INTEGER NOT NULL DEFAULT 0,
    built_in INTEGER NOT NULL DEFAULT 0,
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS service_definition_matchers (
    id TEXT PRIMARY KEY,
    service_definition_id TEXT NOT NULL,
    type TEXT NOT NULL,
    operator TEXT NOT NULL DEFAULT 'exact',
    value TEXT NOT NULL DEFAULT '',
    extra TEXT NOT NULL DEFAULT '',
    weight INTEGER NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY(service_definition_id) REFERENCES service_definitions(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS service_definition_check_templates (
    id TEXT PRIMARY KEY,
    service_definition_id TEXT NOT NULL,
    name TEXT NOT NULL,
    kind TEXT NOT NULL,
    protocol TEXT NOT NULL DEFAULT '',
    address_source TEXT NOT NULL DEFAULT '',
    host_value TEXT NOT NULL DEFAULT '',
    port INTEGER NOT NULL DEFAULT 0,
    path TEXT NOT NULL DEFAULT '',
    method TEXT NOT NULL DEFAULT 'GET',
    interval_seconds INTEGER NOT NULL DEFAULT 60,
    timeout_seconds INTEGER NOT NULL DEFAULT 10,
    expected_status_min INTEGER NOT NULL DEFAULT 200,
    expected_status_max INTEGER NOT NULL DEFAULT 399,
    enabled INTEGER NOT NULL DEFAULT 1,
    sort_order INTEGER NOT NULL DEFAULT 0,
    config_source TEXT NOT NULL DEFAULT 'definition',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY(service_definition_id) REFERENCES service_definitions(id) ON DELETE CASCADE
);

WITH latest_results AS (
    SELECT
        check_id,
        MAX(checked_at) AS checked_at
    FROM check_results
    GROUP BY check_id
)
INSERT INTO health_checks(
    id,
    subject_type,
    subject_id,
    kind,
    target,
    interval_seconds,
    timeout_seconds,
    expected_status_min,
    expected_status_max,
    enabled,
    next_run_at,
    last_run_at,
    last_status,
    consecutive_failures,
    created_at,
    updated_at,
    name,
    sort_order,
    protocol,
    address_source,
    host_value,
    port,
    path,
    method,
    config_source,
    service_definition_id
)
SELECT
    c.id,
    'service',
    c.service_id,
    c.type,
    c.target,
    c.interval_seconds,
    c.timeout_seconds,
    c.expected_status_min,
    c.expected_status_max,
    c.enabled,
    COALESCE(lr.checked_at, strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    COALESCE(lr.checked_at, ''),
    COALESCE(cr.status, 'unknown'),
    0,
    c.created_at,
    c.updated_at,
    c.name,
    0,
    CASE
        WHEN c.type = 'http' AND instr(c.target, '://') > 0 THEN substr(c.target, 1, instr(c.target, '://') - 1)
        ELSE COALESCE(s.scheme, '')
    END,
    COALESCE(s.address_source, 'literal_host'),
    COALESCE(s.host_value, s.host, ''),
    COALESCE(s.port, 0),
    CASE
        WHEN c.type = 'http'
         AND instr(c.target, '://') > 0
         AND instr(substr(c.target, instr(c.target, '://') + 3), '/') > 0
            THEN substr(
                substr(c.target, instr(c.target, '://') + 3),
                instr(substr(c.target, instr(c.target, '://') + 3), '/')
            )
        ELSE COALESCE(s.path, '')
    END,
    CASE WHEN c.type = 'http' THEN 'GET' ELSE '' END,
    'migrated',
    COALESCE(s.service_definition_id, '')
FROM service_checks c
LEFT JOIN services s ON s.id = c.service_id
LEFT JOIN latest_results lr ON lr.check_id = c.id
LEFT JOIN check_results cr ON cr.check_id = c.id AND cr.checked_at = lr.checked_at
LEFT JOIN health_checks h ON h.id = c.id
WHERE h.id IS NULL;

INSERT INTO health_check_results(
    id,
    health_check_id,
    subject_type,
    subject_id,
    status,
    latency_ms,
    message,
    checked_at,
    http_status_code,
    response_size_bytes
)
SELECT
    r.id,
    r.check_id,
    'service',
    r.service_id,
    r.status,
    r.latency_ms,
    COALESCE(r.message, ''),
    r.checked_at,
    0,
    0
FROM check_results r
LEFT JOIN health_check_results hr ON hr.id = r.id
WHERE hr.id IS NULL
  AND r.checked_at >= strftime('%Y-%m-%dT%H:%M:%fZ', 'now', '-30 days');

CREATE INDEX IF NOT EXISTS idx_services_service_definition_id ON services(service_definition_id);
CREATE INDEX IF NOT EXISTS idx_discovered_services_service_definition_id ON discovered_services(service_definition_id);
CREATE INDEX IF NOT EXISTS idx_service_definitions_enabled_priority ON service_definitions(enabled, priority DESC);
CREATE INDEX IF NOT EXISTS idx_service_definition_matchers_definition_id ON service_definition_matchers(service_definition_id, sort_order);
CREATE INDEX IF NOT EXISTS idx_service_definition_check_templates_definition_id ON service_definition_check_templates(service_definition_id, sort_order);
