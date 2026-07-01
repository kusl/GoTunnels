// Package server assembles the HTTP routes and middleware into a ready-to-run
// *http.Server. Route patterns use the Go 1.22 method-aware ServeMux, so no
// third-party router is needed.
package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/kusl/GoTunnels/internal/activity"
	"github.com/kusl/GoTunnels/internal/auth"
	"github.com/kusl/GoTunnels/internal/config"
	"github.com/kusl/GoTunnels/internal/csp"
	"github.com/kusl/GoTunnels/internal/health"
	"github.com/kusl/GoTunnels/internal/httpx"
)

// Deps are the wired dependencies the server needs.
type Deps struct {
	Config         *config.Config
	Log            *slog.Logger
	Auth           *auth.Handlers
	Health         *health.Handler
	CSP            *csp.Handler
	CSPRateLimiter *httpx.RateLimiter
	Pepper         []byte
}

// New builds the configured *http.Server.
func New(d Deps) *http.Server {
	mux := http.NewServeMux()

	// --- unauthenticated: health & info ---
	mux.HandleFunc("GET /healthz", d.Health.Live)
	mux.HandleFunc("GET /readyz", d.Health.Ready)
	mux.HandleFunc("GET /api/info", d.Health.InfoHandler)

	// --- unauthenticated: auth entry points ---
	mux.HandleFunc("POST /api/signup", d.Auth.Signup)
	mux.HandleFunc("POST /api/login", d.Auth.Login)
	mux.HandleFunc("POST /api/passkey/login/begin", d.Auth.PasskeyLoginBegin)
	mux.HandleFunc("POST /api/passkey/login/finish", d.Auth.PasskeyLoginFinish)

	// --- unauthenticated but rate-limited: CSP violation reports ---
	cspChain := httpx.Chain(
		http.HandlerFunc(d.CSP.ServeHTTP),
		d.CSPRateLimiter.LimitByIP(cspRateKey(d.Pepper)),
	)
	mux.Handle("POST /api/csp-reports", cspChain)

	// --- authenticated ---
	authed := d.Auth.RequireAuth
	mux.Handle("POST /api/logout", authed(http.HandlerFunc(d.Auth.Logout)))
	mux.Handle("GET /api/me", authed(http.HandlerFunc(d.Auth.Me)))
	mux.Handle("GET /api/activity", authed(http.HandlerFunc(d.Auth.Activity)))
	mux.Handle("POST /api/passkey/register/begin", authed(http.HandlerFunc(d.Auth.PasskeyRegisterBegin)))
	mux.Handle("POST /api/passkey/register/finish", authed(http.HandlerFunc(d.Auth.PasskeyRegisterFinish)))
	mux.Handle("POST /api/totp/enroll", authed(http.HandlerFunc(d.Auth.TOTPEnroll)))
	mux.Handle("POST /api/totp/confirm", authed(http.HandlerFunc(d.Auth.TOTPConfirm)))
	mux.Handle("POST /api/totp/disable", authed(http.HandlerFunc(d.Auth.TOTPDisable)))

	// Global middleware (outermost first). CORS answers OPTIONS preflight
	// before requests reach the mux, so method-specific routes never 405 on
	// preflight.
	handler := httpx.Chain(mux,
		httpx.RequestID(),
		httpx.Recoverer(d.Log),
		httpx.CORS(d.Config.CORSAllowedOrigins, d.Config.CORSAllowsAny()),
	)
	handler = httpx.Instrument(handler, "gotunnels-api")

	return &http.Server{
		Addr:              d.Config.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}

// cspRateKey derives a rate-limit key from the hashed client IP so the CSP
// endpoint cannot be trivially flooded, while still never storing a raw IP.
func cspRateKey(pepper []byte) func(*http.Request) string {
	return func(r *http.Request) string {
		return activity.HashIP(pepper, activity.ClientIP(r))
	}
}
