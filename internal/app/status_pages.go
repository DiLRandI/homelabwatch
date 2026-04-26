package app

import (
	"context"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
)

func (a *App) ListStatusPages(ctx context.Context) ([]domain.StatusPageListItem, error) {
	return a.store.ListStatusPages(ctx)
}

func (a *App) GetStatusPage(ctx context.Context, id string) (domain.StatusPage, error) {
	return a.store.GetStatusPage(ctx, id)
}

func (a *App) SaveStatusPage(ctx context.Context, input domain.StatusPageInput) (domain.StatusPage, error) {
	item, err := a.store.SaveStatusPage(ctx, input)
	if err != nil {
		return domain.StatusPage{}, err
	}
	a.publish("status-page", item.ID, "upserted", item)
	return item, nil
}

func (a *App) DeleteStatusPage(ctx context.Context, id string) error {
	if err := a.store.DeleteStatusPage(ctx, id); err != nil {
		return err
	}
	a.publish("status-page", id, "deleted", nil)
	return nil
}

func (a *App) ReplaceStatusPageServices(ctx context.Context, pageID string, services []domain.StatusPageServiceInput) (domain.StatusPage, error) {
	item, err := a.store.ReplaceStatusPageServices(ctx, pageID, services)
	if err != nil {
		return domain.StatusPage{}, err
	}
	a.publish("status-page", pageID, "services_updated", item)
	return item, nil
}

func (a *App) CreateStatusPageAnnouncement(ctx context.Context, pageID string, input domain.StatusPageAnnouncementInput) (domain.StatusPageAnnouncement, error) {
	item, err := a.store.CreateStatusPageAnnouncement(ctx, pageID, input)
	if err != nil {
		return domain.StatusPageAnnouncement{}, err
	}
	a.publish("status-page-announcement", item.ID, "upserted", item)
	return item, nil
}

func (a *App) UpdateStatusPageAnnouncement(ctx context.Context, id string, input domain.StatusPageAnnouncementInput) (domain.StatusPageAnnouncement, error) {
	item, err := a.store.UpdateStatusPageAnnouncement(ctx, id, input)
	if err != nil {
		return domain.StatusPageAnnouncement{}, err
	}
	a.publish("status-page-announcement", item.ID, "upserted", item)
	return item, nil
}

func (a *App) DeleteStatusPageAnnouncement(ctx context.Context, id string) error {
	if err := a.store.DeleteStatusPageAnnouncement(ctx, id); err != nil {
		return err
	}
	a.publish("status-page-announcement", id, "deleted", nil)
	return nil
}

func (a *App) GetPublicStatusPage(ctx context.Context, slug string, now time.Time) (domain.PublicStatusPage, error) {
	return a.store.GetPublicStatusPage(ctx, slug, now)
}
