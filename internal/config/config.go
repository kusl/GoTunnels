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
	"encoding/base64"
	"encoding/hex"
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

// Load reads and validates configuration from the environment.
func Load() (*Config, error) {
	c := &Config{
		InstanceID:           getenv("GOTUNNELS_INSTANCE_ID", defaultInstanceID()),
		ServiceName:          getenv("OTEL_SERVICE_NAME", getenv("GOTUNNELS_SERVICE_NAME", "gotunnels-api")),
		Version:              getenv("GOTUNNELS_VERSION", "dev"),
		HTTPAddr:             getenv("GOTUNNELS_HTTP_ADDR", ":8080"),
		ShutdownTimeout:      getdur("GOTUNNELS_SHUTDOWN_TIMEOUT", 15*time.Second),
		DatabaseURL:          getenv("DATABASE_URL", ""),
		DBMaxConns:           int32(getint("GOTUNNELS_DB_MAX_CONNS", 20)),
		DBMinConns:           int32(getint("GOTUNNELS_DB_MIN_CONNS", 2)),
		DBConnectTimeout:     getdur("GOTUNNELS_DB_CONNECT_TIMEOUT", 30*time.Second),
		SessionCookieName:    getenv("GOTUNNELS_SESSION_COOKIE_NAME", "gotunnels_session"),
		SessionTTL:           getdur("GOTUNNELS_SESSION_TTL", 0), // 0 = never expires (see field doc)
		CORSAllowedOrigins:   splitList(getenv("GOTUNNELS_CORS_ALLOWED_ORIGINS", "*")),
		RPID:                 getenv("GOTUNNELS_RP_ID", "localhost"),
		RPDisplayName:        getenv("GOTUNNELS_RP_DISPLAY_NAME", "GoTunnels"),
		RPOrigins:            splitList(getenv("GOTUNNELS_RP_ORIGINS", "http://localhost:8080")),
		RPOriginPatterns:     splitList(getenv("GOTUNNELS_RP_ORIGIN_PATTERNS", "https://*.trycloudflare.com")),
		SignupIPInterval:     getdur("GOTUNNELS_SIGNUP_IP_INTERVAL", 5*time.Minute),
		SignupIPBurst:        getint("GOTUNNELS_SIGNUP_IP_BURST", 1),
		SignupGlobalInterval: getdur("GOTUNNELS_SIGNUP_GLOBAL_INTERVAL", time.Minute),
		SignupGlobalBurst:    getint("GOTUNNELS_SIGNUP_GLOBAL_BURST", 1),
		HostName:             getenv("GOTUNNELS_HOST_NAME", ""),
		CSPMode:              getenv("GOTUNNELS_CSP_MODE", "report-only"),
		CSPPolicy:            getenv("GOTUNNELS_CSP_POLICY", DefaultCSPPolicy),
		Dev:                  getbool("GOTUNNELS_DEV", false),
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
		Compression: strings.ToLower(getenv("OTEL_EXPORTER_OTLP_COMPRESSION", "gzip")),
		MetricsTemporality: strings.ToLower(getenv(
			"OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE", "delta")),
		MetricsHistogram: strings.ToLower(getenv(
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

func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getint(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return n
		}
	}
	return def
}

func getbool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(strings.TrimSpace(v)); err == nil {
			return b
		}
	}
	return def
}

func getdur(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(strings.TrimSpace(v)); err == nil {
			return d
		}
	}
	return def
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

// DecodeKeyMaterial is a helper exposed for tests and tooling: it accepts hex,
// base64 (std or url), or raw bytes and returns them. Unused by Load directly
// but kept for completeness where callers want exact-length keys.
func DecodeKeyMaterial(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if b, err := hex.DecodeString(s); err == nil && len(s)%2 == 0 {
		return b, nil
	}
	if b, err := base64.StdEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	if b, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	return []byte(s), nil
}
