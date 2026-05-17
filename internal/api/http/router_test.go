package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestTopologyRoutes(t *testing.T) {
	handler, application, _, _ := newRouterTestHarness(t)
	bootstrapTestApp(t, application)

	uiReq := httptest.NewRequest(http.MethodGet, "/api/ui/v1/topology", nil)
	uiRec := httptest.NewRecorder()
	handler.ServeHTTP(uiRec, uiReq)
	if uiRec.Code != http.StatusOK {
		t.Fatalf("expected ui topology route to succeed, got %d", uiRec.Code)
	}
	if !strings.Contains(uiRec.Body.String(), `"generatedAt"`) {
		t.Fatalf("expected topology json shape, got %q", uiRec.Body.String())
	}

	token, err := application.CreateAPIToken(context.Background(), domain.CreateAPITokenInput{Name: "Read only", Scope: domain.TokenScopeRead})
	if err != nil {
		t.Fatalf("create api token: %v", err)
	}
	externalReq := httptest.NewRequest(http.MethodGet, "/api/external/v1/topology", nil)
	externalReq.Header.Set("Authorization", "Bearer "+token.Secret)
	externalRec := httptest.NewRecorder()
	handler.ServeHTTP(externalRec, externalReq)
	if externalRec.Code != http.StatusOK {
		t.Fatalf("expected read token to access topology, got %d", externalRec.Code)
	}
}

func TestUITopologySourceCRUDRoutes(t *testing.T) {
	handler, application, _, _ := newRouterTestHarness(t)
	bootstrapTestApp(t, application)

	createReq := trustedJSONRequest(http.MethodPost, "/api/ui/v1/discovery/topology-sources", `{"name":"Core","address":"192.168.1.2","enabled":false,"snmpVersion":"v2c","community":"public","role":"switch","root":true}`)
	createRec := httptest.NewRecorder()
	handler.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected create source, got %d %q", createRec.Code, createRec.Body.String())
	}
	if strings.Contains(createRec.Body.String(), `"community":"public"`) || !strings.Contains(createRec.Body.String(), `"hasCommunity":true`) {
		t.Fatalf("expected redacted credential response, got %q", createRec.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/ui/v1/discovery/topology-sources", nil)
	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK || !strings.Contains(listRec.Body.String(), `"name":"Core"`) {
		t.Fatalf("expected source list, got %d %q", listRec.Code, listRec.Body.String())
	}
	sourceID := strings.Split(strings.Split(createRec.Body.String(), `"id":"`)[1], `"`)[0]

	patchReq := trustedJSONRequest(http.MethodPatch, "/api/ui/v1/discovery/topology-sources/"+sourceID, `{"name":"Core renamed"}`)
	patchRec := httptest.NewRecorder()
	handler.ServeHTTP(patchRec, patchReq)
	if patchRec.Code != http.StatusOK || !strings.Contains(patchRec.Body.String(), `"name":"Core renamed"`) || !strings.Contains(patchRec.Body.String(), `"hasCommunity":true`) {
		t.Fatalf("expected patched source preserving secret, got %d %q", patchRec.Code, patchRec.Body.String())
	}

	runReq := trustedJSONRequest(http.MethodPost, "/api/ui/v1/discovery/topology/run", `{}`)
	runRec := httptest.NewRecorder()
	handler.ServeHTTP(runRec, runReq)
	if runRec.Code != http.StatusAccepted {
		t.Fatalf("expected topology run route to execute, got %d %q", runRec.Code, runRec.Body.String())
	}

	deleteReq := trustedJSONRequest(http.MethodDelete, "/api/ui/v1/discovery/topology-sources/"+sourceID, ``)
	deleteRec := httptest.NewRecorder()
	handler.ServeHTTP(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusNoContent {
		t.Fatalf("expected deleted source, got %d %q", deleteRec.Code, deleteRec.Body.String())
	}
}

func TestExternalTopologySourceRoutesEnforceScope(t *testing.T) {
	handler, application, _, _ := newRouterTestHarness(t)
	bootstrapTestApp(t, application)

	readToken, err := application.CreateAPIToken(context.Background(), domain.CreateAPITokenInput{Name: "Read", Scope: domain.TokenScopeRead})
	if err != nil {
		t.Fatalf("create read token: %v", err)
	}
	writeToken, err := application.CreateAPIToken(context.Background(), domain.CreateAPITokenInput{Name: "Write", Scope: domain.TokenScopeWrite})
	if err != nil {
		t.Fatalf("create write token: %v", err)
	}

	readReq := httptest.NewRequest(http.MethodGet, "/api/external/v1/discovery/topology-sources", nil)
	readReq.Header.Set("Authorization", "Bearer "+readToken.Secret)
	readRec := httptest.NewRecorder()
	handler.ServeHTTP(readRec, readReq)
	if readRec.Code != http.StatusOK {
		t.Fatalf("expected read token to list sources, got %d", readRec.Code)
	}

	blockedReq := httptest.NewRequest(http.MethodPost, "/api/external/v1/discovery/topology-sources", strings.NewReader(`{"name":"Core","address":"192.168.1.2","enabled":true,"snmpVersion":"v2c"}`))
	blockedReq.Header.Set("Authorization", "Bearer "+readToken.Secret)
	blockedReq.Header.Set("Content-Type", "application/json")
	blockedRec := httptest.NewRecorder()
	handler.ServeHTTP(blockedRec, blockedReq)
	if blockedRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected read token mutation rejection, got %d", blockedRec.Code)
	}

	writeReq := httptest.NewRequest(http.MethodPost, "/api/external/v1/discovery/topology-sources", strings.NewReader(`{"name":"Core","address":"192.168.1.2","enabled":true,"snmpVersion":"v2c"}`))
	writeReq.Header.Set("Authorization", "Bearer "+writeToken.Secret)
	writeReq.Header.Set("Content-Type", "application/json")
	writeRec := httptest.NewRecorder()
	handler.ServeHTTP(writeRec, writeReq)
	if writeRec.Code != http.StatusCreated {
		t.Fatalf("expected write token mutation, got %d %q", writeRec.Code, writeRec.Body.String())
	}
}

func TestPublicStatusPageRouteAndStaticFallback(t *testing.T) {
	handler, application, _, cfg := newRouterTestHarness(t)
	bootstrapTestApp(t, application)
	if err := os.WriteFile(filepath.Join(cfg.StaticDir, "index.html"), []byte("<html>app</html>"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	if _, err := application.SaveStatusPage(context.Background(), domain.StatusPageInput{Slug: "example", Title: "Example"}); err != nil {
		t.Fatalf("save status page: %v", err)
	}

	publicReq := httptest.NewRequest(http.MethodGet, "/api/public/v1/status-pages/example", nil)
	publicRec := httptest.NewRecorder()
	handler.ServeHTTP(publicRec, publicReq)
	if publicRec.Code != http.StatusOK {
		t.Fatalf("expected public json, got %d", publicRec.Code)
	}
	if !strings.Contains(publicRec.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("expected json content type, got %q", publicRec.Header().Get("Content-Type"))
	}

	staticReq := httptest.NewRequest(http.MethodGet, "/status/example", nil)
	staticRec := httptest.NewRecorder()
	handler.ServeHTTP(staticRec, staticReq)
	if staticRec.Code != http.StatusOK || !strings.Contains(staticRec.Body.String(), "<html>app</html>") {
		t.Fatalf("expected status route to fall through to static index, got %d %q", staticRec.Code, staticRec.Body.String())
	}
}

func trustedJSONRequest(method, path, body string) *http.Request {
	if body == "" {
		body = "{}"
	}
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Host = "localhost:8080"
	req.RemoteAddr = "127.0.0.1:41234"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:8080")
	req.Header.Set(consoleCSRFHeader, "test-csrf")
	req.AddCookie(&http.Cookie{Name: consoleCSRFCookie, Value: "test-csrf"})
	return req
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

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	application := app.New(cfg, store, events.NewBus(), logger)
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
