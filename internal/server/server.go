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
	"github.com/kusl/GoTunnels/internal/captcha"
	"github.com/kusl/GoTunnels/internal/config"
	"github.com/kusl/GoTunnels/internal/csp"
	"github.com/kusl/GoTunnels/internal/health"
	"github.com/kusl/GoTunnels/internal/httpx"
	"github.com/kusl/GoTunnels/internal/notes"
	"github.com/kusl/GoTunnels/internal/prefs"
)

// Deps are the wired dependencies the server needs.
type Deps struct {
	Config           *config.Config
	Log              *slog.Logger
	Auth             *auth.Handlers
	Health           *health.Handler
	CSP              *csp.Handler
	CSPRateLimiter   *httpx.RateLimiter
	Captcha          *captcha.Handlers
	Notes            *notes.Handlers
	Prefs            *prefs.Handlers
	NotesRateLimiter *httpx.RateLimiter
	// SignupIPLimiter caps account creation per client IP; SignupGlobalLimiter
	// caps it across everyone. Either may be nil to disable that guard (the
	// config maps a zero interval to nil). They apply to the moment an
	// account is actually created: POST /api/signup and the passkey signup
	// *finish* — begin is deliberately unguarded so an abandoned or failed
	// authenticator ceremony never burns a scarce signup token.
	SignupIPLimiter     *httpx.RateLimiter
	SignupGlobalLimiter *httpx.RateLimiter
	Pepper              []byte
}

// New builds the configured *http.Server.
func New(d Deps) *http.Server {
	mux := http.NewServeMux()

	// --- unauthenticated: health & info ---
	mux.HandleFunc("GET /healthz", d.Health.Live)
	mux.HandleFunc("GET /readyz", d.Health.Ready)
	mux.HandleFunc("GET /api/info", d.Health.InfoHandler)

	// --- unauthenticated: auth entry points ---
	mux.Handle("POST /api/signup", d.signupGuard(http.HandlerFunc(d.Auth.Signup)))
	mux.HandleFunc("POST /api/login", d.Auth.Login)
	mux.HandleFunc("POST /api/passkey/login/begin", d.Auth.PasskeyLoginBegin)
	mux.HandleFunc("POST /api/passkey/login/finish", d.Auth.PasskeyLoginFinish)
	mux.HandleFunc("POST /api/passkey/signup/begin", d.Auth.PasskeySignupBegin)
	mux.Handle("POST /api/passkey/signup/finish", d.signupGuard(http.HandlerFunc(d.Auth.PasskeySignupFinish)))

	// --- unauthenticated but rate-limited: CSP violation reports ---
	mux.Handle("POST /api/csp-reports", httpx.Chain(
		http.HandlerFunc(d.CSP.ServeHTTP),
		d.CSPRateLimiter.LimitByIP(hashedIPKey(d.Pepper)),
	))
	// Public, sanitised read side of the same data: the passkeys/security
	// explainer page renders it so visitors can watch CSP work. Shares the
	// CSP limiter so the pair cannot be used to hammer the database.
	mux.Handle("GET /api/csp-reports/recent", httpx.Chain(
		http.HandlerFunc(d.CSP.Recent),
		d.CSPRateLimiter.LimitByIP(hashedIPKey(d.Pepper)),
	))

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

	// --- authenticated: CAPTCHA stats & leaderboard ---
	mux.Handle("GET /api/captcha/stats", authed(http.HandlerFunc(d.Captcha.Stats)))
	mux.Handle("POST /api/captcha/sync", authed(http.HandlerFunc(d.Captcha.Sync)))
	mux.Handle("POST /api/captcha/reset", authed(http.HandlerFunc(d.Captcha.Reset)))
	mux.Handle("GET /api/captcha/leaderboard", authed(http.HandlerFunc(d.Captcha.Leaderboard)))

	// --- authenticated: per-user preferences ---
	mux.Handle("GET /api/prefs/{key}", authed(http.HandlerFunc(d.Prefs.Get)))
	mux.Handle("PUT /api/prefs/{key}", authed(http.HandlerFunc(d.Prefs.Set)))

	// --- authenticated: notes (plain-text microblog) ---
	// POST is rate-limited per user so one account cannot flood the shared
	// feed; the limiter sits inside RequireAuth so the key derives from the
	// verified user, not from anything the client controls.
	mux.Handle("GET /api/notes", authed(http.HandlerFunc(d.Notes.List)))
	mux.Handle("GET /api/notes/authors", authed(http.HandlerFunc(d.Notes.Authors)))
	mux.Handle("POST /api/notes", authed(httpx.Chain(
		http.HandlerFunc(d.Notes.Create),
		d.NotesRateLimiter.LimitByIP(noteRateKey()),
	)))
	mux.Handle("DELETE /api/notes/{id}", authed(http.HandlerFunc(d.Notes.Delete)))

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

// signupGuard wraps an account-creation handler with the configured signup
// rate limits. Order matters: the per-IP guard runs first so one noisy client
// is rejected before it can drain the global bucket everyone shares; the
// global guard runs second as the site-wide backstop.
func (d Deps) signupGuard(h http.Handler) http.Handler {
	var mws []httpx.Middleware
	if d.SignupIPLimiter != nil {
		mws = append(mws, d.SignupIPLimiter.LimitWith(hashedIPKey(d.Pepper),
			"too many signups from your network — please try again in a few minutes"))
	}
	if d.SignupGlobalLimiter != nil {
		mws = append(mws, d.SignupGlobalLimiter.LimitWith(
			func(*http.Request) string { return "global" },
			"signups are briefly rate limited site-wide — please try again in a minute"))
	}
	return httpx.Chain(h, mws...)
}

// hashedIPKey derives a rate-limit key from the hashed client IP, so limits
// track clients without ever storing a raw IP.
func hashedIPKey(pepper []byte) func(*http.Request) string {
	return func(r *http.Request) string {
		return activity.HashIP(pepper, activity.ClientIP(r))
	}
}

// noteRateKey keys the note-creation rate limiter by the authenticated user
// ID. The limiter runs inside RequireAuth, so the user is always present; the
// "anon" fallback only guards against future wiring mistakes.
func noteRateKey() func(*http.Request) string {
	return func(r *http.Request) string {
		if u, ok := auth.CurrentUser(r.Context()); ok {
			return "user:" + u.ID
		}
		return "anon"
	}
}
