package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/deleema/homelabwatch/internal/domain"
)

func (s *Store) ListStatusPages(ctx context.Context) ([]domain.StatusPageListItem, error) {
	rows, err := s.reader().QueryContext(ctx, `
		SELECT
			p.id, p.slug, p.title, p.description, p.enabled, p.created_at, p.updated_at,
			COUNT(DISTINCT sps.service_id) AS service_count,
			COUNT(DISTINCT a.id) AS announcement_count,
			COALESCE(GROUP_CONCAT(svc.status), '') AS statuses
		FROM status_pages p
		LEFT JOIN status_page_services sps ON sps.status_page_id = p.id
		LEFT JOIN services svc ON svc.id = sps.service_id
		LEFT JOIN status_page_announcements a ON a.status_page_id = p.id
		GROUP BY p.id
		ORDER BY p.title COLLATE NOCASE
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.StatusPageListItem{}
	for rows.Next() {
		var item domain.StatusPageListItem
		var enabled int
		var createdAt, updatedAt, statuses string
		if err := rows.Scan(&item.ID, &item.Slug, &item.Title, &item.Description, &enabled, &createdAt, &updatedAt, &item.ServiceCount, &item.AnnouncementCount, &statuses); err != nil {
			return nil, err
		}
		item.Enabled = enabled == 1
		item.CreatedAt = parseTime(createdAt)
		item.UpdatedAt = parseTime(updatedAt)
		item.OverallStatus = rollupStatusFromCSV(statuses, item.ServiceCount)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) GetStatusPage(ctx context.Context, id string) (domain.StatusPage, error) {
	page, err := s.getStatusPageBase(ctx, s.reader(), "id", id)
	if err != nil {
		return domain.StatusPage{}, err
	}
	services, err := s.listStatusPageServices(ctx, page.ID)
	if err != nil {
		return domain.StatusPage{}, err
	}
	announcements, err := s.listStatusPageAnnouncements(ctx, page.ID, false, time.Time{})
	if err != nil {
		return domain.StatusPage{}, err
	}
	page.Services = services
	page.Announcements = announcements
	return page, nil
}

func (s *Store) SaveStatusPage(ctx context.Context, input domain.StatusPageInput) (domain.StatusPage, error) {
	slug := normalizeStatusPageSlug(input.Slug)
	if slug == "" {
		return domain.StatusPage{}, errors.New("status page slug is required")
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return domain.StatusPage{}, errors.New("status page title is required")
	}
	now := time.Now().UTC()
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = newID("stp")
		if _, err := s.db.ExecContext(ctx, `
			INSERT INTO status_pages(id, slug, title, description, enabled, created_at, updated_at)
			VALUES(?, ?, ?, ?, ?, ?, ?)
		`, id, slug, title, strings.TrimSpace(input.Description), boolInt(enabled), now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano)); err != nil {
			return domain.StatusPage{}, err
		}
		return s.GetStatusPage(ctx, id)
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE status_pages
		SET slug = ?, title = ?, description = ?, enabled = ?, updated_at = ?
		WHERE id = ?
	`, slug, title, strings.TrimSpace(input.Description), boolInt(enabled), now.Format(time.RFC3339Nano), id)
	if err != nil {
		return domain.StatusPage{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return domain.StatusPage{}, err
	}
	if affected == 0 {
		return domain.StatusPage{}, sql.ErrNoRows
	}
	return s.GetStatusPage(ctx, id)
}

func (s *Store) DeleteStatusPage(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM status_pages WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) ReplaceStatusPageServices(ctx context.Context, pageID string, services []domain.StatusPageServiceInput) (domain.StatusPage, error) {
	if err := rejectDuplicateStatusPageServices(services); err != nil {
		return domain.StatusPage{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.StatusPage{}, err
	}
	defer tx.Rollback()
	if _, err := s.getStatusPageBase(ctx, tx, "id", pageID); err != nil {
		return domain.StatusPage{}, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM status_page_services WHERE status_page_id = ?`, pageID); err != nil {
		return domain.StatusPage{}, err
	}
	for index, item := range services {
		serviceID := strings.TrimSpace(item.ServiceID)
		if serviceID == "" {
			return domain.StatusPage{}, errors.New("service id is required")
		}
		sortOrder := index
		if item.SortOrder != nil {
			sortOrder = *item.SortOrder
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO status_page_services(status_page_id, service_id, sort_order, display_name)
			VALUES(?, ?, ?, ?)
		`, pageID, serviceID, sortOrder, strings.TrimSpace(item.DisplayName)); err != nil {
			return domain.StatusPage{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE status_pages SET updated_at = ? WHERE id = ?`, nowString(), pageID); err != nil {
		return domain.StatusPage{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.StatusPage{}, err
	}
	return s.GetStatusPage(ctx, pageID)
}

func (s *Store) CreateStatusPageAnnouncement(ctx context.Context, pageID string, input domain.StatusPageAnnouncementInput) (domain.StatusPageAnnouncement, error) {
	if _, err := s.getStatusPageBase(ctx, s.reader(), "id", pageID); err != nil {
		return domain.StatusPageAnnouncement{}, err
	}
	if err := validateStatusPageAnnouncement(input); err != nil {
		return domain.StatusPageAnnouncement{}, err
	}
	now := time.Now().UTC()
	item := domain.StatusPageAnnouncement{
		ID:           newID("sta"),
		StatusPageID: pageID,
		Kind:         input.Kind,
		Title:        strings.TrimSpace(input.Title),
		Message:      strings.TrimSpace(input.Message),
		StartsAt:     input.StartsAt,
		EndsAt:       input.EndsAt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO status_page_announcements(id, status_page_id, kind, title, message, starts_at, ends_at, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, item.ID, item.StatusPageID, item.Kind, item.Title, item.Message, nullableTime(item.StartsAt), nullableTime(item.EndsAt), item.CreatedAt.Format(time.RFC3339Nano), item.UpdatedAt.Format(time.RFC3339Nano)); err != nil {
		return domain.StatusPageAnnouncement{}, err
	}
	return item, nil
}

func (s *Store) UpdateStatusPageAnnouncement(ctx context.Context, id string, input domain.StatusPageAnnouncementInput) (domain.StatusPageAnnouncement, error) {
	if err := validateStatusPageAnnouncement(input); err != nil {
		return domain.StatusPageAnnouncement{}, err
	}
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx, `
		UPDATE status_page_announcements
		SET kind = ?, title = ?, message = ?, starts_at = ?, ends_at = ?, updated_at = ?
		WHERE id = ?
	`, input.Kind, strings.TrimSpace(input.Title), strings.TrimSpace(input.Message), nullableTime(input.StartsAt), nullableTime(input.EndsAt), now.Format(time.RFC3339Nano), id)
	if err != nil {
		return domain.StatusPageAnnouncement{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return domain.StatusPageAnnouncement{}, err
	}
	if affected == 0 {
		return domain.StatusPageAnnouncement{}, sql.ErrNoRows
	}
	return s.getStatusPageAnnouncement(ctx, id)
}

func (s *Store) DeleteStatusPageAnnouncement(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM status_page_announcements WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) GetPublicStatusPage(ctx context.Context, slug string, now time.Time) (domain.PublicStatusPage, error) {
	page, err := s.getStatusPageBase(ctx, s.reader(), "slug", normalizeStatusPageSlug(slug))
	if err != nil {
		return domain.PublicStatusPage{}, err
	}
	if !page.Enabled {
		return domain.PublicStatusPage{}, sql.ErrNoRows
	}
	services, err := s.listStatusPageServices(ctx, page.ID)
	if err != nil {
		return domain.PublicStatusPage{}, err
	}
	announcements, err := s.listStatusPageAnnouncements(ctx, page.ID, true, now)
	if err != nil {
		return domain.PublicStatusPage{}, err
	}
	public := domain.PublicStatusPage{
		Slug:          page.Slug,
		Title:         page.Title,
		Description:   page.Description,
		OverallStatus: rollupStatusFromServices(services),
		LastUpdatedAt: page.UpdatedAt,
		Services:      make([]domain.PublicStatusPageService, 0, len(services)),
		Announcements: make([]domain.PublicStatusPageAnnouncement, 0, len(announcements)),
	}
	for _, service := range services {
		public.LastUpdatedAt = maxTime(public.LastUpdatedAt, service.LastCheckedAt)
		item := domain.PublicStatusPageService{
			Name:          firstNonEmpty(strings.TrimSpace(service.DisplayName), service.ServiceName),
			Status:        service.Status,
			LastCheckedAt: service.LastCheckedAt,
		}
		if service.LatestCheck != nil {
			item.LatestCheck = &domain.PublicCheckSummary{
				Status:         service.LatestCheck.Status,
				LatencyMS:      service.LatestCheck.LatencyMS,
				HTTPStatusCode: service.LatestCheck.HTTPStatusCode,
				CheckedAt:      service.LatestCheck.CheckedAt,
				Message:        sanitizePublicCheckMessage(service.LatestCheck.Status),
			}
			public.LastUpdatedAt = maxTime(public.LastUpdatedAt, service.LatestCheck.CheckedAt)
		}
		public.Services = append(public.Services, item)
	}
	for _, announcement := range announcements {
		public.Announcements = append(public.Announcements, domain.PublicStatusPageAnnouncement{
			Kind:     announcement.Kind,
			Title:    announcement.Title,
			Message:  announcement.Message,
			StartsAt: announcement.StartsAt,
			EndsAt:   announcement.EndsAt,
		})
	}
	return public, nil
}

func (s *Store) getStatusPageBase(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, column, value string) (domain.StatusPage, error) {
	if column != "id" && column != "slug" {
		return domain.StatusPage{}, fmt.Errorf("unsupported status page lookup %q", column)
	}
	var item domain.StatusPage
	var enabled int
	var createdAt, updatedAt string
	if err := queryer.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT id, slug, title, description, enabled, created_at, updated_at
		FROM status_pages
		WHERE %s = ?
	`, column), value).Scan(&item.ID, &item.Slug, &item.Title, &item.Description, &enabled, &createdAt, &updatedAt); err != nil {
		return domain.StatusPage{}, err
	}
	item.Enabled = enabled == 1
	item.CreatedAt = parseTime(createdAt)
	item.UpdatedAt = parseTime(updatedAt)
	return item, nil
}

func (s *Store) listStatusPageServices(ctx context.Context, pageID string) ([]domain.StatusPageService, error) {
	rows, err := s.reader().QueryContext(ctx, `
		SELECT
			sps.status_page_id, sps.service_id, sps.sort_order, sps.display_name,
			svc.name, svc.status, COALESCE(svc.last_checked_at, ''),
			COALESCE(r.id, ''), COALESCE(r.health_check_id, ''), COALESCE(r.status, ''),
			COALESCE(r.latency_ms, 0), COALESCE(r.http_status_code, 0),
			COALESCE(r.response_size_bytes, 0), COALESCE(r.message, ''), COALESCE(r.checked_at, '')
		FROM status_page_services sps
		JOIN services svc ON svc.id = sps.service_id
		LEFT JOIN (
			SELECT hcr1.*
			FROM health_check_results hcr1
			JOIN health_checks hc ON hc.id = hcr1.health_check_id AND hc.enabled = 1 AND hc.subject_type = 'service'
			JOIN (
				SELECT hcr.subject_id, MAX(hcr.checked_at) AS checked_at
				FROM health_check_results hcr
				JOIN health_checks h ON h.id = hcr.health_check_id
				WHERE h.enabled = 1 AND h.subject_type = 'service'
				GROUP BY hcr.subject_id
			) latest ON latest.subject_id = hcr1.subject_id AND latest.checked_at = hcr1.checked_at
		) r ON r.subject_id = svc.id
		WHERE sps.status_page_id = ?
		ORDER BY sps.sort_order, svc.name COLLATE NOCASE
	`, pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.StatusPageService{}
	for rows.Next() {
		var item domain.StatusPageService
		var lastCheckedAt, resultID, resultCheckID, resultStatus, resultMessage, resultCheckedAt string
		var resultLatency, resultResponseSize int64
		var resultHTTPStatus int
		if err := rows.Scan(&item.StatusPageID, &item.ServiceID, &item.SortOrder, &item.DisplayName, &item.ServiceName, &item.Status, &lastCheckedAt, &resultID, &resultCheckID, &resultStatus, &resultLatency, &resultHTTPStatus, &resultResponseSize, &resultMessage, &resultCheckedAt); err != nil {
			return nil, err
		}
		item.LastCheckedAt = parseTime(lastCheckedAt)
		if resultID != "" {
			item.LatestCheck = &domain.CheckResult{
				ID:                resultID,
				CheckID:           resultCheckID,
				ServiceID:         item.ServiceID,
				SubjectType:       domain.HealthCheckSubjectService,
				SubjectID:         item.ServiceID,
				Status:            domain.HealthStatus(resultStatus),
				LatencyMS:         resultLatency,
				HTTPStatusCode:    resultHTTPStatus,
				ResponseSizeBytes: resultResponseSize,
				Message:           resultMessage,
				CheckedAt:         parseTime(resultCheckedAt),
			}
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) listStatusPageAnnouncements(ctx context.Context, pageID string, activeOnly bool, now time.Time) ([]domain.StatusPageAnnouncement, error) {
	query := `
		SELECT id, status_page_id, kind, title, message, COALESCE(starts_at, ''), COALESCE(ends_at, ''), created_at, updated_at
		FROM status_page_announcements
		WHERE status_page_id = ?
	`
	args := []any{pageID}
	if activeOnly {
		nowText := now.UTC().Format(time.RFC3339Nano)
		query += ` AND (starts_at IS NULL OR starts_at <= ?) AND (ends_at IS NULL OR ends_at >= ?)`
		args = append(args, nowText, nowText)
	}
	query += ` ORDER BY COALESCE(starts_at, created_at) DESC, created_at DESC`
	rows, err := s.reader().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.StatusPageAnnouncement{}
	for rows.Next() {
		item, err := scanStatusPageAnnouncement(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) getStatusPageAnnouncement(ctx context.Context, id string) (domain.StatusPageAnnouncement, error) {
	row := s.reader().QueryRowContext(ctx, `
		SELECT id, status_page_id, kind, title, message, COALESCE(starts_at, ''), COALESCE(ends_at, ''), created_at, updated_at
		FROM status_page_announcements
		WHERE id = ?
	`, id)
	return scanStatusPageAnnouncement(row)
}

func scanStatusPageAnnouncement(scanner interface{ Scan(dest ...any) error }) (domain.StatusPageAnnouncement, error) {
	var item domain.StatusPageAnnouncement
	var startsAt, endsAt, createdAt, updatedAt string
	if err := scanner.Scan(&item.ID, &item.StatusPageID, &item.Kind, &item.Title, &item.Message, &startsAt, &endsAt, &createdAt, &updatedAt); err != nil {
		return domain.StatusPageAnnouncement{}, err
	}
	item.StartsAt = parseTime(startsAt)
	item.EndsAt = parseTime(endsAt)
	item.CreatedAt = parseTime(createdAt)
	item.UpdatedAt = parseTime(updatedAt)
	return item, nil
}

func normalizeStatusPageSlug(input string) string {
	var builder strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(input)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
			lastDash = false
			continue
		}
		if (r == '-' || r == '_' || unicode.IsSpace(r)) && !lastDash && builder.Len() > 0 {
			builder.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}

func rejectDuplicateStatusPageServices(services []domain.StatusPageServiceInput) error {
	seen := map[string]bool{}
	for _, service := range services {
		id := strings.TrimSpace(service.ServiceID)
		if id == "" {
			continue
		}
		if seen[id] {
			return fmt.Errorf("duplicate service assignment %q", id)
		}
		seen[id] = true
	}
	return nil
}

func validateStatusPageAnnouncement(input domain.StatusPageAnnouncementInput) error {
	switch input.Kind {
	case domain.StatusPageAnnouncementInfo, domain.StatusPageAnnouncementMaintenance, domain.StatusPageAnnouncementIncident, domain.StatusPageAnnouncementResolved:
	default:
		return errors.New("invalid announcement kind")
	}
	if strings.TrimSpace(input.Title) == "" {
		return errors.New("announcement title is required")
	}
	if strings.TrimSpace(input.Message) == "" {
		return errors.New("announcement message is required")
	}
	if !input.StartsAt.IsZero() && !input.EndsAt.IsZero() && input.EndsAt.Before(input.StartsAt) {
		return errors.New("announcement end must be after start")
	}
	return nil
}

func rollupStatusFromCSV(statuses string, count int) domain.HealthStatus {
	if count == 0 {
		return domain.HealthStatusUnknown
	}
	parts := strings.Split(statuses, ",")
	return rollupStatus(parts)
}

func rollupStatusFromServices(services []domain.StatusPageService) domain.HealthStatus {
	statuses := make([]string, 0, len(services))
	for _, service := range services {
		statuses = append(statuses, string(service.Status))
	}
	return rollupStatus(statuses)
}

func rollupStatus(statuses []string) domain.HealthStatus {
	if len(statuses) == 0 {
		return domain.HealthStatusUnknown
	}
	healthy := 0
	unhealthy := 0
	for _, status := range statuses {
		switch domain.HealthStatus(strings.TrimSpace(status)) {
		case domain.HealthStatusHealthy:
			healthy++
		case domain.HealthStatusUnhealthy:
			unhealthy++
		default:
			return domain.HealthStatusDegraded
		}
	}
	if healthy == len(statuses) {
		return domain.HealthStatusHealthy
	}
	if unhealthy == len(statuses) {
		return domain.HealthStatusUnhealthy
	}
	return domain.HealthStatusDegraded
}

func sanitizePublicCheckMessage(status domain.HealthStatus) string {
	switch status {
	case domain.HealthStatusHealthy:
		return "Check passed."
	case domain.HealthStatusUnhealthy:
		return "Check failed."
	case domain.HealthStatusDegraded:
		return "Check degraded."
	default:
		return "Check status unknown."
	}
}

func maxTime(left, right time.Time) time.Time {
	if right.After(left) {
		return right
	}
	return left
}
