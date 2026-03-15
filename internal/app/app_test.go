package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deleema/homelabwatch/internal/config"
	"github.com/deleema/homelabwatch/internal/events"
	"github.com/deleema/homelabwatch/internal/store/sqlite"
)

func TestEnsureBootstrapAutoGeneratesToken(t *testing.T) {
	application, store, cfg := newTestApp(t, config.Config{
		AutoBootstrap:    true,
		DefaultScanPorts: []int{22, 80},
	})

	result, err := application.EnsureBootstrap(context.Background())
	if err != nil {
		t.Fatalf("ensure bootstrap: %v", err)
	}
	if !result.Performed {
		t.Fatalf("expected bootstrap to run")
	}
	if !result.Generated {
		t.Fatalf("expected token generation on first bootstrap")
	}
	if result.AdminToken == "" {
		t.Fatalf("expected generated admin token")
	}
	if result.AdminTokenFile != cfg.AdminTokenFile {
		t.Fatalf("expected token file %q, got %q", cfg.AdminTokenFile, result.AdminTokenFile)
	}

	status, err := store.BootstrapStatus(context.Background())
	if err != nil {
		t.Fatalf("bootstrap status: %v", err)
	}
	if !status.Initialized {
		t.Fatalf("expected initialized store")
	}

	ok, err := application.ValidateAdminToken(context.Background(), result.AdminToken)
	if err != nil {
		t.Fatalf("validate admin token: %v", err)
	}
	if !ok {
		t.Fatalf("expected generated token to validate")
	}

	content, err := os.ReadFile(cfg.AdminTokenFile)
	if err != nil {
		t.Fatalf("read admin token file: %v", err)
	}
	if strings.TrimSpace(string(content)) != result.AdminToken {
		t.Fatalf("unexpected admin token file contents")
	}
}

func TestEnsureBootstrapUsesConfiguredAdminToken(t *testing.T) {
	application, _, cfg := newTestApp(t, config.Config{
		AutoBootstrap:    true,
		AdminToken:       "preset-token",
		DefaultScanPorts: []int{22, 80},
	})

	result, err := application.EnsureBootstrap(context.Background())
	if err != nil {
		t.Fatalf("ensure bootstrap: %v", err)
	}
	if !result.Performed {
		t.Fatalf("expected bootstrap to run")
	}
	if result.Generated {
		t.Fatalf("did not expect generated token when config token is set")
	}
	if result.AdminToken != "preset-token" {
		t.Fatalf("expected configured token to be used, got %q", result.AdminToken)
	}

	content, err := os.ReadFile(cfg.AdminTokenFile)
	if err != nil {
		t.Fatalf("read admin token file: %v", err)
	}
	if strings.TrimSpace(string(content)) != "preset-token" {
		t.Fatalf("unexpected admin token file contents")
	}
}

func TestEnsureBootstrapSkipsExistingState(t *testing.T) {
	application, _, _ := newTestApp(t, config.Config{
		AutoBootstrap:    true,
		AdminToken:       "first-token",
		DefaultScanPorts: []int{22, 80},
	})

	result, err := application.EnsureBootstrap(context.Background())
	if err != nil {
		t.Fatalf("first ensure bootstrap: %v", err)
	}
	if !result.Performed {
		t.Fatalf("expected initial bootstrap to run")
	}

	next, err := application.EnsureBootstrap(context.Background())
	if err != nil {
		t.Fatalf("second ensure bootstrap: %v", err)
	}
	if next.Performed {
		t.Fatalf("did not expect bootstrap to rerun on existing state")
	}

	ok, err := application.ValidateAdminToken(context.Background(), "first-token")
	if err != nil {
		t.Fatalf("validate admin token: %v", err)
	}
	if !ok {
		t.Fatalf("expected existing token to remain valid")
	}
}

func TestEnsureBootstrapRespectsDisableFlag(t *testing.T) {
	application, store, _ := newTestApp(t, config.Config{
		AutoBootstrap:    false,
		DefaultScanPorts: []int{22, 80},
	})

	result, err := application.EnsureBootstrap(context.Background())
	if err != nil {
		t.Fatalf("ensure bootstrap: %v", err)
	}
	if result.Performed {
		t.Fatalf("did not expect bootstrap to run when disabled")
	}

	status, err := store.BootstrapStatus(context.Background())
	if err != nil {
		t.Fatalf("bootstrap status: %v", err)
	}
	if status.Initialized {
		t.Fatalf("expected uninitialized state when auto-bootstrap is disabled")
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
		AutoBootstrap:    true,
		DefaultScanPorts: []int{22, 80, 443},
		AdminTokenFile:   filepath.Join(dataDir, "admin-token"),
	}
	if overrides.AutoBootstrap != cfg.AutoBootstrap {
		cfg.AutoBootstrap = overrides.AutoBootstrap
	}
	if len(overrides.DefaultScanPorts) > 0 {
		cfg.DefaultScanPorts = append([]int(nil), overrides.DefaultScanPorts...)
	}
	if overrides.AdminToken != "" {
		cfg.AdminToken = overrides.AdminToken
	}
	if overrides.AdminTokenFile != "" {
		cfg.AdminTokenFile = overrides.AdminTokenFile
	}

	store, err := sqlite.New(cfg.DBPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	return New(cfg, store, events.NewBus()), store, cfg
}
