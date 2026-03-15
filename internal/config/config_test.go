package config

import "testing"

func TestLoadUsesDefaultTrustedCIDRs(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if len(cfg.TrustedCIDRs) == 0 {
		t.Fatalf("expected default trusted cidrs")
	}
}
