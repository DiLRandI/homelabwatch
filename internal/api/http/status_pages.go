package httpapi

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
)

type statusPageServicesPayload struct {
	Services []domain.StatusPageServiceInput `json:"services"`
}

func (r *Router) handlePublicStatusPage(w http.ResponseWriter, req *http.Request) {
	item, err := r.app.GetPublicStatusPage(req.Context(), req.PathValue("slug"), time.Now().UTC())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (r *Router) handleStatusPages(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		items, err := r.app.ListStatusPages(req.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var input domain.StatusPageInput
		if err := decodeJSON(req.Body, &input); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item, err := r.app.SaveStatusPage(req.Context(), input)
		writeStatusPageSaveResponse(w, http.StatusCreated, item, err)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleStatusPageByID(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	switch req.Method {
	case http.MethodGet:
		item, err := r.app.GetStatusPage(req.Context(), id)
		if err != nil {
			writeLookupError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodPatch:
		var input domain.StatusPageInput
		if err := decodeJSON(req.Body, &input); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		input.ID = id
		item, err := r.app.SaveStatusPage(req.Context(), input)
		writeStatusPageSaveResponse(w, http.StatusOK, item, err)
	case http.MethodDelete:
		if err := r.app.DeleteStatusPage(req.Context(), id); err != nil {
			writeLookupError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleStatusPageServices(w http.ResponseWriter, req *http.Request) {
	var payload statusPageServicesPayload
	if err := decodeJSON(req.Body, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	item, err := r.app.ReplaceStatusPageServices(req.Context(), req.PathValue("id"), payload.Services)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (r *Router) handleStatusPageAnnouncements(w http.ResponseWriter, req *http.Request) {
	var input domain.StatusPageAnnouncementInput
	if err := decodeJSON(req.Body, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	item, err := r.app.CreateStatusPageAnnouncement(req.Context(), req.PathValue("id"), input)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (r *Router) handleStatusPageAnnouncementByID(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	switch req.Method {
	case http.MethodPatch:
		var input domain.StatusPageAnnouncementInput
		if err := decodeJSON(req.Body, &input); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item, err := r.app.UpdateStatusPageAnnouncement(req.Context(), id, input)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, err)
				return
			}
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodDelete:
		if err := r.app.DeleteStatusPageAnnouncement(req.Context(), id); err != nil {
			writeLookupError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func writeStatusPageSaveResponse(w http.ResponseWriter, status int, item domain.StatusPage, err error) {
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		if strings.Contains(strings.ToLower(err.Error()), "unique") && strings.Contains(strings.ToLower(err.Error()), "status_pages.slug") {
			writeError(w, http.StatusConflict, errors.New("status page slug already exists"))
			return
		}
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, status, item)
}
