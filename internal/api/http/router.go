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
	mux.HandleFunc(routePattern(http.MethodGet, "/healthz"), router.healthz)
	router.registerBootstrapRoutes(mux)
	mux.HandleFunc(routePattern(http.MethodGet, "/api/public/v1/status-pages/{slug}"), router.handlePublicStatusPage)
	router.registerUIRoutes(mux)
	router.registerTokenRoutes(mux, "/api/external/v1")
	router.registerTokenRoutes(mux, "/api/v1")
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
