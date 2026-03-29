ALTER TABLE services ADD COLUMN health_address_source TEXT NOT NULL DEFAULT '';
ALTER TABLE services ADD COLUMN health_host_value TEXT NOT NULL DEFAULT '';
ALTER TABLE services ADD COLUMN health_scheme TEXT NOT NULL DEFAULT '';
ALTER TABLE services ADD COLUMN health_port INTEGER NOT NULL DEFAULT 0;
ALTER TABLE services ADD COLUMN health_path TEXT NOT NULL DEFAULT '';

UPDATE services
SET
    health_address_source = COALESCE(NULLIF(address_source, ''), 'literal_host'),
    health_host_value = COALESCE(NULLIF(host_value, ''), host, ''),
    health_scheme = COALESCE(scheme, ''),
    health_port = COALESCE(port, 0),
    health_path = COALESCE(path, '')
WHERE
    COALESCE(health_address_source, '') = ''
    AND COALESCE(health_host_value, '') = ''
    AND COALESCE(health_scheme, '') = ''
    AND COALESCE(health_port, 0) = 0
    AND COALESCE(health_path, '') = '';
