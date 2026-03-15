package sqlite

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
)

type bookmarkRecord struct {
	ID                      string
	FolderID                string
	FolderName              string
	ServiceID               string
	ServiceName             string
	ServiceSource           domain.ServiceSource
	ServiceSourceRef        string
	ServiceHidden           bool
	ServiceURL              string
	ServiceAddressSource    domain.ServiceAddressSource
	ServiceHostValue        string
	ServiceScheme           string
	ServiceHost             string
	ServicePort             int
	ServicePath             string
	ServiceIcon             string
	ServiceDeviceID         string
	ServiceDeviceName       string
	ServiceHealthStatus     domain.HealthStatus
	DeviceID                string
	DeviceName              string
	ManualName              string
	ManualURL               string
	Description             string
	IconMode                string
	IconValue               string
	UseDevicePrimaryAddress bool
	Scheme                  string
	Host                    string
	Port                    int
	Path                    string
	Position                int
	IsFavorite              bool
	FavoritePosition        int
	ClickCount              int
	LastOpenedAt            time.Time
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

func (s *Store) ListBookmarks(ctx context.Context) ([]domain.Bookmark, error) {
	return s.QueryBookmarks(ctx, domain.BookmarkListOptions{})
}

func (s *Store) QueryBookmarks(ctx context.Context, options domain.BookmarkListOptions) ([]domain.Bookmark, error) {
	records, err := s.loadBookmarkRecords(ctx)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return []domain.Bookmark{}, nil
	}
	tagsByBookmark, err := s.loadBookmarkTags(ctx)
	if err != nil {
		return nil, err
	}
	deviceAddresses, err := s.loadPrimaryDeviceAddresses(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]domain.Bookmark, 0, len(records))
	for _, record := range records {
		item := buildBookmark(record, tagsByBookmark[record.ID], deviceAddresses)
		if bookmarkMatchesFilters(item, options) {
			items = append(items, item)
		}
	}
	sortBookmarks(items)
	return items, nil
}

func (s *Store) GetBookmark(ctx context.Context, id string) (domain.Bookmark, error) {
	items, err := s.QueryBookmarks(ctx, domain.BookmarkListOptions{})
	if err != nil {
		return domain.Bookmark{}, err
	}
	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}
	return domain.Bookmark{}, sql.ErrNoRows
}

func (s *Store) SaveBookmark(ctx context.Context, input domain.BookmarkInput) (domain.Bookmark, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Bookmark{}, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	existing, existingFound, err := s.lookupBookmarkTx(ctx, tx, input.ID)
	if err != nil {
		return domain.Bookmark{}, err
	}
	record := bookmarkRecord{}
	if existingFound {
		record = existing
	} else {
		record = bookmarkRecord{
			ID:           firstNonEmpty(input.ID, newID("bmk")),
			CreatedAt:    now,
			Position:     input.Position,
			UpdatedAt:    now,
			IconMode:     "auto",
			LastOpenedAt: time.Time{},
		}
	}

	manualName := strings.TrimSpace(input.Name)
	manualURL := strings.TrimSpace(input.URL)
	description := strings.TrimSpace(input.Description)
	iconMode := firstNonEmpty(strings.TrimSpace(input.IconMode), record.IconMode, "auto")
	iconValue := strings.TrimSpace(input.IconValue)
	folderID := strings.TrimSpace(input.FolderID)
	serviceID := strings.TrimSpace(input.ServiceID)
	deviceID := strings.TrimSpace(input.DeviceID)
	useDevicePrimaryAddress := input.UseDevicePrimaryAddress

	if input.Monitor != nil && input.Monitor.Enabled {
		if strings.TrimSpace(input.Monitor.ServiceID) != "" {
			serviceID = strings.TrimSpace(input.Monitor.ServiceID)
		} else {
			service, err := s.upsertBookmarkOwnedServiceTx(ctx, tx, record.ID, bookmarkServiceInput{
				DeviceID:   deviceID,
				Name:       firstNonEmpty(strings.TrimSpace(input.Monitor.ServiceName), manualName),
				URL:        manualURL,
				Visible:    input.Monitor.ServiceVisible,
				ExistingID: record.ServiceID,
			})
			if err != nil {
				return domain.Bookmark{}, err
			}
			serviceID = service.ID
		}
	}

	if serviceID != "" {
		if bookmarkID, err := s.findBookmarkIDByServiceTx(ctx, tx, serviceID); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return domain.Bookmark{}, err
		} else if err == nil && bookmarkID != record.ID {
			record.ID = bookmarkID
			reloaded, found, lookupErr := s.lookupBookmarkTx(ctx, tx, bookmarkID)
			if lookupErr != nil {
				return domain.Bookmark{}, lookupErr
			}
			if found {
				record = reloaded
			}
		}
		deviceID = ""
		useDevicePrimaryAddress = false
	}

	scheme, host, port, path, err := parseBookmarkURL(manualURL)
	if err != nil && manualURL != "" {
		return domain.Bookmark{}, err
	}
	if manualName == "" {
		manualName = firstNonEmpty(record.ManualName, serviceDisplayName(record), host)
	}
	if serviceID == "" && manualURL == "" {
		return domain.Bookmark{}, errors.New("bookmark url is required")
	}
	if manualName == "" && serviceID == "" {
		return domain.Bookmark{}, errors.New("bookmark name is required")
	}

	if !existingFound {
		record.CreatedAt = now
	}
	record.FolderID = folderID
	record.ServiceID = serviceID
	record.DeviceID = deviceID
	record.ManualName = manualName
	record.ManualURL = manualURL
	record.Description = description
	record.IconMode = iconMode
	record.IconValue = iconValue
	record.UseDevicePrimaryAddress = useDevicePrimaryAddress
	record.Scheme = scheme
	record.Host = host
	record.Port = port
	record.Path = path
	record.IsFavorite = input.IsFavorite
	record.Position = input.Position
	record.FavoritePosition = input.FavoritePosition
	record.UpdatedAt = now

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO bookmarks(
			id, folder_id, service_id, device_id, manual_name, manual_url, description, icon_mode,
			icon_value, use_device_primary_address, scheme, host, port, path, position, is_favorite,
			favorite_position, click_count, last_opened_at, created_at, updated_at
		) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			folder_id = excluded.folder_id,
			service_id = excluded.service_id,
			device_id = excluded.device_id,
			manual_name = excluded.manual_name,
			manual_url = excluded.manual_url,
			description = excluded.description,
			icon_mode = excluded.icon_mode,
			icon_value = excluded.icon_value,
			use_device_primary_address = excluded.use_device_primary_address,
			scheme = excluded.scheme,
			host = excluded.host,
			port = excluded.port,
			path = excluded.path,
			position = excluded.position,
			is_favorite = excluded.is_favorite,
			favorite_position = excluded.favorite_position,
			updated_at = excluded.updated_at
	`,
		record.ID,
		nullableString(record.FolderID),
		nullableString(record.ServiceID),
		nullableString(record.DeviceID),
		nullableString(record.ManualName),
		nullableString(record.ManualURL),
		nullableString(record.Description),
		record.IconMode,
		nullableString(record.IconValue),
		boolInt(record.UseDevicePrimaryAddress),
		nullableString(record.Scheme),
		nullableString(record.Host),
		record.Port,
		nullableString(record.Path),
		record.Position,
		boolInt(record.IsFavorite),
		nullableInt(record.FavoritePosition),
		record.ClickCount,
		nullableTime(record.LastOpenedAt),
		record.CreatedAt.Format(time.RFC3339Nano),
		record.UpdatedAt.Format(time.RFC3339Nano),
	); err != nil {
		return domain.Bookmark{}, err
	}
	if err := s.syncBookmarkTagsTx(ctx, tx, record.ID, input.Tags, now); err != nil {
		return domain.Bookmark{}, err
	}
	if existingFound && existing.ServiceID != "" && existing.ServiceID != record.ServiceID {
		if err := s.deleteBookmarkOwnedServiceTx(ctx, tx, existing.ServiceID, record.ID); err != nil {
			return domain.Bookmark{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return domain.Bookmark{}, err
	}
	return s.GetBookmark(ctx, record.ID)
}

func (s *Store) CreateBookmarkFromService(ctx context.Context, input domain.CreateBookmarkFromServiceInput) (domain.Bookmark, error) {
	if strings.TrimSpace(input.ServiceID) == "" {
		return domain.Bookmark{}, errors.New("service id is required")
	}
	item, err := s.SaveBookmark(ctx, domain.BookmarkInput{
		ServiceID:        input.ServiceID,
		FolderID:         input.FolderID,
		Tags:             input.Tags,
		Name:             input.Name,
		IconMode:         input.IconMode,
		IconValue:        input.IconValue,
		IsFavorite:       input.IsFavorite,
		FavoritePosition: input.FavoritePosition,
	})
	if err != nil {
		return domain.Bookmark{}, err
	}
	return item, nil
}

func (s *Store) DeleteBookmark(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	record, found, err := s.lookupBookmarkTx(ctx, tx, id)
	if err != nil {
		return err
	}
	if !found {
		return sql.ErrNoRows
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM bookmark_tags WHERE bookmark_id = ?", id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM bookmarks WHERE id = ?", id); err != nil {
		return err
	}
	if record.ServiceID != "" {
		if err := s.deleteBookmarkOwnedServiceTx(ctx, tx, record.ServiceID, id); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM tags
		WHERE id IN (
			SELECT t.id FROM tags t
			LEFT JOIN bookmark_tags bt ON bt.tag_id = t.id
			GROUP BY t.id
			HAVING COUNT(bt.bookmark_id) = 0
		)
	`); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) OpenBookmark(ctx context.Context, id string) (domain.Bookmark, error) {
	now := time.Now().UTC()
	if _, err := s.db.ExecContext(ctx, `
		UPDATE bookmarks
		SET click_count = click_count + 1, last_opened_at = ?, updated_at = ?
		WHERE id = ?
	`, now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano), id); err != nil {
		return domain.Bookmark{}, err
	}
	return s.GetBookmark(ctx, id)
}

func (s *Store) ReorderBookmarks(ctx context.Context, items []domain.BookmarkReorderItem) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := nowString()
	for _, item := range items {
		if _, err := tx.ExecContext(ctx, `
			UPDATE bookmarks
			SET folder_id = ?, position = ?, is_favorite = ?, favorite_position = ?, updated_at = ?
			WHERE id = ?
		`, nullableString(item.FolderID), item.Position, boolInt(item.IsFavorite), nullableInt(item.FavoritePosition), now, item.ID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ListFolders(ctx context.Context) ([]domain.Folder, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			f.id,
			COALESCE(f.parent_id, ''),
			f.name,
			f.slug,
			f.position,
			COALESCE((
				SELECT COUNT(1) FROM bookmarks b WHERE b.folder_id = f.id
			), 0),
			COALESCE((
				SELECT COUNT(1) FROM folders c WHERE c.parent_id = f.id
			), 0),
			f.created_at,
			f.updated_at
		FROM folders f
		ORDER BY COALESCE(f.parent_id, ''), f.position, f.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.Folder
	for rows.Next() {
		var item domain.Folder
		var createdAt, updatedAt string
		if err := rows.Scan(&item.ID, &item.ParentID, &item.Name, &item.Slug, &item.Position, &item.BookmarkCount, &item.ChildCount, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		item.CreatedAt = parseTime(createdAt)
		item.UpdatedAt = parseTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) SaveFolder(ctx context.Context, input domain.FolderInput) (domain.Folder, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Folder{}, err
	}
	defer tx.Rollback()
	now := time.Now().UTC()
	folderID := firstNonEmpty(strings.TrimSpace(input.ID), newID("fld"))
	parentID := strings.TrimSpace(input.ParentID)
	if parentID == folderID {
		return domain.Folder{}, errors.New("folder cannot be its own parent")
	}
	if parentID != "" {
		if err := s.ensureFolderParentAllowedTx(ctx, tx, folderID, parentID); err != nil {
			return domain.Folder{}, err
		}
	}
	createdAt := now
	if existing, found, err := s.getFolderTx(ctx, tx, folderID); err != nil {
		return domain.Folder{}, err
	} else if found {
		createdAt = existing.CreatedAt
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO folders(id, parent_id, name, slug, position, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			parent_id = excluded.parent_id,
			name = excluded.name,
			slug = excluded.slug,
			position = excluded.position,
			updated_at = excluded.updated_at
	`, folderID, nullableString(parentID), strings.TrimSpace(input.Name), slugify(input.Name), input.Position, createdAt.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano)); err != nil {
		return domain.Folder{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Folder{}, err
	}
	return s.GetFolder(ctx, folderID)
}

func (s *Store) GetFolder(ctx context.Context, id string) (domain.Folder, error) {
	items, err := s.ListFolders(ctx)
	if err != nil {
		return domain.Folder{}, err
	}
	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}
	return domain.Folder{}, sql.ErrNoRows
}

func (s *Store) DeleteFolder(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var parentID string
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(parent_id, '') FROM folders WHERE id = ?`, id).Scan(&parentID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE folders SET parent_id = ? WHERE parent_id = ?`, nullableString(parentID), id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE bookmarks SET folder_id = ? WHERE folder_id = ?`, nullableString(parentID), id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM folders WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) ReorderFolders(ctx context.Context, items []domain.FolderReorderItem) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := nowString()
	for _, item := range items {
		if strings.TrimSpace(item.ParentID) == item.ID {
			return errors.New("folder cannot be its own parent")
		}
		if item.ParentID != "" {
			if err := s.ensureFolderParentAllowedTx(ctx, tx, item.ID, item.ParentID); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, `UPDATE folders SET parent_id = ?, position = ?, updated_at = ? WHERE id = ?`, nullableString(item.ParentID), item.Position, now, item.ID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ListTags(ctx context.Context) ([]domain.Tag, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.id, t.name, t.slug, COUNT(bt.bookmark_id), t.created_at, t.updated_at
		FROM tags t
		LEFT JOIN bookmark_tags bt ON bt.tag_id = t.id
		GROUP BY t.id, t.name, t.slug, t.created_at, t.updated_at
		ORDER BY t.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.Tag
	for rows.Next() {
		var item domain.Tag
		var createdAt, updatedAt string
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.BookmarkCount, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		item.CreatedAt = parseTime(createdAt)
		item.UpdatedAt = parseTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

type bookmarkServiceInput struct {
	DeviceID   string
	Name       string
	URL        string
	Visible    bool
	ExistingID string
}

func (s *Store) upsertBookmarkOwnedServiceTx(ctx context.Context, tx *sql.Tx, bookmarkID string, input bookmarkServiceInput) (domain.Service, error) {
	service := normalizeServiceInput(domain.Service{
		ID:       strings.TrimSpace(input.ExistingID),
		Name:     input.Name,
		DeviceID: input.DeviceID,
		Hidden:   !input.Visible,
		URL:      input.URL,
	})
	if service.Name == "" {
		return domain.Service{}, errors.New("service name is required for monitored bookmarks")
	}
	if service.URL == "" {
		return domain.Service{}, errors.New("service url is required for monitored bookmarks")
	}
	now := time.Now().UTC()
	if service.ID == "" {
		service.ID = newID("svc")
		service.CreatedAt = now
	} else {
		existing, err := s.getServiceTx(ctx, tx, service.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return domain.Service{}, err
		}
		if err == nil {
			service.CreatedAt = existing.CreatedAt
		} else {
			service.CreatedAt = now
		}
	}
	service.Source = domain.ServiceSourceManual
	if service.SourceRef == "" {
		service.SourceRef = service.ID
	}
	service.Slug = slugify(service.Name)
	service.Status = domain.HealthStatusUnknown
	if service.HealthConfigMode == "" {
		service.HealthConfigMode = domain.HealthConfigModeAuto
	}
	if service.AddressSource == "" {
		service.AddressSource = domain.ServiceAddressLiteralHost
	}
	if service.AddressSource == domain.ServiceAddressLiteralHost && strings.TrimSpace(service.Host) != "" {
		service.HostValue = service.Host
	} else if service.HostValue == "" {
		service.HostValue = firstNonEmpty(service.Host, extractHost(service.URL))
	}
	service.Host = resolveAddressSourceHost(service.AddressSource, service.HostValue, "")
	service.UpdatedAt = now
	if service.Details == nil {
		service.Details = map[string]any{}
	}
	service.Details["bookmarkOwnerID"] = bookmarkID
	service.Details["bookmarkManaged"] = true
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO services(
			id, name, slug, source_type, source_ref, origin_discovered_service_id, service_definition_id, service_type, health_config_mode, address_source,
			host_value, device_id, icon, scheme, host, port, path, url, hidden, status, last_seen_at,
			last_checked_at, fingerprinted_at, details_json, created_at, updated_at
		) VALUES(?, ?, ?, ?, ?, NULL, NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			slug = excluded.slug,
			service_type = excluded.service_type,
			health_config_mode = excluded.health_config_mode,
			address_source = excluded.address_source,
			host_value = excluded.host_value,
			device_id = excluded.device_id,
			scheme = excluded.scheme,
			host = excluded.host,
			port = excluded.port,
			path = excluded.path,
			url = excluded.url,
			hidden = excluded.hidden,
			details_json = excluded.details_json,
			updated_at = excluded.updated_at
	`, service.ID, service.Name, service.Slug, service.Source, service.SourceRef, service.ServiceType, service.HealthConfigMode, service.AddressSource, service.HostValue, nullableString(service.DeviceID), nullableString(service.Icon), nullableString(service.Scheme), service.Host, service.Port, nullableString(service.Path), service.URL, boolInt(service.Hidden), service.Status, nullableTime(service.LastSeenAt), nullableTime(service.LastCheckedAt), string(mustJSON(service.Details)), service.CreatedAt.Format(time.RFC3339Nano), service.UpdatedAt.Format(time.RFC3339Nano)); err != nil {
		return domain.Service{}, err
	}
	if err := s.ensureDefaultCheckTx(ctx, tx, service); err != nil {
		return domain.Service{}, err
	}
	return s.getServiceTx(ctx, tx, service.ID)
}

func (s *Store) getServiceTx(ctx context.Context, tx *sql.Tx, id string) (domain.Service, error) {
	row := tx.QueryRowContext(ctx, `SELECT s.id, s.name, s.slug, s.source_type, s.source_ref, COALESCE(s.origin_discovered_service_id, ''), COALESCE(s.service_definition_id, ''), COALESCE(s.service_type, ''), COALESCE(s.health_config_mode, 'auto'), COALESCE(s.address_source, 'literal_host'), COALESCE(s.host_value, s.host, ''), COALESCE(s.device_id, ''), COALESCE(d.display_name, d.hostname, ''), COALESCE(s.icon, ''), COALESCE(s.scheme, ''), s.host, s.port, COALESCE(s.path, ''), s.url, s.hidden, s.status, COALESCE(s.last_seen_at, ''), COALESCE(s.last_checked_at, ''), COALESCE(s.fingerprinted_at, ''), s.details_json, s.created_at, s.updated_at FROM services s LEFT JOIN devices d ON d.id = s.device_id WHERE s.id = ?`, id)
	return scanService(row)
}

func (s *Store) FindServiceBySource(ctx context.Context, source domain.ServiceSource, sourceRef string) (domain.Service, error) {
	row := s.db.QueryRowContext(ctx, `SELECT s.id, s.name, s.slug, s.source_type, s.source_ref, COALESCE(s.origin_discovered_service_id, ''), COALESCE(s.service_definition_id, ''), COALESCE(s.service_type, ''), COALESCE(s.health_config_mode, 'auto'), COALESCE(s.address_source, 'literal_host'), COALESCE(s.host_value, s.host, ''), COALESCE(s.device_id, ''), COALESCE(d.display_name, d.hostname, ''), COALESCE(s.icon, ''), COALESCE(s.scheme, ''), s.host, s.port, COALESCE(s.path, ''), s.url, s.hidden, s.status, COALESCE(s.last_seen_at, ''), COALESCE(s.last_checked_at, ''), COALESCE(s.fingerprinted_at, ''), s.details_json, s.created_at, s.updated_at FROM services s LEFT JOIN devices d ON d.id = s.device_id WHERE s.source_type = ? AND s.source_ref = ?`, source, sourceRef)
	return scanService(row)
}

func (s *Store) loadBookmarkRecords(ctx context.Context) ([]bookmarkRecord, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			b.id,
			COALESCE(b.folder_id, ''),
			COALESCE(f.name, ''),
			COALESCE(b.service_id, ''),
			COALESCE(s.name, ''),
			COALESCE(s.source_type, ''),
			COALESCE(s.source_ref, ''),
			COALESCE(s.hidden, 0),
			COALESCE(s.url, ''),
			COALESCE(s.address_source, 'literal_host'),
			COALESCE(s.host_value, s.host, ''),
			COALESCE(s.scheme, ''),
			COALESCE(s.host, ''),
			COALESCE(s.port, 0),
			COALESCE(s.path, ''),
			COALESCE(s.icon, ''),
			COALESCE(s.device_id, ''),
			COALESCE(sd.display_name, sd.hostname, ''),
			COALESCE(s.status, 'unknown'),
			COALESCE(b.device_id, ''),
			COALESCE(md.display_name, md.hostname, ''),
			COALESCE(b.manual_name, ''),
			COALESCE(b.manual_url, ''),
			COALESCE(b.description, ''),
			b.icon_mode,
			COALESCE(b.icon_value, ''),
			b.use_device_primary_address,
			COALESCE(b.scheme, ''),
			COALESCE(b.host, ''),
			b.port,
			COALESCE(b.path, ''),
			b.position,
			b.is_favorite,
			COALESCE(b.favorite_position, 0),
			b.click_count,
			COALESCE(b.last_opened_at, ''),
			b.created_at,
			b.updated_at
		FROM bookmarks b
		LEFT JOIN folders f ON f.id = b.folder_id
		LEFT JOIN services s ON s.id = b.service_id
		LEFT JOIN devices sd ON sd.id = s.device_id
		LEFT JOIN devices md ON md.id = b.device_id
		ORDER BY b.is_favorite DESC, COALESCE(b.favorite_position, b.position), b.position, COALESCE(b.manual_name, s.name)
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []bookmarkRecord
	for rows.Next() {
		var item bookmarkRecord
		var serviceHidden, useDevicePrimary, isFavorite int
		var lastOpenedAt, createdAt, updatedAt string
		if err := rows.Scan(
			&item.ID,
			&item.FolderID,
			&item.FolderName,
			&item.ServiceID,
			&item.ServiceName,
			&item.ServiceSource,
			&item.ServiceSourceRef,
			&serviceHidden,
			&item.ServiceURL,
			&item.ServiceAddressSource,
			&item.ServiceHostValue,
			&item.ServiceScheme,
			&item.ServiceHost,
			&item.ServicePort,
			&item.ServicePath,
			&item.ServiceIcon,
			&item.ServiceDeviceID,
			&item.ServiceDeviceName,
			&item.ServiceHealthStatus,
			&item.DeviceID,
			&item.DeviceName,
			&item.ManualName,
			&item.ManualURL,
			&item.Description,
			&item.IconMode,
			&item.IconValue,
			&useDevicePrimary,
			&item.Scheme,
			&item.Host,
			&item.Port,
			&item.Path,
			&item.Position,
			&isFavorite,
			&item.FavoritePosition,
			&item.ClickCount,
			&lastOpenedAt,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		item.ServiceHidden = serviceHidden == 1
		item.UseDevicePrimaryAddress = useDevicePrimary == 1
		item.IsFavorite = isFavorite == 1
		item.LastOpenedAt = parseTime(lastOpenedAt)
		item.CreatedAt = parseTime(createdAt)
		item.UpdatedAt = parseTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) loadBookmarkTags(ctx context.Context) (map[string][]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT bt.bookmark_id, t.name
		FROM bookmark_tags bt
		JOIN tags t ON t.id = bt.tag_id
		ORDER BY t.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := map[string][]string{}
	for rows.Next() {
		var bookmarkID, tagName string
		if err := rows.Scan(&bookmarkID, &tagName); err != nil {
			return nil, err
		}
		items[bookmarkID] = append(items[bookmarkID], tagName)
	}
	return items, rows.Err()
}

func (s *Store) loadPrimaryDeviceAddresses(ctx context.Context) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT device_id, ip_address
		FROM (
			SELECT
				device_id,
				ip_address,
				ROW_NUMBER() OVER (
					PARTITION BY device_id
					ORDER BY is_primary DESC, last_seen_at DESC, first_seen_at DESC
				) AS row_num
			FROM device_addresses
		)
		WHERE row_num = 1
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := map[string]string{}
	for rows.Next() {
		var deviceID, ipAddress string
		if err := rows.Scan(&deviceID, &ipAddress); err != nil {
			return nil, err
		}
		items[deviceID] = ipAddress
	}
	return items, rows.Err()
}

func buildBookmark(record bookmarkRecord, tags []string, primaryAddresses map[string]string) domain.Bookmark {
	item := domain.Bookmark{
		ID:                      record.ID,
		Description:             record.Description,
		Tags:                    append([]string(nil), tags...),
		FolderID:                record.FolderID,
		FolderName:              record.FolderName,
		ServiceID:               record.ServiceID,
		ServiceName:             record.ServiceName,
		ServiceSource:           record.ServiceSource,
		ServiceSourceRef:        record.ServiceSourceRef,
		ServiceHidden:           record.ServiceHidden,
		IsFavorite:              record.IsFavorite,
		FavoritePosition:        record.FavoritePosition,
		Position:                record.Position,
		ClickCount:              record.ClickCount,
		LastOpenedAt:            record.LastOpenedAt,
		ManualName:              record.ManualName,
		ManualURL:               record.ManualURL,
		IconMode:                record.IconMode,
		IconValue:               record.IconValue,
		UseDevicePrimaryAddress: record.UseDevicePrimaryAddress,
		Scheme:                  record.Scheme,
		Host:                    record.Host,
		Port:                    record.Port,
		Path:                    record.Path,
		CreatedAt:               record.CreatedAt,
		UpdatedAt:               record.UpdatedAt,
	}
	if record.ServiceID != "" {
		item.Name = firstNonEmpty(record.ManualName, record.ServiceName)
		item.Scheme = record.ServiceScheme
		item.Port = record.ServicePort
		item.Path = record.ServicePath
		item.Host = resolveAddressSourceHost(record.ServiceAddressSource, record.ServiceHostValue, primaryAddresses[record.ServiceDeviceID])
		item.URL = buildServiceURL(item.Scheme, item.Host, item.Port, item.Path)
		if item.URL == "" {
			item.URL = record.ServiceURL
		}
		item.DeviceID = record.ServiceDeviceID
		item.DeviceName = record.ServiceDeviceName
		item.HealthStatus = record.ServiceHealthStatus
	} else {
		item.Name = record.ManualName
		item.DeviceID = record.DeviceID
		item.DeviceName = record.DeviceName
		item.HealthStatus = domain.HealthStatusUnknown
		item.URL = record.ManualURL
		if record.UseDevicePrimaryAddress {
			if ipAddress := primaryAddresses[record.DeviceID]; ipAddress != "" {
				item.URL = buildBookmarkURL(record.Scheme, ipAddress, record.Port, record.Path)
				item.Host = ipAddress
			}
		}
	}
	if item.Host == "" {
		item.Host = firstNonEmpty(record.Host, extractBookmarkHost(item.URL))
	}
	if item.IconMode == "" {
		item.IconMode = "auto"
	}
	item.Icon = resolveBookmarkIcon(item, record.ServiceIcon)
	if item.Icon == "" {
		item.Icon = faviconURL(item.URL)
	}
	return item
}

func bookmarkMatchesFilters(item domain.Bookmark, options domain.BookmarkListOptions) bool {
	if options.Favorites != nil && item.IsFavorite != *options.Favorites {
		return false
	}
	if options.FolderID != "" && item.FolderID != options.FolderID {
		return false
	}
	if options.DeviceID != "" && item.DeviceID != options.DeviceID {
		return false
	}
	if options.ServiceID != "" && item.ServiceID != options.ServiceID {
		return false
	}
	if options.Tag != "" {
		match := false
		for _, tag := range item.Tags {
			if strings.EqualFold(tag, options.Tag) || strings.EqualFold(slugify(tag), slugify(options.Tag)) {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}
	if options.Query == "" {
		return true
	}
	query := strings.ToLower(strings.TrimSpace(options.Query))
	fields := []string{item.Name, item.DeviceName, item.ServiceName}
	fields = append(fields, item.Tags...)
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), query) {
			return true
		}
	}
	return false
}

func sortBookmarks(items []domain.Bookmark) {
	slices.SortFunc(items, func(a, b domain.Bookmark) int {
		if a.IsFavorite != b.IsFavorite {
			if a.IsFavorite {
				return -1
			}
			return 1
		}
		if a.IsFavorite && b.IsFavorite && a.FavoritePosition != b.FavoritePosition {
			return a.FavoritePosition - b.FavoritePosition
		}
		if a.FolderID != b.FolderID {
			return strings.Compare(a.FolderID, b.FolderID)
		}
		if a.Position != b.Position {
			return a.Position - b.Position
		}
		return strings.Compare(a.Name, b.Name)
	})
}

func resolveBookmarkIcon(item domain.Bookmark, serviceIcon string) string {
	switch item.IconMode {
	case "external", "uploaded":
		return item.IconValue
	case "auto", "":
		if icon := knownServiceIcon(firstNonEmpty(item.Name, item.ServiceName, serviceIcon)); icon != "" {
			return icon
		}
		if icon := knownServiceIcon(serviceIcon); icon != "" {
			return icon
		}
		if looksLikeAbsoluteURL(serviceIcon) || strings.HasPrefix(serviceIcon, "data:") {
			return serviceIcon
		}
		return faviconURL(item.URL)
	default:
		return item.IconValue
	}
}

func knownServiceIcon(name string) string {
	key := slugify(name)
	if key == "" {
		return ""
	}
	meta := map[string][3]string{
		"grafana":             {"#f97316", "#ffffff", "G"},
		"prometheus":          {"#dc6b2f", "#ffffff", "P"},
		"home-assistant":      {"#2563eb", "#ffffff", "HA"},
		"plex":                {"#111827", "#fbbf24", "P"},
		"jellyfin":            {"#1e293b", "#a78bfa", "J"},
		"portainer":           {"#0f766e", "#ffffff", "Pt"},
		"traefik":             {"#1d4ed8", "#ffffff", "T"},
		"nextcloud":           {"#2563eb", "#ffffff", "N"},
		"homebridge":          {"#ef4444", "#ffffff", "HB"},
		"router":              {"#111827", "#22c55e", "R"},
		"nas":                 {"#334155", "#ffffff", "N"},
		"home-assistant-8123": {"#2563eb", "#ffffff", "HA"},
	}
	if item, ok := meta[key]; ok {
		return svgBadgeDataURI(item[0], item[1], item[2])
	}
	return ""
}

func svgBadgeDataURI(background, foreground, label string) string {
	content := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><rect width="64" height="64" rx="18" fill="%s"/><text x="32" y="37" text-anchor="middle" font-family="Arial, sans-serif" font-size="24" font-weight="700" fill="%s">%s</text></svg>`, background, foreground, label)
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(content))
}

func faviconURL(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return fmt.Sprintf("%s://%s/favicon.ico", parsed.Scheme, parsed.Host)
}

func looksLikeAbsoluteURL(value string) bool {
	parsed, err := url.Parse(strings.TrimSpace(value))
	return err == nil && parsed.Scheme != "" && parsed.Host != ""
}

func extractBookmarkHost(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}

func buildBookmarkURL(scheme, host string, port int, path string) string {
	if strings.TrimSpace(scheme) == "" {
		scheme = "http"
	}
	base := fmt.Sprintf("%s://%s", scheme, host)
	if port > 0 && port != 80 && port != 443 {
		base = fmt.Sprintf("%s:%d", base, port)
	}
	if path == "" || path == "/" {
		return base
	}
	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "?") {
		return base + path
	}
	return base + "/" + path
}

func parseBookmarkURL(rawURL string) (string, string, int, string, error) {
	if strings.TrimSpace(rawURL) == "" {
		return "", "", 0, "", nil
	}
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", "", 0, "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", "", 0, "", errors.New("bookmark url must be absolute")
	}
	port := 0
	if parsed.Port() != "" {
		fmt.Sscanf(parsed.Port(), "%d", &port)
	}
	path := parsed.EscapedPath()
	if parsed.RawQuery != "" {
		path += "?" + parsed.RawQuery
	}
	return parsed.Scheme, parsed.Hostname(), port, path, nil
}

func normalizeServiceInput(service domain.Service) domain.Service {
	service = domain.Service(service)
	if service.Details == nil {
		service.Details = map[string]any{}
	}
	service.Source = domain.ServiceSourceManual
	if service.URL != "" && (service.Host == "" || service.Scheme == "") {
		if parsed, err := url.Parse(service.URL); err == nil {
			service.Scheme = firstNonEmpty(service.Scheme, parsed.Scheme)
			service.Host = firstNonEmpty(service.Host, parsed.Hostname())
			service.Path = firstNonEmpty(service.Path, parsed.EscapedPath())
			if parsed.RawQuery != "" {
				service.Path = firstNonEmpty(service.Path, "?"+parsed.RawQuery)
			}
			if parsed.Port() != "" {
				fmt.Sscanf(parsed.Port(), "%d", &service.Port)
			}
		}
	}
	if service.URL == "" {
		service.URL = buildServiceURL(service.Scheme, service.Host, service.Port, service.Path)
	}
	return service
}

func serviceDisplayName(record bookmarkRecord) string {
	return firstNonEmpty(record.ServiceName, record.ManualName)
}

func (s *Store) lookupBookmarkTx(ctx context.Context, tx *sql.Tx, id string) (bookmarkRecord, bool, error) {
	if strings.TrimSpace(id) == "" {
		return bookmarkRecord{}, false, nil
	}
	row := tx.QueryRowContext(ctx, `
		SELECT
			id,
			COALESCE(folder_id, ''),
			COALESCE(service_id, ''),
			COALESCE(device_id, ''),
			COALESCE(manual_name, ''),
			COALESCE(manual_url, ''),
			COALESCE(description, ''),
			icon_mode,
			COALESCE(icon_value, ''),
			use_device_primary_address,
			COALESCE(scheme, ''),
			COALESCE(host, ''),
			port,
			COALESCE(path, ''),
			position,
			is_favorite,
			COALESCE(favorite_position, 0),
			click_count,
			COALESCE(last_opened_at, ''),
			created_at,
			updated_at
		FROM bookmarks
		WHERE id = ?
	`, id)
	var item bookmarkRecord
	var useDevicePrimary, isFavorite int
	var lastOpenedAt, createdAt, updatedAt string
	err := row.Scan(
		&item.ID,
		&item.FolderID,
		&item.ServiceID,
		&item.DeviceID,
		&item.ManualName,
		&item.ManualURL,
		&item.Description,
		&item.IconMode,
		&item.IconValue,
		&useDevicePrimary,
		&item.Scheme,
		&item.Host,
		&item.Port,
		&item.Path,
		&item.Position,
		&isFavorite,
		&item.FavoritePosition,
		&item.ClickCount,
		&lastOpenedAt,
		&createdAt,
		&updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return bookmarkRecord{}, false, nil
	}
	if err != nil {
		return bookmarkRecord{}, false, err
	}
	item.UseDevicePrimaryAddress = useDevicePrimary == 1
	item.IsFavorite = isFavorite == 1
	item.LastOpenedAt = parseTime(lastOpenedAt)
	item.CreatedAt = parseTime(createdAt)
	item.UpdatedAt = parseTime(updatedAt)
	return item, true, nil
}

func (s *Store) findBookmarkIDByServiceTx(ctx context.Context, tx *sql.Tx, serviceID string) (string, error) {
	var bookmarkID string
	err := tx.QueryRowContext(ctx, `SELECT id FROM bookmarks WHERE service_id = ?`, serviceID).Scan(&bookmarkID)
	return bookmarkID, err
}

func (s *Store) syncBookmarkTagsTx(ctx context.Context, tx *sql.Tx, bookmarkID string, tags []string, now time.Time) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM bookmark_tags WHERE bookmark_id = ?`, bookmarkID); err != nil {
		return err
	}
	for _, tag := range normalizeTags(tags) {
		tagID, err := s.ensureTagTx(ctx, tx, tag, now)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO bookmark_tags(bookmark_id, tag_id, created_at) VALUES(?, ?, ?)`, bookmarkID, tagID, now.Format(time.RFC3339Nano)); err != nil {
			return err
		}
	}
	_, err := tx.ExecContext(ctx, `
		DELETE FROM tags
		WHERE id IN (
			SELECT t.id FROM tags t
			LEFT JOIN bookmark_tags bt ON bt.tag_id = t.id
			GROUP BY t.id
			HAVING COUNT(bt.bookmark_id) = 0
		)
	`)
	return err
}

func normalizeTags(tags []string) []string {
	seen := map[string]struct{}{}
	items := make([]string, 0, len(tags))
	for _, tag := range tags {
		value := strings.TrimSpace(tag)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		items = append(items, value)
	}
	slices.Sort(items)
	return items
}

func (s *Store) ensureTagTx(ctx context.Context, tx *sql.Tx, name string, now time.Time) (string, error) {
	var tagID string
	err := tx.QueryRowContext(ctx, `SELECT id FROM tags WHERE name = ? COLLATE NOCASE`, name).Scan(&tagID)
	if err == nil {
		if _, updateErr := tx.ExecContext(ctx, `UPDATE tags SET slug = ?, updated_at = ? WHERE id = ?`, slugify(name), now.Format(time.RFC3339Nano), tagID); updateErr != nil {
			return "", updateErr
		}
		return tagID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	tagID = newID("tag")
	if _, err := tx.ExecContext(ctx, `INSERT INTO tags(id, name, slug, created_at, updated_at) VALUES(?, ?, ?, ?, ?)`, tagID, name, slugify(name), now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano)); err != nil {
		return "", err
	}
	return tagID, nil
}

func (s *Store) ensureFolderParentAllowedTx(ctx context.Context, tx *sql.Tx, folderID, parentID string) error {
	current := parentID
	for current != "" {
		if current == folderID {
			return errors.New("folder parent would create a cycle")
		}
		var next string
		err := tx.QueryRowContext(ctx, `SELECT COALESCE(parent_id, '') FROM folders WHERE id = ?`, current).Scan(&next)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		current = next
	}
	return nil
}

func (s *Store) getFolderTx(ctx context.Context, tx *sql.Tx, id string) (domain.Folder, bool, error) {
	row := tx.QueryRowContext(ctx, `SELECT id, COALESCE(parent_id, ''), name, slug, position, created_at, updated_at FROM folders WHERE id = ?`, id)
	var item domain.Folder
	var createdAt, updatedAt string
	err := row.Scan(&item.ID, &item.ParentID, &item.Name, &item.Slug, &item.Position, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Folder{}, false, nil
	}
	if err != nil {
		return domain.Folder{}, false, err
	}
	item.CreatedAt = parseTime(createdAt)
	item.UpdatedAt = parseTime(updatedAt)
	return item, true, nil
}

func (s *Store) deleteBookmarkOwnedServiceTx(ctx context.Context, tx *sql.Tx, serviceID, bookmarkID string) error {
	if strings.TrimSpace(serviceID) == "" {
		return nil
	}
	var hidden int
	var detailsJSON string
	err := tx.QueryRowContext(ctx, `SELECT hidden, details_json FROM services WHERE id = ?`, serviceID).Scan(&hidden, &detailsJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if hidden != 1 {
		return nil
	}
	var details map[string]any
	_ = json.Unmarshal([]byte(detailsJSON), &details)
	if details["bookmarkOwnerID"] != bookmarkID {
		return nil
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM services WHERE id = ?`, serviceID)
	return err
}

func nullableInt(value int) any {
	if value == 0 {
		return nil
	}
	return value
}

func bookmarkAssetFileName(assetURL string) string {
	return filepath.Base(strings.TrimSpace(assetURL))
}

func isIPAddress(value string) bool {
	return net.ParseIP(strings.TrimSpace(value)) != nil
}
