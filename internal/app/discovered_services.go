package app

import (
	"context"
	"slices"

	"github.com/deleema/homelabwatch/internal/domain"
	"github.com/deleema/homelabwatch/internal/monitoring"
)

func (a *App) ListDiscoveredServices(ctx context.Context) ([]domain.DiscoveredService, error) {
	return a.store.ListDiscoveredServices(ctx)
}

func (a *App) GetDiscoveredService(ctx context.Context, id string) (domain.DiscoveredService, error) {
	return a.store.GetDiscoveredService(ctx, id)
}

func (a *App) SaveDiscoverySettings(ctx context.Context, input domain.DiscoverySettings) (domain.DiscoverySettings, error) {
	return a.store.SaveDiscoverySettings(ctx, input)
}

func (a *App) IgnoreDiscoveredService(ctx context.Context, id string) (domain.DiscoveredService, error) {
	item, err := a.store.IgnoreDiscoveredService(ctx, id)
	if err != nil {
		return domain.DiscoveredService{}, err
	}
	a.publish("discovered-service", item.ID, "ignored", item)
	return item, nil
}

func (a *App) RestoreDiscoveredService(ctx context.Context, id string) (domain.DiscoveredService, error) {
	item, err := a.store.RestoreDiscoveredService(ctx, id)
	if err != nil {
		return domain.DiscoveredService{}, err
	}
	a.publish("discovered-service", item.ID, "restored", item)
	return item, nil
}

func (a *App) CreateBookmarkFromDiscoveredService(ctx context.Context, input domain.CreateBookmarkFromDiscoveredServiceInput) (domain.Bookmark, error) {
	item, err := a.store.GetDiscoveredService(ctx, input.DiscoveredServiceID)
	if err != nil {
		return domain.Bookmark{}, err
	}
	if item.AcceptedBookmarkID != "" {
		return a.store.GetBookmark(ctx, item.AcceptedBookmarkID)
	}

	serviceID := item.AcceptedServiceID
	createdService := false
	if serviceID == "" {
		service, err := a.SaveManualService(ctx, domain.Service{
			Name:                      firstNonEmpty(input.Name, item.Name),
			Source:                    preferredDiscoverySource(item.SourceTypes),
			SourceRef:                 "accepted:" + item.ID,
			OriginDiscoveredServiceID: item.ID,
			ServiceType:               item.ServiceType,
			AddressSource:             item.AddressSource,
			HostValue:                 item.HostValue,
			DeviceID:                  item.DeviceID,
			Icon:                      item.Icon,
			Scheme:                    item.Scheme,
			Host:                      item.Host,
			Port:                      item.Port,
			Path:                      item.Path,
			URL:                       item.URL,
			Status:                    item.Status,
			LastSeenAt:                item.LastSeenAt,
			Details: map[string]any{
				"discoveredServiceID": item.ID,
				"sourceTypes":         item.SourceTypes,
			},
		})
		if err != nil {
			return domain.Bookmark{}, err
		}
		serviceID = service.ID
		createdService = true
		if err := a.store.CopyDiscoveredChecksToService(ctx, item.ID, serviceID); err != nil {
			_ = a.store.DeleteService(ctx, serviceID)
			return domain.Bookmark{}, err
		}
	}

	bookmark, err := a.store.CreateBookmarkFromService(ctx, domain.CreateBookmarkFromServiceInput{
		ServiceID:        serviceID,
		FolderID:         input.FolderID,
		Tags:             input.Tags,
		Name:             input.Name,
		IconMode:         input.IconMode,
		IconValue:        input.IconValue,
		IsFavorite:       input.IsFavorite,
		FavoritePosition: input.FavoritePosition,
	})
	if err != nil {
		if createdService && serviceID != "" {
			_ = a.store.DeleteService(ctx, serviceID)
		}
		return domain.Bookmark{}, err
	}
	if err := a.store.SaveDiscoveredServiceBookmarkLink(ctx, bookmark.ID, item.ID); err != nil {
		return domain.Bookmark{}, err
	}
	discovered, err := a.store.MarkDiscoveredServiceAccepted(ctx, item.ID, serviceID, bookmark.ID)
	if err != nil {
		return domain.Bookmark{}, err
	}
	a.publish("discovered-service", discovered.ID, "accepted", discovered)
	a.publish("bookmark", bookmark.ID, "upserted", bookmark)
	return bookmark, nil
}

func (a *App) runDiscoveredMonitoring(ctx context.Context) (int, error) {
	checks, err := a.store.GetDiscoveredChecksDue(ctx)
	if err != nil {
		return 0, err
	}
	for _, check := range checks {
		result := monitoring.RunAdhocCheck(ctx, check)
		outcome, err := a.store.SaveCheckResultWithOutcome(ctx, result)
		if err != nil {
			return 0, err
		}
		a.publishCheckOutcome(outcome)
	}
	return len(checks), nil
}

func (a *App) shouldAutoBookmark(item domain.DiscoveredService, settings domain.DiscoverySettings) bool {
	if item.State != domain.DiscoveryStatePending {
		return false
	}
	if settings.BookmarkPolicy != domain.BookmarkAutomationAutoHighConfidence {
		return false
	}
	if item.ConfidenceScore < settings.AutoBookmarkMinConfidence {
		return false
	}
	for _, source := range item.SourceTypes {
		if slices.Contains(settings.AutoBookmarkSources, source) {
			return true
		}
	}
	return false
}

func preferredDiscoverySource(sources []domain.ServiceSource) domain.ServiceSource {
	for _, candidate := range []domain.ServiceSource{
		domain.ServiceSourceDocker,
		domain.ServiceSourceMDNS,
		domain.ServiceSourceLAN,
	} {
		for _, source := range sources {
			if source == candidate {
				return source
			}
		}
	}
	return domain.ServiceSourceLAN
}
