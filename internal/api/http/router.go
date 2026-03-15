package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"strings"

	"github.com/deleema/homelabwatch/internal/api/sse"
	"github.com/deleema/homelabwatch/internal/app"
	"github.com/deleema/homelabwatch/internal/config"
	"github.com/deleema/homelabwatch/internal/domain"
)

type Router struct {
	app             *app.App
	config          config.Config
	sse             http.Handler
	trustedNetworks []netip.Prefix
}

func NewRouter(application *app.App, cfg config.Config) http.Handler {
	router := &Router{
		app:             application,
		config:          cfg,
		sse:             sse.NewHandler(busAdapter{application: application}),
		trustedNetworks: parseTrustedNetworks(cfg.TrustedCIDRs),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", router.healthz)
	mux.HandleFunc("GET /api/v1/bootstrap/status", router.handleBootstrapStatus)
	mux.HandleFunc("GET /api/ui/v1/bootstrap", router.handleUIBootstrap)
	mux.Handle("POST /api/ui/v1/setup", router.withTrustedConsole(http.HandlerFunc(router.handleSetup)))
	mux.HandleFunc("GET /api/ui/v1/dashboard", router.handleDashboard)
	mux.HandleFunc("GET /api/ui/v1/settings", router.handleSettings)
	mux.Handle("POST /api/ui/v1/settings/api-tokens", router.withTrustedConsole(http.HandlerFunc(router.handleAPITokens)))
	mux.Handle("DELETE /api/ui/v1/settings/api-tokens/{id}", router.withTrustedConsole(http.HandlerFunc(router.handleAPITokenByID)))
	mux.HandleFunc("GET /api/ui/v1/services", router.handleServices)
	mux.Handle("POST /api/ui/v1/services", router.withTrustedConsole(http.HandlerFunc(router.handleServices)))
	mux.HandleFunc("GET /api/ui/v1/services/{id}", router.handleServiceByID)
	mux.Handle("PATCH /api/ui/v1/services/{id}", router.withTrustedConsole(http.HandlerFunc(router.handleServiceByID)))
	mux.Handle("DELETE /api/ui/v1/services/{id}", router.withTrustedConsole(http.HandlerFunc(router.handleServiceByID)))
	mux.HandleFunc("GET /api/ui/v1/services/{id}/events", router.handleServiceEvents)
	mux.HandleFunc("GET /api/ui/v1/services/{id}/checks", router.handleServiceChecks)
	mux.Handle("POST /api/ui/v1/services/{id}/checks", router.withTrustedConsole(http.HandlerFunc(router.handleServiceChecks)))
	mux.Handle("PATCH /api/ui/v1/checks/{id}", router.withTrustedConsole(http.HandlerFunc(router.handleCheckByID)))
	mux.Handle("DELETE /api/ui/v1/checks/{id}", router.withTrustedConsole(http.HandlerFunc(router.handleCheckByID)))
	mux.HandleFunc("GET /api/ui/v1/devices", router.handleDevices)
	mux.HandleFunc("GET /api/ui/v1/devices/{id}", router.handleDeviceByID)
	mux.Handle("PATCH /api/ui/v1/devices/{id}", router.withTrustedConsole(http.HandlerFunc(router.handleDeviceByID)))
	mux.HandleFunc("GET /api/ui/v1/bookmarks", router.handleBookmarks)
	mux.Handle("POST /api/ui/v1/bookmarks", router.withTrustedConsole(http.HandlerFunc(router.handleBookmarks)))
	mux.Handle("PATCH /api/ui/v1/bookmarks/{id}", router.withTrustedConsole(http.HandlerFunc(router.handleBookmarkByID)))
	mux.Handle("DELETE /api/ui/v1/bookmarks/{id}", router.withTrustedConsole(http.HandlerFunc(router.handleBookmarkByID)))
	mux.HandleFunc("GET /api/ui/v1/discovery/docker-endpoints", router.handleDockerEndpoints)
	mux.Handle("POST /api/ui/v1/discovery/docker-endpoints", router.withTrustedConsole(http.HandlerFunc(router.handleDockerEndpoints)))
	mux.Handle("PATCH /api/ui/v1/discovery/docker-endpoints/{id}", router.withTrustedConsole(http.HandlerFunc(router.handleDockerEndpointByID)))
	mux.Handle("DELETE /api/ui/v1/discovery/docker-endpoints/{id}", router.withTrustedConsole(http.HandlerFunc(router.handleDockerEndpointByID)))
	mux.HandleFunc("GET /api/ui/v1/discovery/scan-targets", router.handleScanTargets)
	mux.Handle("POST /api/ui/v1/discovery/scan-targets", router.withTrustedConsole(http.HandlerFunc(router.handleScanTargets)))
	mux.Handle("PATCH /api/ui/v1/discovery/scan-targets/{id}", router.withTrustedConsole(http.HandlerFunc(router.handleScanTargetByID)))
	mux.Handle("DELETE /api/ui/v1/discovery/scan-targets/{id}", router.withTrustedConsole(http.HandlerFunc(router.handleScanTargetByID)))
	mux.Handle("POST /api/ui/v1/discovery/run", router.withTrustedConsole(http.HandlerFunc(router.handleDiscoveryRun)))
	mux.Handle("POST /api/ui/v1/monitoring/run", router.withTrustedConsole(http.HandlerFunc(router.handleMonitoringRun)))
	mux.Handle("GET /api/ui/v1/events", router.sse)

	mux.Handle("GET /api/external/v1/dashboard", router.withExternalToken(domain.TokenScopeRead, http.HandlerFunc(router.handleDashboard)))
	mux.Handle("GET /api/external/v1/settings", router.withExternalToken(domain.TokenScopeRead, http.HandlerFunc(router.handleSettings)))
	mux.Handle("GET /api/external/v1/services", router.withExternalToken(domain.TokenScopeRead, http.HandlerFunc(router.handleServices)))
	mux.Handle("POST /api/external/v1/services", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleServices)))
	mux.Handle("GET /api/external/v1/services/{id}", router.withExternalToken(domain.TokenScopeRead, http.HandlerFunc(router.handleServiceByID)))
	mux.Handle("PATCH /api/external/v1/services/{id}", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleServiceByID)))
	mux.Handle("DELETE /api/external/v1/services/{id}", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleServiceByID)))
	mux.Handle("GET /api/external/v1/services/{id}/events", router.withExternalToken(domain.TokenScopeRead, http.HandlerFunc(router.handleServiceEvents)))
	mux.Handle("GET /api/external/v1/services/{id}/checks", router.withExternalToken(domain.TokenScopeRead, http.HandlerFunc(router.handleServiceChecks)))
	mux.Handle("POST /api/external/v1/services/{id}/checks", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleServiceChecks)))
	mux.Handle("PATCH /api/external/v1/checks/{id}", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleCheckByID)))
	mux.Handle("DELETE /api/external/v1/checks/{id}", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleCheckByID)))
	mux.Handle("GET /api/external/v1/devices", router.withExternalToken(domain.TokenScopeRead, http.HandlerFunc(router.handleDevices)))
	mux.Handle("GET /api/external/v1/devices/{id}", router.withExternalToken(domain.TokenScopeRead, http.HandlerFunc(router.handleDeviceByID)))
	mux.Handle("PATCH /api/external/v1/devices/{id}", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleDeviceByID)))
	mux.Handle("GET /api/external/v1/bookmarks", router.withExternalToken(domain.TokenScopeRead, http.HandlerFunc(router.handleBookmarks)))
	mux.Handle("POST /api/external/v1/bookmarks", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleBookmarks)))
	mux.Handle("PATCH /api/external/v1/bookmarks/{id}", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleBookmarkByID)))
	mux.Handle("DELETE /api/external/v1/bookmarks/{id}", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleBookmarkByID)))
	mux.Handle("GET /api/external/v1/discovery/docker-endpoints", router.withExternalToken(domain.TokenScopeRead, http.HandlerFunc(router.handleDockerEndpoints)))
	mux.Handle("POST /api/external/v1/discovery/docker-endpoints", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleDockerEndpoints)))
	mux.Handle("PATCH /api/external/v1/discovery/docker-endpoints/{id}", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleDockerEndpointByID)))
	mux.Handle("DELETE /api/external/v1/discovery/docker-endpoints/{id}", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleDockerEndpointByID)))
	mux.Handle("GET /api/external/v1/discovery/scan-targets", router.withExternalToken(domain.TokenScopeRead, http.HandlerFunc(router.handleScanTargets)))
	mux.Handle("POST /api/external/v1/discovery/scan-targets", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleScanTargets)))
	mux.Handle("PATCH /api/external/v1/discovery/scan-targets/{id}", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleScanTargetByID)))
	mux.Handle("DELETE /api/external/v1/discovery/scan-targets/{id}", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleScanTargetByID)))
	mux.Handle("POST /api/external/v1/discovery/run", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleDiscoveryRun)))
	mux.Handle("POST /api/external/v1/monitoring/run", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleMonitoringRun)))

	mux.Handle("GET /api/v1/dashboard", router.withExternalToken(domain.TokenScopeRead, http.HandlerFunc(router.handleDashboard)))
	mux.Handle("GET /api/v1/settings", router.withExternalToken(domain.TokenScopeRead, http.HandlerFunc(router.handleSettings)))
	mux.Handle("GET /api/v1/services", router.withExternalToken(domain.TokenScopeRead, http.HandlerFunc(router.handleServices)))
	mux.Handle("POST /api/v1/services", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleServices)))
	mux.Handle("GET /api/v1/services/{id}", router.withExternalToken(domain.TokenScopeRead, http.HandlerFunc(router.handleServiceByID)))
	mux.Handle("PATCH /api/v1/services/{id}", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleServiceByID)))
	mux.Handle("DELETE /api/v1/services/{id}", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleServiceByID)))
	mux.Handle("GET /api/v1/services/{id}/events", router.withExternalToken(domain.TokenScopeRead, http.HandlerFunc(router.handleServiceEvents)))
	mux.Handle("GET /api/v1/services/{id}/checks", router.withExternalToken(domain.TokenScopeRead, http.HandlerFunc(router.handleServiceChecks)))
	mux.Handle("POST /api/v1/services/{id}/checks", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleServiceChecks)))
	mux.Handle("PATCH /api/v1/checks/{id}", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleCheckByID)))
	mux.Handle("DELETE /api/v1/checks/{id}", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleCheckByID)))
	mux.Handle("GET /api/v1/devices", router.withExternalToken(domain.TokenScopeRead, http.HandlerFunc(router.handleDevices)))
	mux.Handle("GET /api/v1/devices/{id}", router.withExternalToken(domain.TokenScopeRead, http.HandlerFunc(router.handleDeviceByID)))
	mux.Handle("PATCH /api/v1/devices/{id}", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleDeviceByID)))
	mux.Handle("GET /api/v1/bookmarks", router.withExternalToken(domain.TokenScopeRead, http.HandlerFunc(router.handleBookmarks)))
	mux.Handle("POST /api/v1/bookmarks", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleBookmarks)))
	mux.Handle("PATCH /api/v1/bookmarks/{id}", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleBookmarkByID)))
	mux.Handle("DELETE /api/v1/bookmarks/{id}", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleBookmarkByID)))
	mux.Handle("GET /api/v1/discovery/docker-endpoints", router.withExternalToken(domain.TokenScopeRead, http.HandlerFunc(router.handleDockerEndpoints)))
	mux.Handle("POST /api/v1/discovery/docker-endpoints", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleDockerEndpoints)))
	mux.Handle("PATCH /api/v1/discovery/docker-endpoints/{id}", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleDockerEndpointByID)))
	mux.Handle("DELETE /api/v1/discovery/docker-endpoints/{id}", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleDockerEndpointByID)))
	mux.Handle("GET /api/v1/discovery/scan-targets", router.withExternalToken(domain.TokenScopeRead, http.HandlerFunc(router.handleScanTargets)))
	mux.Handle("POST /api/v1/discovery/scan-targets", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleScanTargets)))
	mux.Handle("PATCH /api/v1/discovery/scan-targets/{id}", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleScanTargetByID)))
	mux.Handle("DELETE /api/v1/discovery/scan-targets/{id}", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleScanTargetByID)))
	mux.Handle("POST /api/v1/discovery/run", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleDiscoveryRun)))
	mux.Handle("POST /api/v1/monitoring/run", router.withExternalToken(domain.TokenScopeWrite, http.HandlerFunc(router.handleMonitoringRun)))
	mux.HandleFunc("/", router.handleStatic)
	return mux
}

type busAdapter struct {
	application *app.App
}

func (b busAdapter) Subscribe(buffer int) chan domain.EventEnvelope {
	return b.application.SubscribeEvents(buffer)
}
func (b busAdapter) Unsubscribe(ch chan domain.EventEnvelope) { b.application.UnsubscribeEvents(ch) }

func (r *Router) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (r *Router) handleBootstrapStatus(w http.ResponseWriter, req *http.Request) {
	status, err := r.app.BootstrapStatus(req.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (r *Router) handleUIBootstrap(w http.ResponseWriter, req *http.Request) {
	status, err := r.app.BootstrapStatus(req.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	csrfToken, err := issueConsoleCSRF(w, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, domain.UIBootstrap{
		Initialized:    status.Initialized,
		TrustedNetwork: r.isTrustedNetwork(req),
		CSRFToken:      csrfToken,
	})
}

func (r *Router) handleSetup(w http.ResponseWriter, req *http.Request) {
	var input domain.SetupInput
	if err := decodeJSON(req.Body, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := r.app.Setup(req.Context(), input); err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "already completed") {
			status = http.StatusConflict
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "initialized"})
}

func (r *Router) handleDashboard(w http.ResponseWriter, req *http.Request) {
	item, err := r.app.Dashboard(req.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (r *Router) handleSettings(w http.ResponseWriter, req *http.Request) {
	item, err := r.app.Settings(req.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	item.AppSettings.TrustedCIDRs = append([]string(nil), r.config.TrustedCIDRs...)
	item.AppSettings.TrustedNetwork = r.isTrustedNetwork(req)
	writeJSON(w, http.StatusOK, item)
}

func (r *Router) handleAPITokens(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		var input domain.CreateAPITokenInput
		if err := decodeJSON(req.Body, &input); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item, err := r.app.CreateAPIToken(req.Context(), input)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleAPITokenByID(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodDelete:
		if err := r.app.RevokeAPIToken(req.Context(), req.PathValue("id")); err != nil {
			writeLookupError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleServices(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		items, err := r.app.ListServices(req.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var item domain.Service
		if err := decodeJSON(req.Body, &item); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		saved, err := r.app.SaveManualService(req.Context(), item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, saved)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleServiceByID(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	switch req.Method {
	case http.MethodGet:
		item, err := r.app.GetService(req.Context(), id)
		if err != nil {
			writeLookupError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodPatch:
		var item domain.Service
		if err := decodeJSON(req.Body, &item); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item.ID = id
		saved, err := r.app.SaveManualService(req.Context(), item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, saved)
	case http.MethodDelete:
		if err := r.app.DeleteService(req.Context(), id); err != nil {
			writeLookupError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (r *Router) handleServiceEvents(w http.ResponseWriter, req *http.Request) {
	items, err := r.app.ListServiceEvents(req.Context(), req.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (r *Router) handleServiceChecks(w http.ResponseWriter, req *http.Request) {
	serviceID := req.PathValue("id")
	switch req.Method {
	case http.MethodGet:
		service, err := r.app.GetService(req.Context(), serviceID)
		if err != nil {
			writeLookupError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, service.Checks)
	case http.MethodPost:
		var item domain.ServiceCheck
		if err := decodeJSON(req.Body, &item); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item.ServiceID = serviceID
		saved, err := r.app.SaveServiceCheck(req.Context(), item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, saved)
	}
}

func (r *Router) handleCheckByID(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	switch req.Method {
	case http.MethodPatch:
		var item domain.ServiceCheck
		if err := decodeJSON(req.Body, &item); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item.ID = id
		saved, err := r.app.SaveServiceCheck(req.Context(), item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, saved)
	case http.MethodDelete:
		if err := r.app.DeleteServiceCheck(req.Context(), id); err != nil {
			writeLookupError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (r *Router) handleDevices(w http.ResponseWriter, req *http.Request) {
	items, err := r.app.ListDevices(req.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (r *Router) handleDeviceByID(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	switch req.Method {
	case http.MethodGet:
		item, err := r.app.GetDevice(req.Context(), id)
		if err != nil {
			writeLookupError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodPatch:
		var payload struct {
			DisplayName *string `json:"displayName"`
			Hidden      *bool   `json:"hidden"`
		}
		if err := decodeJSON(req.Body, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item, err := r.app.UpdateDevice(req.Context(), id, payload.DisplayName, payload.Hidden)
		if err != nil {
			writeLookupError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	}
}

func (r *Router) handleBookmarks(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		items, err := r.app.ListBookmarks(req.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var item domain.Bookmark
		if err := decodeJSON(req.Body, &item); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		saved, err := r.app.SaveBookmark(req.Context(), item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, saved)
	}
}

func (r *Router) handleBookmarkByID(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	switch req.Method {
	case http.MethodPatch:
		var item domain.Bookmark
		if err := decodeJSON(req.Body, &item); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item.ID = id
		saved, err := r.app.SaveBookmark(req.Context(), item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, saved)
	case http.MethodDelete:
		if err := r.app.DeleteBookmark(req.Context(), id); err != nil {
			writeLookupError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (r *Router) handleDockerEndpoints(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		items, err := r.app.ListDockerEndpoints(req.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var item domain.DockerEndpoint
		if err := decodeJSON(req.Body, &item); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		saved, err := r.app.SaveDockerEndpoint(req.Context(), item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, saved)
	}
}

func (r *Router) handleDockerEndpointByID(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	switch req.Method {
	case http.MethodPatch:
		var item domain.DockerEndpoint
		if err := decodeJSON(req.Body, &item); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item.ID = id
		saved, err := r.app.SaveDockerEndpoint(req.Context(), item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, saved)
	case http.MethodDelete:
		if err := r.app.DeleteDockerEndpoint(req.Context(), id); err != nil {
			writeLookupError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (r *Router) handleScanTargets(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		items, err := r.app.ListScanTargets(req.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var item domain.ScanTarget
		if err := decodeJSON(req.Body, &item); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		saved, err := r.app.SaveScanTarget(req.Context(), item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, saved)
	}
}

func (r *Router) handleScanTargetByID(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	switch req.Method {
	case http.MethodPatch:
		var item domain.ScanTarget
		if err := decodeJSON(req.Body, &item); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item.ID = id
		saved, err := r.app.SaveScanTarget(req.Context(), item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, saved)
	case http.MethodDelete:
		if err := r.app.DeleteScanTarget(req.Context(), id); err != nil {
			writeLookupError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (r *Router) handleDiscoveryRun(w http.ResponseWriter, req *http.Request) {
	if err := r.app.TriggerDiscovery(req.Context()); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "discovery-triggered"})
}

func (r *Router) handleMonitoringRun(w http.ResponseWriter, req *http.Request) {
	if err := r.app.TriggerMonitoring(req.Context()); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "monitoring-triggered"})
}

func (r *Router) handleStatic(w http.ResponseWriter, req *http.Request) {
	if strings.HasPrefix(req.URL.Path, "/api/") {
		http.NotFound(w, req)
		return
	}
	cleanPath := filepath.Clean(strings.TrimPrefix(req.URL.Path, "/"))
	if cleanPath == "." {
		cleanPath = "index.html"
	}
	target := filepath.Join(r.config.StaticDir, cleanPath)
	if info, err := os.Stat(target); err == nil && !info.IsDir() {
		http.ServeFile(w, req, target)
		return
	}
	http.ServeFile(w, req, filepath.Join(r.config.StaticDir, "index.html"))
}

func decodeJSON(body io.ReadCloser, target any) error {
	defer body.Close()
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeLookupError(w http.ResponseWriter, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeError(w, http.StatusInternalServerError, err)
}
