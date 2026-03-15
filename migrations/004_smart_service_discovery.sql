ALTER TABLE services ADD COLUMN origin_discovered_service_id TEXT;
ALTER TABLE services ADD COLUMN service_type TEXT NOT NULL DEFAULT '';
ALTER TABLE services ADD COLUMN address_source TEXT NOT NULL DEFAULT 'literal_host';
ALTER TABLE services ADD COLUMN host_value TEXT NOT NULL DEFAULT '';

UPDATE services
SET host_value = host
WHERE COALESCE(host_value, '') = '';

ALTER TABLE bookmarks ADD COLUMN discovered_service_id TEXT;

CREATE TABLE IF NOT EXISTS discovered_services (
    id TEXT PRIMARY KEY,
    device_id TEXT,
    merge_key TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    service_type TEXT NOT NULL DEFAULT '',
    confidence_score INTEGER NOT NULL DEFAULT 0,
    address_source TEXT NOT NULL DEFAULT 'literal_host',
    host_value TEXT NOT NULL DEFAULT '',
    scheme TEXT NOT NULL DEFAULT '',
    port INTEGER NOT NULL DEFAULT 0,
    path TEXT NOT NULL DEFAULT '',
    url TEXT NOT NULL DEFAULT '',
    icon TEXT NOT NULL DEFAULT '',
    state TEXT NOT NULL DEFAULT 'pending',
    ignore_fingerprint TEXT NOT NULL DEFAULT '',
    automation_mode TEXT NOT NULL DEFAULT 'manual',
    status TEXT NOT NULL DEFAULT 'unknown',
    last_checked_at TEXT,
    last_fingerprinted_at TEXT,
    accepted_service_id TEXT,
    accepted_bookmark_id TEXT,
    details_json TEXT NOT NULL DEFAULT '{}',
    first_seen_at TEXT NOT NULL,
    last_seen_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE SET NULL,
    FOREIGN KEY(accepted_service_id) REFERENCES services(id) ON DELETE SET NULL,
    FOREIGN KEY(accepted_bookmark_id) REFERENCES bookmarks(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS discovered_service_evidence (
    id TEXT PRIMARY KEY,
    discovered_service_id TEXT NOT NULL,
    source_type TEXT NOT NULL,
    source_ref TEXT NOT NULL,
    service_type_hint TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL DEFAULT '',
    host TEXT NOT NULL DEFAULT '',
    port INTEGER NOT NULL DEFAULT 0,
    path TEXT NOT NULL DEFAULT '',
    url TEXT NOT NULL DEFAULT '',
    fingerprint_hash TEXT NOT NULL DEFAULT '',
    details_json TEXT NOT NULL DEFAULT '{}',
    first_seen_at TEXT NOT NULL,
    last_seen_at TEXT NOT NULL,
    UNIQUE(source_type, source_ref),
    FOREIGN KEY(discovered_service_id) REFERENCES discovered_services(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS health_checks (
    id TEXT PRIMARY KEY,
    subject_type TEXT NOT NULL,
    subject_id TEXT NOT NULL,
    kind TEXT NOT NULL,
    target TEXT NOT NULL DEFAULT '',
    interval_seconds INTEGER NOT NULL DEFAULT 60,
    timeout_seconds INTEGER NOT NULL DEFAULT 10,
    expected_status_min INTEGER NOT NULL DEFAULT 200,
    expected_status_max INTEGER NOT NULL DEFAULT 399,
    enabled INTEGER NOT NULL DEFAULT 1,
    next_run_at TEXT,
    last_run_at TEXT,
    last_status TEXT NOT NULL DEFAULT 'unknown',
    consecutive_failures INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE(subject_type, subject_id, kind, target)
);

CREATE TABLE IF NOT EXISTS health_check_results (
    id TEXT PRIMARY KEY,
    health_check_id TEXT NOT NULL,
    subject_type TEXT NOT NULL,
    subject_id TEXT NOT NULL,
    status TEXT NOT NULL,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    message TEXT NOT NULL DEFAULT '',
    checked_at TEXT NOT NULL,
    FOREIGN KEY(health_check_id) REFERENCES health_checks(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_services_origin_discovered_service_id ON services(origin_discovered_service_id);
CREATE INDEX IF NOT EXISTS idx_services_address_source ON services(address_source, device_id);
CREATE INDEX IF NOT EXISTS idx_bookmarks_discovered_service_id ON bookmarks(discovered_service_id);
CREATE INDEX IF NOT EXISTS idx_discovered_services_state_last_seen ON discovered_services(state, last_seen_at DESC);
CREATE INDEX IF NOT EXISTS idx_discovered_services_device_id ON discovered_services(device_id);
CREATE INDEX IF NOT EXISTS idx_discovered_service_evidence_discovered_id ON discovered_service_evidence(discovered_service_id);
CREATE INDEX IF NOT EXISTS idx_health_checks_due ON health_checks(enabled, next_run_at);
CREATE INDEX IF NOT EXISTS idx_health_check_results_checked_at ON health_check_results(health_check_id, checked_at DESC);
