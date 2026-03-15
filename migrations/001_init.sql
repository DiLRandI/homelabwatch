CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS app_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS docker_endpoints (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    kind TEXT NOT NULL,
    address TEXT NOT NULL,
    tls_ca_path TEXT,
    tls_cert_path TEXT,
    tls_key_path TEXT,
    enabled INTEGER NOT NULL DEFAULT 1,
    scan_interval_seconds INTEGER NOT NULL DEFAULT 30,
    last_success_at TEXT,
    last_error TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS scan_targets (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    cidr TEXT NOT NULL UNIQUE,
    auto_detected INTEGER NOT NULL DEFAULT 0,
    enabled INTEGER NOT NULL DEFAULT 1,
    scan_interval_seconds INTEGER NOT NULL DEFAULT 300,
    common_ports TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS devices (
    id TEXT PRIMARY KEY,
    identity_key TEXT NOT NULL UNIQUE,
    primary_mac TEXT,
    hostname TEXT,
    display_name TEXT,
    identity_confidence TEXT NOT NULL,
    hidden INTEGER NOT NULL DEFAULT 0,
    first_seen_at TEXT NOT NULL,
    last_seen_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS device_addresses (
    id TEXT PRIMARY KEY,
    device_id TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    mac_address TEXT,
    interface_name TEXT,
    is_primary INTEGER NOT NULL DEFAULT 0,
    first_seen_at TEXT NOT NULL,
    last_seen_at TEXT NOT NULL,
    UNIQUE(device_id, ip_address),
    FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS device_ports (
    id TEXT PRIMARY KEY,
    device_id TEXT NOT NULL,
    port INTEGER NOT NULL,
    protocol TEXT NOT NULL,
    service_hint TEXT,
    open INTEGER NOT NULL DEFAULT 1,
    first_seen_at TEXT NOT NULL,
    last_seen_at TEXT NOT NULL,
    UNIQUE(device_id, port, protocol),
    FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS services (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    source_type TEXT NOT NULL,
    source_ref TEXT NOT NULL,
    device_id TEXT,
    icon TEXT,
    scheme TEXT,
    host TEXT NOT NULL,
    port INTEGER NOT NULL DEFAULT 0,
    path TEXT,
    url TEXT NOT NULL,
    hidden INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'unknown',
    last_seen_at TEXT,
    last_checked_at TEXT,
    details_json TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE(source_type, source_ref),
    FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS service_checks (
    id TEXT PRIMARY KEY,
    service_id TEXT NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    target TEXT NOT NULL,
    interval_seconds INTEGER NOT NULL DEFAULT 60,
    timeout_seconds INTEGER NOT NULL DEFAULT 10,
    expected_status_min INTEGER NOT NULL DEFAULT 200,
    expected_status_max INTEGER NOT NULL DEFAULT 399,
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY(service_id) REFERENCES services(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS check_results (
    id TEXT PRIMARY KEY,
    check_id TEXT NOT NULL,
    service_id TEXT NOT NULL,
    status TEXT NOT NULL,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    message TEXT,
    checked_at TEXT NOT NULL,
    FOREIGN KEY(check_id) REFERENCES service_checks(id) ON DELETE CASCADE,
    FOREIGN KEY(service_id) REFERENCES services(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS service_events (
    id TEXT PRIMARY KEY,
    service_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    status TEXT NOT NULL,
    message TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY(service_id) REFERENCES services(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS bookmarks (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    description TEXT,
    icon TEXT,
    tags_json TEXT NOT NULL DEFAULT '[]',
    sort_order INTEGER NOT NULL DEFAULT 0,
    service_id TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY(service_id) REFERENCES services(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS job_state (
    job_name TEXT PRIMARY KEY,
    last_run_at TEXT,
    last_success_at TEXT,
    last_error TEXT,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_devices_primary_mac ON devices(primary_mac);
CREATE INDEX IF NOT EXISTS idx_device_addresses_ip ON device_addresses(ip_address);
CREATE INDEX IF NOT EXISTS idx_services_device_id ON services(device_id);
CREATE INDEX IF NOT EXISTS idx_check_results_check_id ON check_results(check_id, checked_at DESC);
CREATE INDEX IF NOT EXISTS idx_service_events_service_id ON service_events(service_id, created_at DESC);
