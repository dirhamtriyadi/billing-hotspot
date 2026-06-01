// Command api is the entrypoint for the radius-api management service.
//
// @title                     Radius API
// @version                   1.0
// @description               Management API over the FreeRADIUS SQL schema: profiles, users, sessions, NAS clients and CoA disconnect.
// @BasePath                  /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in                        header
// @name                      X-API-Key
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dirhamt/billing-hotspot/radius-api/internal/config"
	"github.com/dirhamt/billing-hotspot/radius-api/internal/database"
	"github.com/dirhamt/billing-hotspot/radius-api/internal/server"
)

func main() {
	cfg := config.Load()
	setupLogger(cfg)

	db, err := database.Connect(cfg.DB, cfg.App.IsProduction())
	if err != nil {
		slog.Error("failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("database connected")

	if err := database.Migrate(db); err != nil {
		slog.Error("failed to run migrations", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("migrations applied")

	engine := server.New(cfg, db)
	srv := &http.Server{
		Addr:              ":" + cfg.App.Port,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("radius-api listening", slog.String("addr", srv.Addr), slog.String("env", cfg.App.Env))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	slog.Info("shutting down radius-api")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", slog.Any("error", err))
	}
}

func setupLogger(cfg *config.Config) {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	var handler slog.Handler
	if cfg.App.IsProduction() {
		opts.Level = slog.LevelInfo
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(handler))
}
