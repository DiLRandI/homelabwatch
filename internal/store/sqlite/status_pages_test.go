package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
)

func TestStatusPagesCreateAssignPublicReadAndSanitize(t *testing.T) {
	store := newBootstrappedStore(t)
	ctx := context.Background()

	service := saveStatusPageTestService(t, store, "NAS", domain.HealthStatusHealthy)
	checks, err := store.ListServiceChecks(ctx, service.ID)
	if err != nil || len(checks) == 0 {
		t.Fatalf("list checks: %v len=%d", err, len(checks))
	}
	_, err = store.SaveCheckResultWithOutcome(ctx, domain.CheckResult{
		CheckID:        checks[0].ID,
		Status:         domain.HealthStatusHealthy,
		LatencyMS:      42,
		HTTPStatusCode: 204,
		Message:        "dial 192.168.1.10: private raw error",
		CheckedAt:      time.Date(2026, 4, 26, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("save result: %v", err)
	}

	page, err := store.SaveStatusPage(ctx, domain.StatusPageInput{Slug: "Home Lab", Title: "Home Lab"})
	if err != nil {
		t.Fatalf("save page: %v", err)
	}
	page, err = store.ReplaceStatusPageServices(ctx, page.ID, []domain.StatusPageServiceInput{
		{ServiceID: service.ID, DisplayName: "Storage"},
	})
	if err != nil {
		t.Fatalf("replace services: %v", err)
	}
	if page.Services[0].DisplayName != "Storage" || page.Services[0].SortOrder != 0 {
		t.Fatalf("service assignment = %+v", page.Services[0])
	}

	public, err := store.GetPublicStatusPage(ctx, "home-lab", time.Date(2026, 4, 26, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("public read: %v", err)
	}
	if public.Services[0].Name != "Storage" {
		t.Fatalf("public name = %q", public.Services[0].Name)
	}
	if public.Services[0].LatestCheck.Message == "dial 192.168.1.10: private raw error" {
		t.Fatalf("public check message exposed raw error")
	}
}

func TestStatusPagesValidationCascadeAnnouncementsAndRollup(t *testing.T) {
	store := newBootstrappedStore(t)
	ctx := context.Background()
	healthy := saveStatusPageTestService(t, store, "Healthy", domain.HealthStatusHealthy)
	unhealthy := saveStatusPageTestService(t, store, "Unhealthy", domain.HealthStatusUnhealthy)

	page, err := store.SaveStatusPage(ctx, domain.StatusPageInput{Slug: "ops", Title: "Ops"})
	if err != nil {
		t.Fatalf("save page: %v", err)
	}
	if _, err := store.SaveStatusPage(ctx, domain.StatusPageInput{Slug: "OPS", Title: "Duplicate"}); err == nil {
		t.Fatalf("expected duplicate slug to fail")
	}
	if _, err := store.ReplaceStatusPageServices(ctx, page.ID, []domain.StatusPageServiceInput{{ServiceID: healthy.ID}, {ServiceID: healthy.ID}}); err == nil {
		t.Fatalf("expected duplicate service assignment to fail")
	}
	if _, err := store.ReplaceStatusPageServices(ctx, page.ID, []domain.StatusPageServiceInput{{ServiceID: healthy.ID}, {ServiceID: unhealthy.ID}}); err != nil {
		t.Fatalf("replace services: %v", err)
	}
	items, err := store.ListStatusPages(ctx)
	if err != nil {
		t.Fatalf("list pages: %v", err)
	}
	if items[0].OverallStatus != domain.HealthStatusDegraded {
		t.Fatalf("mixed rollup = %s", items[0].OverallStatus)
	}

	now := time.Date(2026, 4, 26, 10, 0, 0, 0, time.UTC)
	if _, err := store.CreateStatusPageAnnouncement(ctx, page.ID, domain.StatusPageAnnouncementInput{
		Kind: domain.StatusPageAnnouncementIncident, Title: "Active", Message: "Investigating", StartsAt: now.Add(-time.Hour), EndsAt: now.Add(time.Hour),
	}); err != nil {
		t.Fatalf("create active announcement: %v", err)
	}
	if _, err := store.CreateStatusPageAnnouncement(ctx, page.ID, domain.StatusPageAnnouncementInput{
		Kind: domain.StatusPageAnnouncementInfo, Title: "Expired", Message: "Done", StartsAt: now.Add(-2 * time.Hour), EndsAt: now.Add(-time.Hour),
	}); err != nil {
		t.Fatalf("create expired announcement: %v", err)
	}
	public, err := store.GetPublicStatusPage(ctx, "ops", now)
	if err != nil {
		t.Fatalf("public page: %v", err)
	}
	if len(public.Announcements) != 1 || public.Announcements[0].Title != "Active" {
		t.Fatalf("active announcements = %+v", public.Announcements)
	}

	disabled := false
	if _, err := store.SaveStatusPage(ctx, domain.StatusPageInput{ID: page.ID, Slug: "ops", Title: "Ops", Enabled: &disabled}); err != nil {
		t.Fatalf("disable page: %v", err)
	}
	if _, err := store.GetPublicStatusPage(ctx, "ops", now); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("disabled public read err = %v", err)
	}
}

func saveStatusPageTestService(t *testing.T, store *Store, name string, status domain.HealthStatus) domain.Service {
	t.Helper()
	service, err := store.SaveManualService(context.Background(), domain.Service{
		Name:   name,
		Scheme: "http",
		Host:   "127.0.0.1",
		Port:   8080,
		Status: status,
	})
	if err != nil {
		t.Fatalf("save service: %v", err)
	}
	return service
}
