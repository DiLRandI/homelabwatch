CREATE TABLE IF NOT EXISTS topology_sources (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    port INTEGER NOT NULL DEFAULT 161,
    enabled INTEGER NOT NULL DEFAULT 1,
    poll_interval_seconds INTEGER NOT NULL DEFAULT 300,
    timeout_ms INTEGER NOT NULL DEFAULT 1500,
    retries INTEGER NOT NULL DEFAULT 1,
    snmp_version TEXT NOT NULL,
    community TEXT,
    username TEXT,
    auth_protocol TEXT,
    auth_passphrase TEXT,
    privacy_protocol TEXT,
    privacy_passphrase TEXT,
    role TEXT NOT NULL DEFAULT 'unknown',
    root INTEGER NOT NULL DEFAULT 0,
    last_success_at TEXT,
    last_error TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_topology_sources_enabled ON topology_sources(enabled);
CREATE INDEX IF NOT EXISTS idx_topology_sources_address ON topology_sources(address);

CREATE TABLE IF NOT EXISTS topology_lldp_links (
    id TEXT PRIMARY KEY,
    source_id TEXT NOT NULL,
    local_chassis_id TEXT,
    local_system_name TEXT,
    local_port_id TEXT,
    local_port_name TEXT,
    local_port_description TEXT,
    local_if_index INTEGER,
    remote_chassis_id TEXT,
    remote_system_name TEXT,
    remote_port_id TEXT,
    remote_port_description TEXT,
    remote_management_address TEXT,
    last_seen_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY(source_id) REFERENCES topology_sources(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_topology_lldp_links_unique ON topology_lldp_links(source_id, local_port_id, remote_chassis_id, remote_port_id);

CREATE TABLE IF NOT EXISTS topology_mac_links (
    id TEXT PRIMARY KEY,
    source_id TEXT NOT NULL,
    mac_address TEXT NOT NULL,
    vlan INTEGER,
    bridge_port INTEGER,
    if_index INTEGER,
    if_name TEXT,
    if_description TEXT,
    status TEXT,
    last_seen_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY(source_id) REFERENCES topology_sources(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_topology_mac_links_mac ON topology_mac_links(mac_address);
CREATE INDEX IF NOT EXISTS idx_topology_mac_links_source_if ON topology_mac_links(source_id, if_index);

CREATE TABLE IF NOT EXISTS topology_interfaces (
    id TEXT PRIMARY KEY,
    source_id TEXT NOT NULL,
    if_index INTEGER NOT NULL,
    if_name TEXT,
    if_description TEXT,
    if_alias TEXT,
    if_type INTEGER,
    oper_status TEXT,
    speed_bps INTEGER,
    last_seen_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY(source_id) REFERENCES topology_sources(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_topology_interfaces_unique ON topology_interfaces(source_id, if_index);
