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
	"github.com/kusl/GoTunnels/internal/captcha"
	"github.com/kusl/GoTunnels/internal/config"
	"github.com/kusl/GoTunnels/internal/csp"
	"github.com/kusl/GoTunnels/internal/database"
	"github.com/kusl/GoTunnels/internal/health"
	"github.com/kusl/GoTunnels/internal/httpx"
	"github.com/kusl/GoTunnels/internal/maintenance"
	"github.com/kusl/GoTunnels/internal/notes"
	"github.com/kusl/GoTunnels/internal/prefs"
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
	slog.SetDefault(log)
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

	// The relying-party provider validates the static RP config at boot and
	// additionally derives per-origin relying parties for origins matching
	// cfg.RPOriginPatterns — the fix for passkeys breaking when the tunnel
	// hostname the browser sees differs from the one configured at startup.
	rp, err := auth.NewRPProvider(auth.RPConfig{
		RPID:           cfg.RPID,
		RPDisplayName:  cfg.RPDisplayName,
		RPOrigins:      cfg.RPOrigins,
		OriginPatterns: cfg.RPOriginPatterns,
	})
	if err != nil {
		return err
	}

	authHandlers := auth.NewHandlers(st, rp, rec, log, auth.Settings{
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

	captchaHandlers := captcha.NewHandlers(st, log)
	notesHandlers := notes.NewHandlers(st, log)
	prefsHandlers := prefs.NewHandlers(st, log)

	signupIP := intervalLimiter(cfg.SignupIPInterval, cfg.SignupIPBurst)
	signupGlobal := intervalLimiter(cfg.SignupGlobalInterval, cfg.SignupGlobalBurst)
	authLimiter := intervalLimiter(cfg.AuthAttemptInterval, cfg.AuthAttemptBurst)

	janitorCtx, stopJanitor := context.WithCancel(rootCtx)
	defer stopJanitor()
	janitor := maintenance.New(st, log, maintenance.Config{
		Interval:           cfg.MaintenanceInterval,
		HealthLogRetention: cfg.HealthLogRetention,
		CSPReportRetention: cfg.CSPReportRetention,
	})
	janitorDone := make(chan struct{})
	go func() {
		defer close(janitorDone)
		janitor.Run(janitorCtx)
	}()
	defer func() {
		stopJanitor()
		<-janitorDone
	}()

	srv := server.New(server.Deps{
		Config:         cfg,
		Log:            log,
		Auth:           authHandlers,
		Health:         healthHandler,
		CSP:            cspHandler,
		CSPRateLimiter: httpx.NewRateLimiter(5, 20), // 5 rps, burst 20, per hashed IP
		Captcha:        captchaHandlers,
		Notes:          notesHandlers,
		Prefs:          prefsHandlers,
		// One note every 2 seconds sustained, short bursts of 5, per user.
		NotesRateLimiter:    httpx.NewRateLimiter(0.5, 5),
		SignupIPLimiter:     signupIP,
		SignupGlobalLimiter: signupGlobal,
		AuthLimiter:         authLimiter,
		Pepper:              cfg.IPHashPepper(),
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

// intervalLimiter builds a "one every interval, with burst" limiter, or nil
// (guard disabled) when the interval is zero.
func intervalLimiter(interval time.Duration, burst int) *httpx.RateLimiter {
	if interval <= 0 {
		return nil
	}
	return httpx.NewRateLimiter(config.PerInterval(interval), float64(burst))
}
