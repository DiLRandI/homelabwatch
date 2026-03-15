package httpapi

import (
	"net/http"

	"github.com/deleema/homelabwatch/internal/domain"
)

func (r *Router) handleServiceCheckTest(w http.ResponseWriter, req *http.Request) {
	var input domain.EndpointTestInput
	if err := decodeJSON(req.Body, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	item, err := r.app.TestServiceCheck(req.Context(), req.PathValue("id"), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (r *Router) handleServiceDefinitions(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		items, err := r.app.ListServiceDefinitions(req.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var input domain.ServiceDefinitionInput
		if err := decodeJSON(req.Body, &input); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item, err := r.app.SaveServiceDefinition(req.Context(), input)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, item)
	}
}

func (r *Router) handleServiceDefinitionByID(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	switch req.Method {
	case http.MethodPatch:
		var input domain.ServiceDefinitionInput
		if err := decodeJSON(req.Body, &input); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		input.ID = id
		item, err := r.app.SaveServiceDefinition(req.Context(), input)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodDelete:
		if err := r.app.DeleteServiceDefinition(req.Context(), id); err != nil {
			writeLookupError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (r *Router) handleServiceDefinitionReapply(w http.ResponseWriter, req *http.Request) {
	if err := r.app.ReapplyServiceDefinition(req.Context(), req.PathValue("id")); err != nil {
		writeLookupError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
