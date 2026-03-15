CREATE TABLE IF NOT EXISTS folders (
    id TEXT PRIMARY KEY,
    parent_id TEXT,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY(parent_id) REFERENCES folders(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS tags (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL COLLATE NOCASE UNIQUE,
    slug TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

ALTER TABLE bookmarks RENAME TO bookmarks_legacy;

CREATE TABLE bookmarks (
    id TEXT PRIMARY KEY,
    folder_id TEXT,
    service_id TEXT,
    device_id TEXT,
    manual_name TEXT,
    manual_url TEXT,
    description TEXT,
    icon_mode TEXT NOT NULL DEFAULT 'auto',
    icon_value TEXT,
    use_device_primary_address INTEGER NOT NULL DEFAULT 0,
    scheme TEXT,
    host TEXT,
    port INTEGER NOT NULL DEFAULT 0,
    path TEXT,
    position INTEGER NOT NULL DEFAULT 0,
    is_favorite INTEGER NOT NULL DEFAULT 0,
    favorite_position INTEGER,
    click_count INTEGER NOT NULL DEFAULT 0,
    last_opened_at TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY(folder_id) REFERENCES folders(id) ON DELETE SET NULL,
    FOREIGN KEY(service_id) REFERENCES services(id) ON DELETE SET NULL,
    FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS bookmark_tags (
    bookmark_id TEXT NOT NULL,
    tag_id TEXT NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY(bookmark_id, tag_id),
    FOREIGN KEY(bookmark_id) REFERENCES bookmarks(id) ON DELETE CASCADE,
    FOREIGN KEY(tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

INSERT INTO bookmarks(
    id,
    service_id,
    manual_name,
    manual_url,
    description,
    icon_mode,
    icon_value,
    position,
    created_at,
    updated_at
)
SELECT
    id,
    NULLIF(service_id, ''),
    name,
    url,
    NULLIF(description, ''),
    CASE
        WHEN COALESCE(icon, '') <> '' THEN 'external'
        ELSE 'auto'
    END,
    NULLIF(icon, ''),
    sort_order,
    created_at,
    updated_at
FROM bookmarks_legacy;

INSERT OR IGNORE INTO tags(id, name, slug, created_at, updated_at)
SELECT
    'tag_' || lower(hex(randomblob(8))),
    tag_name,
    lower(replace(replace(replace(replace(tag_name, ' ', '-'), '_', '-'), '/', '-'), ':', '-')),
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM (
    SELECT DISTINCT trim(CAST(value AS TEXT)) AS tag_name
    FROM bookmarks_legacy, json_each(bookmarks_legacy.tags_json)
    WHERE trim(CAST(value AS TEXT)) <> ''
);

INSERT OR IGNORE INTO bookmark_tags(bookmark_id, tag_id, created_at)
SELECT
    b.id,
    t.id,
    CURRENT_TIMESTAMP
FROM bookmarks_legacy b
JOIN json_each(b.tags_json) j
JOIN tags t ON t.name = trim(CAST(j.value AS TEXT))
WHERE trim(CAST(j.value AS TEXT)) <> '';

DROP TABLE bookmarks_legacy;

CREATE INDEX IF NOT EXISTS idx_folders_parent_position ON folders(parent_id, position, name);
CREATE INDEX IF NOT EXISTS idx_bookmarks_folder_position ON bookmarks(folder_id, position, manual_name);
CREATE INDEX IF NOT EXISTS idx_bookmarks_favorite_position ON bookmarks(is_favorite, favorite_position, position);
CREATE INDEX IF NOT EXISTS idx_bookmarks_device_id ON bookmarks(device_id);
CREATE INDEX IF NOT EXISTS idx_bookmark_tags_tag_id ON bookmark_tags(tag_id, bookmark_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_bookmarks_service_unique ON bookmarks(service_id) WHERE service_id IS NOT NULL;
