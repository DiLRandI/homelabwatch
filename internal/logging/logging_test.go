package logging

import (
	"log/slog"
	"testing"
)

func TestConfigFromEnvDefaultsToInfo(t *testing.T) {
	cfg := ConfigFromEnv(func(string) string { return "" })
	if cfg.Level != slog.LevelInfo {
		t.Fatalf("expected info level, got %s", cfg.Level.String())
	}
	if cfg.InvalidValue != "" {
		t.Fatalf("expected no invalid value, got %q", cfg.InvalidValue)
	}
}

func TestConfigFromEnvParsesSupportedLevels(t *testing.T) {
	tests := []struct {
		name  string
		value string
		level slog.Level
	}{
		{name: "debug", value: "debug", level: slog.LevelDebug},
		{name: "info", value: "INFO", level: slog.LevelInfo},
		{name: "warn", value: "warn", level: slog.LevelWarn},
		{name: "warning", value: "warning", level: slog.LevelWarn},
		{name: "error", value: "error", level: slog.LevelError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := ConfigFromEnv(func(string) string { return test.value })
			if cfg.Level != test.level {
				t.Fatalf("expected %s, got %s", test.level.String(), cfg.Level.String())
			}
			if cfg.InvalidValue != "" {
				t.Fatalf("expected no invalid value, got %q", cfg.InvalidValue)
			}
		})
	}
}

func TestConfigFromEnvFallsBackToInfoForInvalidValue(t *testing.T) {
	cfg := ConfigFromEnv(func(string) string { return "trace" })
	if cfg.Level != slog.LevelInfo {
		t.Fatalf("expected info level, got %s", cfg.Level.String())
	}
	if cfg.InvalidValue != "trace" {
		t.Fatalf("expected invalid value to be recorded, got %q", cfg.InvalidValue)
	}
}
