package app

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

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
		if item.AcceptedServiceID == "" && serviceID != "" {
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

func (a *App) runDiscoveredMonitoring(ctx context.Context) error {
	items, err := a.store.ListDiscoveredServicesDueForHealth(ctx, time.Minute)
	if err != nil {
		return err
	}
	for _, item := range items {
		check := domain.ServiceCheck{
			ID:                "discovered-" + item.ID,
			ServiceID:         item.ID,
			TimeoutSeconds:    5,
			ExpectedStatusMin: 200,
			ExpectedStatusMax: 399,
			Enabled:           true,
		}
		switch {
		case strings.HasPrefix(item.URL, "http://"), strings.HasPrefix(item.URL, "https://"):
			check.Type = domain.CheckTypeHTTP
			check.Target = item.URL
		case item.Host != "" && item.Port > 0:
			check.Type = domain.CheckTypeTCP
			check.Target = fmt.Sprintf("%s:%d", item.Host, item.Port)
		default:
			check.Type = domain.CheckTypePing
			check.Target = item.Host
		}
		result := monitoring.RunAdhocCheck(ctx, check)
		if err := a.store.UpdateDiscoveredServiceHealth(ctx, item.ID, result.Status, result.CheckedAt); err != nil {
			return err
		}
	}
	return nil
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
