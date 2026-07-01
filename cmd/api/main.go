// Command api is the GoTunnels backend entrypoint. It loads central
// configuration, initialises telemetry, connects to Postgres, applies
// migrations, wires the HTTP server, and runs until interrupted.
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

	"github.com/kusl/GoTunnels/internal/activity"
	"github.com/kusl/GoTunnels/internal/auth"
	"github.com/kusl/GoTunnels/internal/config"
	"github.com/kusl/GoTunnels/internal/csp"
	"github.com/kusl/GoTunnels/internal/database"
	"github.com/kusl/GoTunnels/internal/health"
	"github.com/kusl/GoTunnels/internal/httpx"
	"github.com/kusl/GoTunnels/internal/server"
	"github.com/kusl/GoTunnels/internal/store"
	"github.com/kusl/GoTunnels/internal/telemetry"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := run(); err != nil {
		// Fall back to the standard logger if telemetry never came up.
		slog.Error("fatal", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if version != "dev" {
		cfg.Version = version
	}

	// Telemetry (logger is always usable, even on exporter failure).
	tel, err := telemetry.Setup(rootCtx, cfg)
	if err != nil {
		return err
	}
	log := tel.Logger
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tel.Shutdown(shutdownCtx)
	}()

	log.Info("starting gotunnels-api",
		slog.String("instance_id", cfg.InstanceID),
		slog.String("version", cfg.Version),
		slog.String("addr", cfg.HTTPAddr),
		slog.Bool("telemetry", cfg.Telemetry.Enabled),
		slog.String("csp_mode", cfg.CSPMode),
		slog.String("rp_id", cfg.RPID),
	)

	// Database + migrations.
	pool, err := database.Connect(rootCtx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	applied, err := database.Migrate(rootCtx, pool)
	if err != nil {
		return err
	}
	log.Info("migrations applied", slog.Int("count", len(applied)), slog.Any("versions", applied))

	// Wiring.
	st := store.New(pool)
	rec := activity.NewRecorder(st, cfg.IPHashPepper())

	wa, err := auth.NewWebAuthn(cfg.RPID, cfg.RPDisplayName, cfg.RPOrigins)
	if err != nil {
		return err
	}

	authHandlers := auth.NewHandlers(st, wa, rec, log, auth.Settings{
		SessionTTL:   cfg.SessionTTL,
		TOTPKey:      cfg.TOTPAESKey(),
		Issuer:       cfg.RPDisplayName,
		CookieName:   cfg.SessionCookieName,
		CookieSecure: !cfg.Dev,
	})

	healthHandler := health.NewHandler(st, log, health.Info{
		Service:     cfg.ServiceName,
		InstanceID:  cfg.InstanceID,
		Version:     cfg.Version,
		CSPMode:     cfg.CSPMode,
		CSPPolicy:   cfg.CSPPolicy,
		TelemetryOn: cfg.Telemetry.Enabled,
	})

	cspHandler := csp.NewHandler(st, log, cfg.IPHashPepper())

	srv := server.New(server.Deps{
		Config:         cfg,
		Log:            log,
		Auth:           authHandlers,
		Health:         healthHandler,
		CSP:            cspHandler,
		CSPRateLimiter: httpx.NewRateLimiter(5, 20), // 5 rps, burst 20, per hashed IP
		Pepper:         cfg.IPHashPepper(),
	})

	// Run the server and wait for either a serve error or a shutdown signal.
	serveErr := make(chan error, 1)
	go func() {
		log.Info("http server listening", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case err := <-serveErr:
		return err
	case <-rootCtx.Done():
		log.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", slog.String("error", err.Error()))
		return err
	}
	log.Info("shutdown complete")
	return nil
}
