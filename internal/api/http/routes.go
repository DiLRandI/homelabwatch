package httpapi

import (
	"net/http"

	"github.com/deleema/homelabwatch/internal/domain"
)

type routeSpec struct {
	handler http.HandlerFunc
	method  string
	path    string
	scope   domain.TokenScope
}

func routePattern(method, path string) string {
	return method + " " + path
}

func routePatternWithPrefix(method, prefix, path string) string {
	return routePattern(method, prefix+path)
}

func (r *Router) registerBootstrapRoutes(mux *http.ServeMux) {
	mux.HandleFunc(routePattern(http.MethodGet, "/api/v1/bootstrap/status"), r.handleBootstrapStatus)
	mux.HandleFunc(routePattern(http.MethodGet, "/api/ui/v1/bootstrap"), r.handleUIBootstrap)
	mux.Handle(routePattern(http.MethodPost, "/api/ui/v1/setup"), r.withTrustedConsole(http.HandlerFunc(r.handleSetup)))
}

func (r *Router) registerUIRoutes(mux *http.ServeMux) {
	openRoutes := []routeSpec{
		{method: http.MethodGet, path: "/api/ui/v1/dashboard", handler: r.handleDashboard},
		{method: http.MethodGet, path: "/api/ui/v1/status-pages", handler: r.handleStatusPages},
		{method: http.MethodGet, path: "/api/ui/v1/status-pages/{id}", handler: r.handleStatusPageByID},
		{method: http.MethodGet, path: "/api/ui/v1/settings", handler: r.handleSettings},
		{method: http.MethodGet, path: "/api/ui/v1/services", handler: r.handleServices},
		{method: http.MethodGet, path: "/api/ui/v1/services/{id}", handler: r.handleServiceByID},
		{method: http.MethodGet, path: "/api/ui/v1/services/{id}/events", handler: r.handleServiceEvents},
		{method: http.MethodGet, path: "/api/ui/v1/services/{id}/checks", handler: r.handleServiceChecks},
		{method: http.MethodGet, path: "/api/ui/v1/devices", handler: r.handleDevices},
		{method: http.MethodGet, path: "/api/ui/v1/devices/{id}", handler: r.handleDeviceByID},
		{method: http.MethodGet, path: "/api/ui/v1/bookmark-assets/{name}", handler: r.handleBookmarkAssetByName},
		{method: http.MethodGet, path: "/api/ui/v1/bookmarks", handler: r.handleBookmarks},
		{method: http.MethodGet, path: "/api/ui/v1/bookmarks/export", handler: r.handleBookmarkExport},
		{method: http.MethodGet, path: "/api/ui/v1/bookmarks/{id}/open", handler: r.handleBookmarkOpen},
		{method: http.MethodGet, path: "/api/ui/v1/folders", handler: r.handleFolders},
		{method: http.MethodGet, path: "/api/ui/v1/tags", handler: r.handleTags},
		{method: http.MethodGet, path: "/api/ui/v1/discovery/docker-endpoints", handler: r.handleDockerEndpoints},
		{method: http.MethodGet, path: "/api/ui/v1/discovery/scan-targets", handler: r.handleScanTargets},
		{method: http.MethodGet, path: "/api/ui/v1/discovered-services", handler: r.handleDiscoveredServices},
		{method: http.MethodGet, path: "/api/ui/v1/service-definitions", handler: r.handleServiceDefinitions},
		{method: http.MethodGet, path: "/api/ui/v1/notifications/channels", handler: r.handleNotificationChannels},
		{method: http.MethodGet, path: "/api/ui/v1/notifications/rules", handler: r.handleNotificationRules},
		{method: http.MethodGet, path: "/api/ui/v1/notifications/deliveries", handler: r.handleNotificationDeliveries},
	}

	trustedRoutes := []routeSpec{
		{method: http.MethodPost, path: "/api/ui/v1/status-pages", handler: r.handleStatusPages},
		{method: http.MethodPatch, path: "/api/ui/v1/status-pages/{id}", handler: r.handleStatusPageByID},
		{method: http.MethodDelete, path: "/api/ui/v1/status-pages/{id}", handler: r.handleStatusPageByID},
		{method: http.MethodPut, path: "/api/ui/v1/status-pages/{id}/services", handler: r.handleStatusPageServices},
		{method: http.MethodPost, path: "/api/ui/v1/status-pages/{id}/announcements", handler: r.handleStatusPageAnnouncements},
		{method: http.MethodPatch, path: "/api/ui/v1/status-page-announcements/{id}", handler: r.handleStatusPageAnnouncementByID},
		{method: http.MethodDelete, path: "/api/ui/v1/status-page-announcements/{id}", handler: r.handleStatusPageAnnouncementByID},
		{method: http.MethodPost, path: "/api/ui/v1/settings/api-tokens", handler: r.handleAPITokens},
		{method: http.MethodDelete, path: "/api/ui/v1/settings/api-tokens/{id}", handler: r.handleAPITokenByID},
		{method: http.MethodPost, path: "/api/ui/v1/services", handler: r.handleServices},
		{method: http.MethodPatch, path: "/api/ui/v1/services/{id}", handler: r.handleServiceByID},
		{method: http.MethodDelete, path: "/api/ui/v1/services/{id}", handler: r.handleServiceByID},
		{method: http.MethodPost, path: "/api/ui/v1/services/{id}/checks", handler: r.handleServiceChecks},
		{method: http.MethodPost, path: "/api/ui/v1/services/{id}/checks/test", handler: r.handleServiceCheckTest},
		{method: http.MethodPatch, path: "/api/ui/v1/checks/{id}", handler: r.handleCheckByID},
		{method: http.MethodDelete, path: "/api/ui/v1/checks/{id}", handler: r.handleCheckByID},
		{method: http.MethodPatch, path: "/api/ui/v1/devices/{id}", handler: r.handleDeviceByID},
		{method: http.MethodPost, path: "/api/ui/v1/bookmark-assets", handler: r.handleBookmarkAssets},
		{method: http.MethodPost, path: "/api/ui/v1/bookmarks", handler: r.handleBookmarks},
		{method: http.MethodPost, path: "/api/ui/v1/bookmarks/from-service", handler: r.handleBookmarkFromService},
		{method: http.MethodPost, path: "/api/ui/v1/bookmarks/reorder", handler: r.handleBookmarkReorder},
		{method: http.MethodPost, path: "/api/ui/v1/bookmarks/import", handler: r.handleBookmarkImport},
		{method: http.MethodPut, path: "/api/ui/v1/bookmarks/{id}", handler: r.handleBookmarkByID},
		{method: http.MethodPatch, path: "/api/ui/v1/bookmarks/{id}", handler: r.handleBookmarkByID},
		{method: http.MethodDelete, path: "/api/ui/v1/bookmarks/{id}", handler: r.handleBookmarkByID},
		{method: http.MethodPost, path: "/api/ui/v1/folders", handler: r.handleFolders},
		{method: http.MethodPost, path: "/api/ui/v1/folders/reorder", handler: r.handleFolderReorder},
		{method: http.MethodPut, path: "/api/ui/v1/folders/{id}", handler: r.handleFolderByID},
		{method: http.MethodDelete, path: "/api/ui/v1/folders/{id}", handler: r.handleFolderByID},
		{method: http.MethodPost, path: "/api/ui/v1/discovery/docker-endpoints", handler: r.handleDockerEndpoints},
		{method: http.MethodPatch, path: "/api/ui/v1/discovery/docker-endpoints/{id}", handler: r.handleDockerEndpointByID},
		{method: http.MethodDelete, path: "/api/ui/v1/discovery/docker-endpoints/{id}", handler: r.handleDockerEndpointByID},
		{method: http.MethodPost, path: "/api/ui/v1/discovery/scan-targets", handler: r.handleScanTargets},
		{method: http.MethodPatch, path: "/api/ui/v1/discovery/scan-targets/{id}", handler: r.handleScanTargetByID},
		{method: http.MethodDelete, path: "/api/ui/v1/discovery/scan-targets/{id}", handler: r.handleScanTargetByID},
		{method: http.MethodPost, path: "/api/ui/v1/discovered-services/{id}/bookmark", handler: r.handleDiscoveredServiceBookmark},
		{method: http.MethodPost, path: "/api/ui/v1/discovered-services/{id}/ignore", handler: r.handleDiscoveredServiceIgnore},
		{method: http.MethodPost, path: "/api/ui/v1/discovered-services/{id}/restore", handler: r.handleDiscoveredServiceRestore},
		{method: http.MethodPatch, path: "/api/ui/v1/discovery/settings", handler: r.handleDiscoverySettings},
		{method: http.MethodPost, path: "/api/ui/v1/discovery/run", handler: r.handleDiscoveryRun},
		{method: http.MethodPost, path: "/api/ui/v1/monitoring/run", handler: r.handleMonitoringRun},
		{method: http.MethodPost, path: "/api/ui/v1/service-definitions", handler: r.handleServiceDefinitions},
		{method: http.MethodPatch, path: "/api/ui/v1/service-definitions/{id}", handler: r.handleServiceDefinitionByID},
		{method: http.MethodDelete, path: "/api/ui/v1/service-definitions/{id}", handler: r.handleServiceDefinitionByID},
		{method: http.MethodPost, path: "/api/ui/v1/service-definitions/{id}/reapply", handler: r.handleServiceDefinitionReapply},
		{method: http.MethodPost, path: "/api/ui/v1/notifications/channels", handler: r.handleNotificationChannels},
		{method: http.MethodPatch, path: "/api/ui/v1/notifications/channels/{id}", handler: r.handleNotificationChannelByID},
		{method: http.MethodDelete, path: "/api/ui/v1/notifications/channels/{id}", handler: r.handleNotificationChannelByID},
		{method: http.MethodPost, path: "/api/ui/v1/notifications/channels/{id}/test", handler: r.handleNotificationChannelTest},
		{method: http.MethodPost, path: "/api/ui/v1/notifications/rules", handler: r.handleNotificationRules},
		{method: http.MethodPatch, path: "/api/ui/v1/notifications/rules/{id}", handler: r.handleNotificationRuleByID},
		{method: http.MethodDelete, path: "/api/ui/v1/notifications/rules/{id}", handler: r.handleNotificationRuleByID},
	}

	for _, route := range openRoutes {
		mux.HandleFunc(routePattern(route.method, route.path), route.handler)
	}
	for _, route := range trustedRoutes {
		mux.Handle(routePattern(route.method, route.path), r.withTrustedConsole(http.HandlerFunc(route.handler)))
	}

	mux.Handle(routePattern(http.MethodGet, "/api/ui/v1/events"), r.sse)
}

func (r *Router) registerTokenRoutes(mux *http.ServeMux, prefix string) {
	routes := []routeSpec{
		{method: http.MethodGet, path: "/dashboard", scope: domain.TokenScopeRead, handler: r.handleDashboard},
		{method: http.MethodGet, path: "/settings", scope: domain.TokenScopeRead, handler: r.handleSettings},
		{method: http.MethodGet, path: "/services", scope: domain.TokenScopeRead, handler: r.handleServices},
		{method: http.MethodPost, path: "/services", scope: domain.TokenScopeWrite, handler: r.handleServices},
		{method: http.MethodGet, path: "/services/{id}", scope: domain.TokenScopeRead, handler: r.handleServiceByID},
		{method: http.MethodPatch, path: "/services/{id}", scope: domain.TokenScopeWrite, handler: r.handleServiceByID},
		{method: http.MethodDelete, path: "/services/{id}", scope: domain.TokenScopeWrite, handler: r.handleServiceByID},
		{method: http.MethodGet, path: "/services/{id}/events", scope: domain.TokenScopeRead, handler: r.handleServiceEvents},
		{method: http.MethodGet, path: "/services/{id}/checks", scope: domain.TokenScopeRead, handler: r.handleServiceChecks},
		{method: http.MethodPost, path: "/services/{id}/checks", scope: domain.TokenScopeWrite, handler: r.handleServiceChecks},
		{method: http.MethodPost, path: "/services/{id}/checks/test", scope: domain.TokenScopeWrite, handler: r.handleServiceCheckTest},
		{method: http.MethodPatch, path: "/checks/{id}", scope: domain.TokenScopeWrite, handler: r.handleCheckByID},
		{method: http.MethodDelete, path: "/checks/{id}", scope: domain.TokenScopeWrite, handler: r.handleCheckByID},
		{method: http.MethodGet, path: "/devices", scope: domain.TokenScopeRead, handler: r.handleDevices},
		{method: http.MethodGet, path: "/devices/{id}", scope: domain.TokenScopeRead, handler: r.handleDeviceByID},
		{method: http.MethodPatch, path: "/devices/{id}", scope: domain.TokenScopeWrite, handler: r.handleDeviceByID},
		{method: http.MethodGet, path: "/bookmarks", scope: domain.TokenScopeRead, handler: r.handleBookmarks},
		{method: http.MethodPost, path: "/bookmarks", scope: domain.TokenScopeWrite, handler: r.handleBookmarks},
		{method: http.MethodPost, path: "/bookmarks/from-service", scope: domain.TokenScopeWrite, handler: r.handleBookmarkFromService},
		{method: http.MethodPost, path: "/bookmarks/reorder", scope: domain.TokenScopeWrite, handler: r.handleBookmarkReorder},
		{method: http.MethodPost, path: "/bookmarks/import", scope: domain.TokenScopeWrite, handler: r.handleBookmarkImport},
		{method: http.MethodGet, path: "/bookmarks/export", scope: domain.TokenScopeRead, handler: r.handleBookmarkExport},
		{method: http.MethodPut, path: "/bookmarks/{id}", scope: domain.TokenScopeWrite, handler: r.handleBookmarkByID},
		{method: http.MethodPatch, path: "/bookmarks/{id}", scope: domain.TokenScopeWrite, handler: r.handleBookmarkByID},
		{method: http.MethodGet, path: "/bookmarks/{id}/open", scope: domain.TokenScopeRead, handler: r.handleBookmarkOpen},
		{method: http.MethodDelete, path: "/bookmarks/{id}", scope: domain.TokenScopeWrite, handler: r.handleBookmarkByID},
		{method: http.MethodGet, path: "/folders", scope: domain.TokenScopeRead, handler: r.handleFolders},
		{method: http.MethodPost, path: "/folders", scope: domain.TokenScopeWrite, handler: r.handleFolders},
		{method: http.MethodPost, path: "/folders/reorder", scope: domain.TokenScopeWrite, handler: r.handleFolderReorder},
		{method: http.MethodPut, path: "/folders/{id}", scope: domain.TokenScopeWrite, handler: r.handleFolderByID},
		{method: http.MethodDelete, path: "/folders/{id}", scope: domain.TokenScopeWrite, handler: r.handleFolderByID},
		{method: http.MethodGet, path: "/tags", scope: domain.TokenScopeRead, handler: r.handleTags},
		{method: http.MethodGet, path: "/discovery/docker-endpoints", scope: domain.TokenScopeRead, handler: r.handleDockerEndpoints},
		{method: http.MethodPost, path: "/discovery/docker-endpoints", scope: domain.TokenScopeWrite, handler: r.handleDockerEndpoints},
		{method: http.MethodPatch, path: "/discovery/docker-endpoints/{id}", scope: domain.TokenScopeWrite, handler: r.handleDockerEndpointByID},
		{method: http.MethodDelete, path: "/discovery/docker-endpoints/{id}", scope: domain.TokenScopeWrite, handler: r.handleDockerEndpointByID},
		{method: http.MethodGet, path: "/discovery/scan-targets", scope: domain.TokenScopeRead, handler: r.handleScanTargets},
		{method: http.MethodPost, path: "/discovery/scan-targets", scope: domain.TokenScopeWrite, handler: r.handleScanTargets},
		{method: http.MethodPatch, path: "/discovery/scan-targets/{id}", scope: domain.TokenScopeWrite, handler: r.handleScanTargetByID},
		{method: http.MethodDelete, path: "/discovery/scan-targets/{id}", scope: domain.TokenScopeWrite, handler: r.handleScanTargetByID},
		{method: http.MethodGet, path: "/discovered-services", scope: domain.TokenScopeRead, handler: r.handleDiscoveredServices},
		{method: http.MethodPost, path: "/discovered-services/{id}/bookmark", scope: domain.TokenScopeWrite, handler: r.handleDiscoveredServiceBookmark},
		{method: http.MethodPost, path: "/discovered-services/{id}/ignore", scope: domain.TokenScopeWrite, handler: r.handleDiscoveredServiceIgnore},
		{method: http.MethodPost, path: "/discovered-services/{id}/restore", scope: domain.TokenScopeWrite, handler: r.handleDiscoveredServiceRestore},
		{method: http.MethodPatch, path: "/discovery/settings", scope: domain.TokenScopeWrite, handler: r.handleDiscoverySettings},
		{method: http.MethodPost, path: "/discovery/run", scope: domain.TokenScopeWrite, handler: r.handleDiscoveryRun},
		{method: http.MethodPost, path: "/monitoring/run", scope: domain.TokenScopeWrite, handler: r.handleMonitoringRun},
		{method: http.MethodGet, path: "/service-definitions", scope: domain.TokenScopeRead, handler: r.handleServiceDefinitions},
		{method: http.MethodPost, path: "/service-definitions", scope: domain.TokenScopeWrite, handler: r.handleServiceDefinitions},
		{method: http.MethodPatch, path: "/service-definitions/{id}", scope: domain.TokenScopeWrite, handler: r.handleServiceDefinitionByID},
		{method: http.MethodDelete, path: "/service-definitions/{id}", scope: domain.TokenScopeWrite, handler: r.handleServiceDefinitionByID},
		{method: http.MethodPost, path: "/service-definitions/{id}/reapply", scope: domain.TokenScopeWrite, handler: r.handleServiceDefinitionReapply},
		{method: http.MethodGet, path: "/notifications/channels", scope: domain.TokenScopeRead, handler: r.handleNotificationChannels},
		{method: http.MethodPost, path: "/notifications/channels", scope: domain.TokenScopeWrite, handler: r.handleNotificationChannels},
		{method: http.MethodPatch, path: "/notifications/channels/{id}", scope: domain.TokenScopeWrite, handler: r.handleNotificationChannelByID},
		{method: http.MethodDelete, path: "/notifications/channels/{id}", scope: domain.TokenScopeWrite, handler: r.handleNotificationChannelByID},
		{method: http.MethodPost, path: "/notifications/channels/{id}/test", scope: domain.TokenScopeWrite, handler: r.handleNotificationChannelTest},
		{method: http.MethodGet, path: "/notifications/rules", scope: domain.TokenScopeRead, handler: r.handleNotificationRules},
		{method: http.MethodPost, path: "/notifications/rules", scope: domain.TokenScopeWrite, handler: r.handleNotificationRules},
		{method: http.MethodPatch, path: "/notifications/rules/{id}", scope: domain.TokenScopeWrite, handler: r.handleNotificationRuleByID},
		{method: http.MethodDelete, path: "/notifications/rules/{id}", scope: domain.TokenScopeWrite, handler: r.handleNotificationRuleByID},
		{method: http.MethodGet, path: "/notifications/deliveries", scope: domain.TokenScopeRead, handler: r.handleNotificationDeliveries},
	}

	for _, route := range routes {
		mux.Handle(
			routePatternWithPrefix(route.method, prefix, route.path),
			r.withExternalToken(route.scope, http.HandlerFunc(route.handler)),
		)
	}
}
