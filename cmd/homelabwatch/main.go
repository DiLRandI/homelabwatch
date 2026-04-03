package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apihttp "github.com/deleema/homelabwatch/internal/api/http"
	"github.com/deleema/homelabwatch/internal/app"
	"github.com/deleema/homelabwatch/internal/config"
	"github.com/deleema/homelabwatch/internal/events"
	"github.com/deleema/homelabwatch/internal/logging"
	"github.com/deleema/homelabwatch/internal/store/sqlite"
)

func main() {
	logger, logCfg := logging.NewFromEnv()
	if logCfg.InvalidValue != "" {
		logger.Warn("invalid LOG_LEVEL; defaulting to info", "value", logCfg.InvalidValue)
	}

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "err", err)
		os.Exit(1)
	}
	store, err := sqlite.New(cfg.DBPath)
	if err != nil {
		logger.Error("failed to open store", "path", cfg.DBPath, "err", err)
		os.Exit(1)
	}
	bus := events.NewBus()
	application := app.New(cfg, store, bus, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	application.Start(ctx)

	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           apihttp.NewRouter(application, cfg),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("homelabwatch listening", "addr", cfg.ListenAddr, "log_level", logCfg.Level.String())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown requested")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Warn("http server shutdown failed", "err", err)
	}
	if err := store.Close(); err != nil {
		logger.Warn("store close failed", "err", err)
	}
	logger.Info("shutdown complete")
}
