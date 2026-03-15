PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS app_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS docker_endpoints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    kind TEXT NOT NULL DEFAULT 'remote',
    base_url TEXT NOT NULL DEFAULT '',
    socket_path TEXT NOT NULL DEFAULT '',
    tls_ca_path TEXT NOT NULL DEFAULT '',
    tls_cert_path TEXT NOT NULL DEFAULT '',
    tls_key_path TEXT NOT NULL DEFAULT '',
    enabled INTEGER NOT NULL DEFAULT 1,
    interval_sec INTEGER NOT NULL DEFAULT 30,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS scan_targets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    cidr TEXT NOT NULL UNIQUE,
    source TEXT NOT NULL DEFAULT 'configured',
    enabled INTEGER NOT NULL DEFAULT 1,
    interval_sec INTEGER NOT NULL DEFAULT 300,
    port_profile TEXT NOT NULL DEFAULT 'default',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS devices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    identity_key TEXT NOT NULL UNIQUE,
    primary_mac TEXT NOT NULL DEFAULT '',
    hostname TEXT NOT NULL DEFAULT '',
    display_name TEXT NOT NULL DEFAULT '',
    confidence TEXT NOT NULL DEFAULT 'high',
    vendor TEXT NOT NULL DEFAULT '',
    first_seen_at TEXT NOT NULL,
    last_seen_at TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS device_addresses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id INTEGER NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    ip_address TEXT NOT NULL DEFAULT '',
    mac_address TEXT NOT NULL DEFAULT '',
    interface_name TEXT NOT NULL DEFAULT '',
    cidr TEXT NOT NULL DEFAULT '',
    is_primary INTEGER NOT NULL DEFAULT 0,
    first_seen_at TEXT NOT NULL,
    last_seen_at TEXT NOT NULL,
    UNIQUE(device_id, ip_address, mac_address)
);

CREATE TABLE IF NOT EXISTS device_ports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id INTEGER NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    port INTEGER NOT NULL,
    protocol TEXT NOT NULL,
    service_hint TEXT NOT NULL DEFAULT '',
    last_seen_at TEXT NOT NULL,
    UNIQUE(device_id, port, protocol)
);

CREATE TABLE IF NOT EXISTS services (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    kind TEXT NOT NULL DEFAULT '',
    origin TEXT NOT NULL,
    source_ref TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    device_id INTEGER REFERENCES devices(id) ON DELETE SET NULL,
    scheme TEXT NOT NULL DEFAULT '',
    host TEXT NOT NULL DEFAULT '',
    port INTEGER NOT NULL DEFAULT 0,
    url TEXT NOT NULL DEFAULT '',
    icon TEXT NOT NULL DEFAULT '',
    labels_json TEXT NOT NULL DEFAULT '[]',
    hidden INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'unknown',
    last_seen_at TEXT NOT NULL DEFAULT '',
    last_check_at TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(origin, source_ref)
);

CREATE TABLE IF NOT EXISTS service_checks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    target TEXT NOT NULL DEFAULT '',
    interval_sec INTEGER NOT NULL DEFAULT 60,
    timeout_ms INTEGER NOT NULL DEFAULT 3000,
    config_json TEXT NOT NULL DEFAULT '',
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS check_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    check_id INTEGER NOT NULL REFERENCES service_checks(id) ON DELETE CASCADE,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    success INTEGER NOT NULL,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    message TEXT NOT NULL DEFAULT '',
    observed_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS service_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    service_id INTEGER REFERENCES services(id) ON DELETE CASCADE,
    device_id INTEGER REFERENCES devices(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    message TEXT NOT NULL,
    payload_json TEXT NOT NULL DEFAULT '',
    occurred_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS bookmarks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    service_id INTEGER REFERENCES services(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    tags_json TEXT NOT NULL DEFAULT '[]',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS job_state (
    job_name TEXT PRIMARY KEY,
    last_run_at TEXT NOT NULL DEFAULT '',
    last_error TEXT NOT NULL DEFAULT '',
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_services_device_id ON services(device_id);
CREATE INDEX IF NOT EXISTS idx_service_checks_service_id ON service_checks(service_id);
CREATE INDEX IF NOT EXISTS idx_check_results_service_id ON check_results(service_id);
CREATE INDEX IF NOT EXISTS idx_service_events_service_id ON service_events(service_id);
CREATE INDEX IF NOT EXISTS idx_device_addresses_device_id ON device_addresses(device_id);
CREATE INDEX IF NOT EXISTS idx_device_ports_device_id ON device_ports(device_id);
