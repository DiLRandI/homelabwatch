package httpapi

import (
	"net/http"

	"github.com/deleema/homelabwatch/internal/domain"
)

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
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
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
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
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
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
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
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
