package app

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/deleema/homelabwatch/internal/config"
	"github.com/deleema/homelabwatch/internal/domain"
	"github.com/deleema/homelabwatch/internal/events"
	"github.com/deleema/homelabwatch/internal/store/sqlite"
)

func TestSetupInitializesWorkspace(t *testing.T) {
	application, store, _ := newTestApp(t, config.Config{
		DefaultScanPorts: []int{22, 80},
	})

	err := application.Setup(context.Background(), domain.SetupInput{
		ApplianceName:    "Rack Alpha",
		AutoScanEnabled:  true,
		DefaultScanPorts: []int{22, 80},
		RunDiscovery:     false,
	})
	if err != nil {
		t.Fatalf("setup workspace: %v", err)
	}

	status, err := store.BootstrapStatus(context.Background())
	if err != nil {
		t.Fatalf("bootstrap status: %v", err)
	}
	if !status.Initialized {
		t.Fatalf("expected initialized store")
	}

	settings, err := application.Settings(context.Background())
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if settings.AppSettings.ApplianceName != "Rack Alpha" {
		t.Fatalf("expected appliance name to persist, got %q", settings.AppSettings.ApplianceName)
	}
}

func TestSetupRejectsSecondRun(t *testing.T) {
	application, _, _ := newTestApp(t, config.Config{
		DefaultScanPorts: []int{22, 80},
	})

	if err := application.Setup(context.Background(), domain.SetupInput{
		ApplianceName:    "Rack Alpha",
		DefaultScanPorts: []int{22, 80},
	}); err != nil {
		t.Fatalf("first setup: %v", err)
	}

	err := application.Setup(context.Background(), domain.SetupInput{
		ApplianceName:    "Rack Beta",
		DefaultScanPorts: []int{22, 80},
	})
	if err == nil {
		t.Fatalf("expected second setup attempt to fail")
	}
}

func TestCreateAPITokenSupportsScopedValidation(t *testing.T) {
	application, _, _ := newTestApp(t, config.Config{
		DefaultScanPorts: []int{22, 80},
	})

	if err := application.Setup(context.Background(), domain.SetupInput{
		ApplianceName:    "Rack Alpha",
		DefaultScanPorts: []int{22, 80},
	}); err != nil {
		t.Fatalf("setup workspace: %v", err)
	}

	created, err := application.CreateAPIToken(context.Background(), domain.CreateAPITokenInput{
		Name:  "Read only",
		Scope: domain.TokenScopeRead,
	})
	if err != nil {
		t.Fatalf("create api token: %v", err)
	}
	if created.Secret == "" {
		t.Fatalf("expected raw api token secret")
	}

	ok, err := application.ValidateAPIToken(context.Background(), created.Secret, domain.TokenScopeRead)
	if err != nil {
		t.Fatalf("validate read token: %v", err)
	}
	if !ok {
		t.Fatalf("expected read token to validate for read scope")
	}

	ok, err = application.ValidateAPIToken(context.Background(), created.Secret, domain.TokenScopeWrite)
	if err != nil {
		t.Fatalf("validate write scope with read token: %v", err)
	}
	if ok {
		t.Fatalf("expected read token to fail write scope validation")
	}
}

func newTestApp(t *testing.T, overrides config.Config) (*App, *sqlite.Store, config.Config) {
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
	if len(overrides.DefaultScanPorts) > 0 {
		cfg.DefaultScanPorts = append([]int(nil), overrides.DefaultScanPorts...)
	}
	if len(overrides.TrustedCIDRs) > 0 {
		cfg.TrustedCIDRs = append([]string(nil), overrides.TrustedCIDRs...)
	}

	store, err := sqlite.New(cfg.DBPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	return New(cfg, store, events.NewBus()), store, cfg
}
