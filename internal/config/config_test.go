package config

import (
	"path/filepath"
	"testing"
)

func TestLoadDerivesAdminTokenFileFromDataDir(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("HOMELABWATCH_DATA_DIR", dataDir)
	t.Setenv("HOMELABWATCH_DB_PATH", filepath.Join(dataDir, "homelabwatch.db"))
	t.Setenv("HOMELABWATCH_ADMIN_TOKEN_FILE", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	expected := filepath.Join(dataDir, "admin-token")
	if cfg.AdminTokenFile != expected {
		t.Fatalf("expected admin token file %q, got %q", expected, cfg.AdminTokenFile)
	}
}
