package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apihttp "github.com/deleema/homelabwatch/internal/api/http"
	"github.com/deleema/homelabwatch/internal/app"
	"github.com/deleema/homelabwatch/internal/config"
	"github.com/deleema/homelabwatch/internal/events"
	"github.com/deleema/homelabwatch/internal/store/sqlite"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	store, err := sqlite.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to open store: %v", err)
	}
	bus := events.NewBus()
	application := app.New(cfg, store, bus)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	application.Start(ctx)

	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           apihttp.NewRouter(application, cfg),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("homelabwatch listening on %s", cfg.ListenAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("http server failed: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
	_ = store.Close()
}
