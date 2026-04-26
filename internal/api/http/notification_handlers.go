package httpapi

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/deleema/homelabwatch/internal/domain"
)

func (r *Router) handleNotificationChannels(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		items, err := r.app.ListNotificationChannels(req.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var item domain.NotificationChannel
		if err := decodeJSON(req.Body, &item); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		saved, err := r.app.SaveNotificationChannel(req.Context(), item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, saved)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleNotificationChannelByID(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	switch req.Method {
	case http.MethodPatch:
		var item domain.NotificationChannel
		if err := decodeJSON(req.Body, &item); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item.ID = id
		saved, err := r.app.SaveNotificationChannel(req.Context(), item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, saved)
	case http.MethodDelete:
		if err := r.app.DeleteNotificationChannel(req.Context(), id); err != nil {
			writeLookupError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleNotificationChannelTest(w http.ResponseWriter, req *http.Request) {
	if err := decodeJSON(req.Body, &map[string]any{}); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	item, err := r.app.TestNotificationChannel(req.Context(), req.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (r *Router) handleNotificationRules(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		items, err := r.app.ListNotificationRules(req.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var item domain.NotificationRule
		if err := decodeJSON(req.Body, &item); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		saved, err := r.app.SaveNotificationRule(req.Context(), item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, saved)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleNotificationRuleByID(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	switch req.Method {
	case http.MethodPatch:
		var item domain.NotificationRule
		if err := decodeJSON(req.Body, &item); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item.ID = id
		saved, err := r.app.SaveNotificationRule(req.Context(), item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, saved)
	case http.MethodDelete:
		if err := r.app.DeleteNotificationRule(req.Context(), id); err != nil {
			writeLookupError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleNotificationDeliveries(w http.ResponseWriter, req *http.Request) {
	limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
	items, err := r.app.ListNotificationDeliveries(req.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}
