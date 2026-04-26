CREATE TABLE IF NOT EXISTS status_pages (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS status_page_services (
    status_page_id TEXT NOT NULL,
    service_id TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    display_name TEXT NOT NULL DEFAULT '',
    PRIMARY KEY(status_page_id, service_id),
    FOREIGN KEY(status_page_id) REFERENCES status_pages(id) ON DELETE CASCADE,
    FOREIGN KEY(service_id) REFERENCES services(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS status_page_announcements (
    id TEXT PRIMARY KEY,
    status_page_id TEXT NOT NULL,
    kind TEXT NOT NULL CHECK(kind IN ('info', 'maintenance', 'incident', 'resolved')),
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    starts_at TEXT,
    ends_at TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY(status_page_id) REFERENCES status_pages(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_status_pages_slug ON status_pages(slug);
CREATE INDEX IF NOT EXISTS idx_status_page_services_service_id ON status_page_services(service_id);
CREATE INDEX IF NOT EXISTS idx_status_page_services_page_order ON status_page_services(status_page_id, sort_order);
CREATE INDEX IF NOT EXISTS idx_status_page_announcements_page_window ON status_page_announcements(status_page_id, starts_at, ends_at);
