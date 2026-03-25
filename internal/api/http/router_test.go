package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deleema/homelabwatch/internal/app"
	"github.com/deleema/homelabwatch/internal/config"
	"github.com/deleema/homelabwatch/internal/domain"
	"github.com/deleema/homelabwatch/internal/events"
	"github.com/deleema/homelabwatch/internal/store/sqlite"
)

func TestHealthzAndBootstrapRoutes(t *testing.T) {
	handler, _, _, _ := newRouterTestHarness(t)

	healthReq := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	healthRec := httptest.NewRecorder()
	handler.ServeHTTP(healthRec, healthReq)
	if healthRec.Code != http.StatusOK {
		t.Fatalf("expected healthz to succeed, got %d", healthRec.Code)
	}

	bootstrapReq := httptest.NewRequest(http.MethodGet, "/api/ui/v1/bootstrap", nil)
	bootstrapReq.Host = "localhost:8080"
	bootstrapReq.RemoteAddr = "127.0.0.1:41234"
	bootstrapRec := httptest.NewRecorder()
	handler.ServeHTTP(bootstrapRec, bootstrapReq)
	if bootstrapRec.Code != http.StatusOK {
		t.Fatalf("expected bootstrap to succeed, got %d", bootstrapRec.Code)
	}
	if len(bootstrapRec.Result().Cookies()) == 0 {
		t.Fatalf("expected bootstrap to issue a console csrf cookie")
	}
}

func TestTrustedConsoleRouteRequiresCSRF(t *testing.T) {
	handler, _, _, _ := newRouterTestHarness(t)

	req := httptest.NewRequest(http.MethodPost, "/api/ui/v1/setup", strings.NewReader(`{"applianceName":"Lab","defaultScanPorts":[22,80]}`))
	req.Host = "localhost:8080"
	req.RemoteAddr = "127.0.0.1:41234"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:8080")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected missing csrf to be rejected, got %d", rec.Code)
	}
}

func TestExternalTokenRoutesEnforceScope(t *testing.T) {
	handler, application, _, _ := newRouterTestHarness(t)
	bootstrapTestApp(t, application)

	token, err := application.CreateAPIToken(context.Background(), domain.CreateAPITokenInput{
		Name:  "Read only",
		Scope: domain.TokenScopeRead,
	})
	if err != nil {
		t.Fatalf("create api token: %v", err)
	}

	readReq := httptest.NewRequest(http.MethodGet, "/api/external/v1/services", nil)
	readReq.Header.Set("Authorization", "Bearer "+token.Secret)
	readRec := httptest.NewRecorder()
	handler.ServeHTTP(readRec, readReq)
	if readRec.Code != http.StatusOK {
		t.Fatalf("expected read token to access services, got %d", readRec.Code)
	}

	writeReq := httptest.NewRequest(http.MethodPost, "/api/external/v1/discovery/run", nil)
	writeReq.Header.Set("Authorization", "Bearer "+token.Secret)
	writeRec := httptest.NewRecorder()
	handler.ServeHTTP(writeRec, writeReq)
	if writeRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected read token to fail write route, got %d", writeRec.Code)
	}
}

func newRouterTestHarness(t *testing.T) (http.Handler, *app.App, *sqlite.Store, config.Config) {
	t.Helper()

	dataDir := t.TempDir()
	cfg := config.Config{
		ListenAddr:       ":0",
		DataDir:          dataDir,
		DBPath:           filepath.Join(dataDir, "homelabwatch.db"),
		StaticDir:        dataDir,
		SeedDockerSocket: false,
		DefaultScanPorts: []int{22, 80, 443},
		TrustedCIDRs:     []string{"127.0.0.1/32"},
	}

	store, err := sqlite.New(cfg.DBPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	application := app.New(cfg, store, events.NewBus())
	return NewRouter(application, cfg), application, store, cfg
}

func bootstrapTestApp(t *testing.T, application *app.App) {
	t.Helper()

	err := application.Setup(context.Background(), domain.SetupInput{
		ApplianceName:    "Rack Alpha",
		AutoScanEnabled:  true,
		DefaultScanPorts: []int{22, 80},
		RunDiscovery:     false,
	})
	if err != nil {
		t.Fatalf("bootstrap app: %v", err)
	}
}
