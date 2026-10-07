package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/karabas/yakamoz/internal/app"
	"github.com/karabas/yakamoz/internal/config"
	"github.com/karabas/yakamoz/internal/platform/logging"
	"github.com/karabas/yakamoz/internal/platform/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger := logging.New()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var handler http.Handler
	if cfg.Store == "postgres" {
		pool, err := postgres.Open(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatal(err)
		}
		defer pool.Close()
		if err := postgres.Migrate(ctx, pool, "migrations"); err != nil {
			log.Fatal(err)
		}
		handler = app.NewPostgres(ctx, cfg, logger, pool)
	} else {
		handler = app.NewMemory(cfg, logger)
	}

	server := &http.Server{Addr: cfg.Addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		logger.Info("server listening", "addr", cfg.Addr, "store", cfg.Store)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		logger.Error("server shutdown failed", "error", err)
	}
}
