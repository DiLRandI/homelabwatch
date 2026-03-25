package httpapi

import (
	"net/http"
	"strings"

	"github.com/deleema/homelabwatch/internal/domain"
)

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
