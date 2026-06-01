// Command api is the entrypoint for the billing backend HTTP server.
//
// @title                      Billing Hotspot API
// @version                    1.0
// @description                Backend billing, package & voucher API for a FreeRADIUS/Mikrotik hotspot. Provisions credentials via the radius-api and accepts payments via Xendit, Midtrans and Tripay.
// @BasePath                   /api/v1
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
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

	"github.com/dirhamt/billing-hotspot/backend/internal/config"
	"github.com/dirhamt/billing-hotspot/backend/internal/database"
	"github.com/dirhamt/billing-hotspot/backend/internal/server"
	"github.com/dirhamt/billing-hotspot/backend/internal/validatorx"
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

	if err := validatorx.Setup(); err != nil {
		slog.Error("failed to set up validator", slog.Any("error", err))
		os.Exit(1)
	}

	engine := server.New(cfg, db)
	srv := &http.Server{
		Addr:              ":" + cfg.App.Port,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("server listening", slog.String("addr", srv.Addr), slog.String("env", cfg.App.Env))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	slog.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", slog.Any("error", err))
	}
}

func setupLogger(cfg *config.Config) {
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	if cfg.App.IsProduction() {
		opts.Level = slog.LevelInfo
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(handler))
}
