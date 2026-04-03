package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

const levelEnv = "LOG_LEVEL"

type Config struct {
	Level        slog.Level
	InvalidValue string
}

func ConfigFromEnv(getenv func(string) string) Config {
	raw := strings.TrimSpace(getenv(levelEnv))
	if raw == "" {
		return Config{Level: slog.LevelInfo}
	}

	switch strings.ToLower(raw) {
	case "debug":
		return Config{Level: slog.LevelDebug}
	case "info":
		return Config{Level: slog.LevelInfo}
	case "warn", "warning":
		return Config{Level: slog.LevelWarn}
	case "error":
		return Config{Level: slog.LevelError}
	default:
		return Config{Level: slog.LevelInfo, InvalidValue: raw}
	}
}

func NewLogger(cfg Config, writer io.Writer) *slog.Logger {
	if writer == nil {
		writer = os.Stdout
	}
	return slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: cfg.Level}))
}

func NewFromEnv() (*slog.Logger, Config) {
	cfg := ConfigFromEnv(os.Getenv)
	logger := NewLogger(cfg, os.Stdout)
	slog.SetDefault(logger)
	return logger, cfg
}
