package httpapi

import (
	"errors"
	"io"
	"net/http"

	"github.com/deleema/homelabwatch/internal/domain"
)

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
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
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
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
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
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
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
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleDiscoveredServices(w http.ResponseWriter, req *http.Request) {
	items, err := r.app.ListDiscoveredServices(req.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (r *Router) handleDiscoveredServiceBookmark(w http.ResponseWriter, req *http.Request) {
	var input domain.CreateBookmarkFromDiscoveredServiceInput
	if err := decodeJSON(req.Body, &input); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	input.DiscoveredServiceID = req.PathValue("id")
	item, err := r.app.CreateBookmarkFromDiscoveredService(req.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (r *Router) handleDiscoveredServiceIgnore(w http.ResponseWriter, req *http.Request) {
	item, err := r.app.IgnoreDiscoveredService(req.Context(), req.PathValue("id"))
	if err != nil {
		writeLookupError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (r *Router) handleDiscoveredServiceRestore(w http.ResponseWriter, req *http.Request) {
	item, err := r.app.RestoreDiscoveredService(req.Context(), req.PathValue("id"))
	if err != nil {
		writeLookupError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (r *Router) handleDiscoverySettings(w http.ResponseWriter, req *http.Request) {
	var input domain.DiscoverySettings
	if err := decodeJSON(req.Body, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	item, err := r.app.SaveDiscoverySettings(req.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
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
