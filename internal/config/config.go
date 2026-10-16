// Package config is the single source of truth for every runtime setting.
//
// Everything the application can be tuned with is read here, from environment
// variables, and nowhere else. No other package reads os.Getenv. This keeps
// the "all configuration is central in one location" promise from the design
// notes: to see everything GoTunnels can be configured with, read this file.
package config

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all resolved runtime settings for the API service.
type Config struct {
	// Identity -----------------------------------------------------------
	InstanceID  string // unique per running instance; also an OTel resource attribute
	ServiceName string // OTel service.name
	Version     string // build version (set via -ldflags at build time)

	// HTTP ---------------------------------------------------------------
	HTTPAddr        string        // listen address, e.g. ":8080"
	ShutdownTimeout time.Duration // graceful shutdown budget

	// Database -----------------------------------------------------------
	DatabaseURL      string
	DBMaxConns       int32
	DBMinConns       int32
	DBConnectTimeout time.Duration

	// Sessions -----------------------------------------------------------
	// SessionTTL is how long a session stays valid after it is issued. A value
	// of 0 (the default) means sessions never expire on their own: they live
	// until the user explicitly logs out (here or "everywhere"). GoTunnels
	// never logs anyone out for inactivity — this is a deliberate product
	// choice (see docs/ARCHITECTURE.md). Set a positive duration to opt into
	// expiring sessions instead.
	SessionCookieName string
	SessionTTL        time.Duration

	// Secrets (generated per-instance by scripts/up.sh, never committed) -
	ipHashPepper []byte   // used as sha256(pepper || ip)
	totpAESKey   [32]byte // AES-256-GCM key for encrypting TOTP secrets at rest

	// CORS ---------------------------------------------------------------
	// Exact allowed origins. Because the browser talks to the API cross-origin
	// (frontend tunnel URL != API tunnel URL) and sends credentials, we cannot
	// use the "*" wildcard together with credentials; we echo an allowed origin.
	CORSAllowedOrigins []string

	// WebAuthn / passkeys -----------------------------------------------
	RPID          string   // relying-party ID: the frontend's registrable domain
	RPDisplayName string   // human-readable name shown by the authenticator
	RPOrigins     []string // full origins the browser will present, e.g. https://x.trycloudflare.com
	// RPOriginPatterns lets the WebAuthn relying party be derived per request
	// from the browser's Origin header when it matches one of these wildcard
	// patterns (e.g. "https://*.trycloudflare.com"). Quick Tunnel hostnames
	// rotate, and container env can go stale (see docs/ARCHITECTURE.md), so
	// this makes passkeys self-healing instead of frozen at boot time. The
	// patterns gate only the WebAuthn RP derivation — they are deliberately
	// NOT added to the CORS allow-list, because CORS + credentials must stay
	// pinned to the one real frontend origin.
	RPOriginPatterns []string

	// Signup rate limits --------------------------------------------------
	// Account creation is limited much harder than ordinary API traffic:
	// one signup per SignupIPInterval per client IP (hashed), and one per
	// SignupGlobalInterval across the whole instance. A zero interval
	// disables that limiter. Bursts allow small clusters (a classroom
	// behind one NAT can raise SignupIPBurst).
	SignupIPInterval     time.Duration
	SignupIPBurst        int
	SignupGlobalInterval time.Duration
	SignupGlobalBurst    int

	// Authentication attempt rate limit ----------------------------------
	// Password login, passkey login, and TOTP confirm/disable are limited to
	// one attempt per AuthAttemptInterval per client (hashed IP, or user for
	// authenticated TOTP calls), with AuthAttemptBurst of headroom. Brute
	// forcing a password or a 6-digit code becomes impractical. Zero disables.
	AuthAttemptInterval time.Duration
	AuthAttemptBurst    int

	// Maintenance ---------------------------------------------------------
	// A background janitor deletes expired WebAuthn ceremonies and prunes
	// health_check_log and csp_reports older than their retention. A zero
	// retention keeps that table forever.
	MaintenanceInterval time.Duration
	HealthLogRetention  time.Duration
	CSPReportRetention  time.Duration

	// Telemetry host identity ---------------------------------------------
	// HostName becomes the OTel resource attribute host.name. Inside a
	// container os.Hostname() is the container ID, so scripts/up.sh passes
	// the real machine's hostname through this variable. Empty means "let
	// the SDK detect it" (which yields the container hostname).
	HostName string

	// Content Security Policy -------------------------------------------
	// The header itself is emitted by Caddy on the frontend. These values let
	// the API report what mode it believes is active and are surfaced on the
	// health/info endpoint so the whole system agrees on one central policy.
	CSPMode   string // "report-only" or "enforce"
	CSPPolicy string // the policy string (informational mirror of Caddy's)

	// Telemetry ----------------------------------------------------------
	Telemetry TelemetryConfig

	// Dev conveniences ---------------------------------------------------
	Dev bool // when true, missing secrets are generated ephemerally with a warning
}

// TelemetryConfig captures OTLP/HTTP exporter settings. When Enabled is false
// the service logs to stdout only and installs no-op trace/metric providers.
type TelemetryConfig struct {
	Enabled     bool
	EndpointURL string            // full base URL, e.g. https://api.uptrace.dev
	Headers     map[string]string // e.g. {"uptrace-dsn": "..."}
	Insecure    bool              // allow http:// (in-cluster collectors)
	Compression string            // "gzip" or ""

	// MetricsTemporality mirrors OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE
	// ("delta", "cumulative", or "lowmemory"; normalized to lowercase). Uptrace
	// prefers delta temporality, so that is the default.
	MetricsTemporality string
	// MetricsHistogram mirrors OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION
	// ("base2_exponential_bucket_histogram" or "explicit_bucket_histogram";
	// normalized to lowercase). Exponential histograms compress better and give
	// Uptrace more accurate percentiles, so they are the default.
	MetricsHistogram string
}

// Load reads and validates configuration from the environment. A variable
// that is set but cannot be parsed is an error, never a silent fallback to the
// default: "GOTUNNELS_SESSION_TTL=7days" must not quietly mean "never".
func Load() (*Config, error) {
	e := &env{}
	c := &Config{
		InstanceID:           e.str("GOTUNNELS_INSTANCE_ID", defaultInstanceID()),
		ServiceName:          e.str("OTEL_SERVICE_NAME", e.str("GOTUNNELS_SERVICE_NAME", "gotunnels-api")),
		Version:              e.str("GOTUNNELS_VERSION", "dev"),
		HTTPAddr:             e.str("GOTUNNELS_HTTP_ADDR", ":8080"),
		ShutdownTimeout:      e.dur("GOTUNNELS_SHUTDOWN_TIMEOUT", 15*time.Second),
		DatabaseURL:          e.str("DATABASE_URL", ""),
		DBMaxConns:           e.int32("GOTUNNELS_DB_MAX_CONNS", 20),
		DBMinConns:           e.int32("GOTUNNELS_DB_MIN_CONNS", 2),
		DBConnectTimeout:     e.dur("GOTUNNELS_DB_CONNECT_TIMEOUT", 30*time.Second),
		SessionCookieName:    e.str("GOTUNNELS_SESSION_COOKIE_NAME", "gotunnels_session"),
		SessionTTL:           e.dur("GOTUNNELS_SESSION_TTL", 0),
		CORSAllowedOrigins:   splitList(e.str("GOTUNNELS_CORS_ALLOWED_ORIGINS", "*")),
		RPID:                 e.str("GOTUNNELS_RP_ID", "localhost"),
		RPDisplayName:        e.str("GOTUNNELS_RP_DISPLAY_NAME", "GoTunnels"),
		RPOrigins:            splitList(e.str("GOTUNNELS_RP_ORIGINS", "http://localhost:8080")),
		RPOriginPatterns:     splitList(e.str("GOTUNNELS_RP_ORIGIN_PATTERNS", "https://*.trycloudflare.com")),
		SignupIPInterval:     e.dur("GOTUNNELS_SIGNUP_IP_INTERVAL", 5*time.Minute),
		SignupIPBurst:        e.int("GOTUNNELS_SIGNUP_IP_BURST", 1),
		SignupGlobalInterval: e.dur("GOTUNNELS_SIGNUP_GLOBAL_INTERVAL", time.Minute),
		SignupGlobalBurst:    e.int("GOTUNNELS_SIGNUP_GLOBAL_BURST", 1),
		AuthAttemptInterval:  e.dur("GOTUNNELS_AUTH_ATTEMPT_INTERVAL", 6*time.Second),
		AuthAttemptBurst:     e.int("GOTUNNELS_AUTH_ATTEMPT_BURST", 10),
		MaintenanceInterval:  e.dur("GOTUNNELS_MAINTENANCE_INTERVAL", time.Hour),
		HealthLogRetention:   e.dur("GOTUNNELS_HEALTH_LOG_RETENTION", 7*24*time.Hour),
		CSPReportRetention:   e.dur("GOTUNNELS_CSP_REPORT_RETENTION", 30*24*time.Hour),
		HostName:             e.str("GOTUNNELS_HOST_NAME", ""),
		CSPMode:              e.str("GOTUNNELS_CSP_MODE", "report-only"),
		CSPPolicy:            e.str("GOTUNNELS_CSP_POLICY", DefaultCSPPolicy),
		Dev:                  e.bool("GOTUNNELS_DEV", false),
	}
	if err := e.err(); err != nil {
		return nil, err
	}

	if err := c.resolveSecrets(); err != nil {
		return nil, err
	}
	c.resolveTelemetry()

	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

// DefaultCSPPolicy locks everything to same-origin: no third-party scripts,
// styles, images, fonts, media, or frames. Everything is self-hosted. It is
// emitted in report-only mode by default (see the Caddyfile), so violations
// are reported but nothing is blocked yet.
const DefaultCSPPolicy = "default-src 'self'; " +
	"script-src 'self'; " +
	"style-src 'self'; " +
	"img-src 'self'; " +
	"font-src 'self'; " +
	"connect-src 'self' https:; " +
	"media-src 'self'; " +
	"object-src 'none'; " +
	"frame-src 'none'; " +
	"frame-ancestors 'none'; " +
	"base-uri 'self'; " +
	"form-action 'self'"

func (c *Config) resolveSecrets() error {
	// IP hashing pepper.
	if pepper := os.Getenv("GOTUNNELS_IP_HASH_PEPPER"); pepper != "" {
		c.ipHashPepper = []byte(pepper)
	} else if c.Dev {
		c.ipHashPepper = mustRandom(32)
	} else {
		return fmt.Errorf("config: GOTUNNELS_IP_HASH_PEPPER is required (set GOTUNNELS_DEV=1 to auto-generate for local dev)")
	}

	// TOTP encryption key. Accept any string; derive a fixed 32-byte AES key
	// via SHA-256 so operators can supply `openssl rand -base64 32` output
	// without worrying about exact byte length.
	if raw := os.Getenv("GOTUNNELS_TOTP_ENCRYPTION_KEY"); raw != "" {
		c.totpAESKey = sha256.Sum256([]byte(raw))
	} else if c.Dev {
		c.totpAESKey = sha256.Sum256(mustRandom(32))
	} else {
		return fmt.Errorf("config: GOTUNNELS_TOTP_ENCRYPTION_KEY is required (set GOTUNNELS_DEV=1 to auto-generate for local dev)")
	}
	return nil
}

func (c *Config) resolveTelemetry() {
	// Preference order:
	//   1. UPTRACE_DSN (convenience) -> derive endpoint + uptrace-dsn header
	//   2. OTEL_EXPORTER_OTLP_ENDPOINT (+ OTEL_EXPORTER_OTLP_HEADERS)
	//   3. disabled (stdout logging, no-op traces/metrics)
	//
	// The two metrics knobs follow the OpenTelemetry spec's environment
	// variables. Their spec values are UPPERCASE (e.g. DELTA,
	// BASE2_EXPONENTIAL_BUCKET_HISTOGRAM) but we normalize to lowercase so
	// comparisons elsewhere are simple and either casing works.
	tc := TelemetryConfig{
		Headers:     map[string]string{},
		Compression: strings.ToLower(envStr("OTEL_EXPORTER_OTLP_COMPRESSION", "gzip")),
		MetricsTemporality: strings.ToLower(envStr(
			"OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE", "delta")),
		MetricsHistogram: strings.ToLower(envStr(
			"OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION",
			"base2_exponential_bucket_histogram")),
	}

	if dsn := os.Getenv("UPTRACE_DSN"); dsn != "" {
		endpoint, insecure := endpointFromDSN(dsn)
		tc.Enabled = true
		tc.EndpointURL = endpoint
		tc.Insecure = insecure
		tc.Headers["uptrace-dsn"] = dsn
	} else if ep := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); ep != "" {
		tc.Enabled = true
		tc.EndpointURL = ep
		tc.Insecure = strings.HasPrefix(ep, "http://")
		for k, v := range parseHeaders(os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")) {
			tc.Headers[k] = v
		}
	}
	c.Telemetry = tc
}

// Validate checks invariants that must hold before the server starts.
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("config: DATABASE_URL is required")
	}
	if c.HTTPAddr == "" {
		return fmt.Errorf("config: GOTUNNELS_HTTP_ADDR is required")
	}
	if c.SessionTTL < 0 {
		return fmt.Errorf("config: GOTUNNELS_SESSION_TTL must not be negative (0 disables expiry — sessions live until logout)")
	}
	switch c.CSPMode {
	case "report-only", "enforce":
	default:
		return fmt.Errorf("config: GOTUNNELS_CSP_MODE must be 'report-only' or 'enforce', got %q", c.CSPMode)
	}
	if len(c.RPOrigins) == 0 {
		return fmt.Errorf("config: GOTUNNELS_RP_ORIGINS must contain at least one origin")
	}
	if c.SignupIPInterval < 0 || c.SignupGlobalInterval < 0 {
		return fmt.Errorf("config: signup rate-limit intervals must not be negative")
	}
	if c.SignupIPInterval > 0 && c.SignupIPBurst < 1 {
		return fmt.Errorf("config: GOTUNNELS_SIGNUP_IP_BURST must be >= 1 when the IP limiter is enabled")
	}
	if c.SignupGlobalInterval > 0 && c.SignupGlobalBurst < 1 {
		return fmt.Errorf("config: GOTUNNELS_SIGNUP_GLOBAL_BURST must be >= 1 when the global limiter is enabled")
	}
	if c.AuthAttemptInterval < 0 {
		return fmt.Errorf("config: GOTUNNELS_AUTH_ATTEMPT_INTERVAL must not be negative")
	}
	if c.AuthAttemptInterval > 0 && c.AuthAttemptBurst < 1 {
		return fmt.Errorf("config: GOTUNNELS_AUTH_ATTEMPT_BURST must be >= 1 when the auth limiter is enabled")
	}
	if c.DBMaxConns < 1 {
		return fmt.Errorf("config: GOTUNNELS_DB_MAX_CONNS must be >= 1")
	}
	if c.DBMinConns < 0 || c.DBMinConns > c.DBMaxConns {
		return fmt.Errorf("config: GOTUNNELS_DB_MIN_CONNS must be between 0 and GOTUNNELS_DB_MAX_CONNS")
	}
	if c.MaintenanceInterval <= 0 {
		return fmt.Errorf("config: GOTUNNELS_MAINTENANCE_INTERVAL must be positive")
	}
	if c.HealthLogRetention < 0 || c.CSPReportRetention < 0 {
		return fmt.Errorf("config: retention periods must not be negative (0 keeps rows forever)")
	}
	return nil
}

// PerInterval converts "one event per d" into a token-bucket rate in events
// per second. A non-positive interval yields 0, which callers treat as "this
// limiter is disabled".
func PerInterval(d time.Duration) float64 {
	if d <= 0 {
		return 0
	}
	return 1 / d.Seconds()
}

// IPHashPepper returns a copy-safe reference to the pepper bytes.
func (c *Config) IPHashPepper() []byte { return c.ipHashPepper }

// TOTPAESKey returns the 32-byte AES key for TOTP secret encryption.
func (c *Config) TOTPAESKey() [32]byte { return c.totpAESKey }

// CORSAllowsAny reports whether the wildcard "*" origin is configured.
func (c *Config) CORSAllowsAny() bool {
	for _, o := range c.CORSAllowedOrigins {
		if o == "*" {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// small env helpers
// ---------------------------------------------------------------------------

// env reads typed values and collects every parse failure so Load can report
// all misconfigurations at once.
type env struct {
	errs []error
}

func (e *env) err() error { return errors.Join(e.errs...) }

// raw returns the trimmed value and whether the variable is set to anything
// at all. A whitespace-only value counts as set, so it can deliberately clear
// a list setting such as GOTUNNELS_RP_ORIGIN_PATTERNS.
func (e *env) raw(key string) (string, bool) {
	v, ok := os.LookupEnv(key)
	return strings.TrimSpace(v), ok && v != ""
}

func (e *env) str(key, def string) string {
	if v, ok := e.raw(key); ok {
		return v
	}
	return def
}

func (e *env) int(key string, def int) int {
	v, ok := e.raw(key)
	if !ok {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		e.errs = append(e.errs, fmt.Errorf("config: %s=%q is not an integer", key, v))
		return def
	}
	return n
}

func (e *env) int32(key string, def int32) int32 {
	v, ok := e.raw(key)
	if !ok {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		e.errs = append(e.errs, fmt.Errorf("config: %s=%q is not a 32-bit integer", key, v))
		return def
	}
	return int32(n)
}

func (e *env) bool(key string, def bool) bool {
	v, ok := e.raw(key)
	if !ok {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		e.errs = append(e.errs, fmt.Errorf("config: %s=%q is not a boolean", key, v))
		return def
	}
	return b
}

func (e *env) dur(key string, def time.Duration) time.Duration {
	v, ok := e.raw(key)
	if !ok {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		e.errs = append(e.errs, fmt.Errorf("config: %s=%q is not a duration (use Go syntax such as 90s, 15m, 168h)", key, v))
		return def
	}
	return d
}

// splitList parses a comma-or-space separated list, trimming empties.
func splitList(s string) []string {
	fields := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' || r == '\n' })
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

// parseHeaders parses OTEL_EXPORTER_OTLP_HEADERS ("k1=v1,k2=v2").
func parseHeaders(s string) map[string]string {
	out := map[string]string{}
	for _, pair := range strings.Split(s, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		k, v, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}
		out[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return out
}

// endpointFromDSN extracts the OTLP base endpoint from an Uptrace DSN of the
// form "https://TOKEN@host[:port][/path]". Returns the base URL and whether it
// is insecure (http).
func endpointFromDSN(dsn string) (endpoint string, insecure bool) {
	// Very small, dependency-free parse: scheme://[userinfo@]host
	scheme := "https"
	rest := dsn
	if i := strings.Index(rest, "://"); i >= 0 {
		scheme = rest[:i]
		rest = rest[i+3:]
	}
	if at := strings.LastIndex(rest, "@"); at >= 0 {
		rest = rest[at+1:]
	}
	// strip any path/query
	if q := strings.IndexAny(rest, "/?"); q >= 0 {
		rest = rest[:q]
	}
	return scheme + "://" + rest, scheme == "http"
}

// defaultInstanceID returns a stable-ish identifier used when the operator did
// not supply one (mostly for bare `go run` during development).
func defaultInstanceID() string {
	if h, err := os.Hostname(); err == nil && h != "" {
		return h
	}
	return "local-" + hex.EncodeToString(mustRandom(4))
}

func mustRandom(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("config: crypto/rand failed: " + err.Error())
	}
	return b
}

func envStr(key, def string) string {
	return (&env{}).str(key, def)
}
