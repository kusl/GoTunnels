package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoad_RequiresDatabaseURL(t *testing.T) {
	t.Setenv("GOTUNNELS_DEV", "1") // auto-generate secrets
	// DATABASE_URL deliberately unset.
	if _, err := Load(); err == nil {
		t.Fatal("expected error when DATABASE_URL is missing")
	}
}

func TestLoad_RequiresSecretsWhenNotDev(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x/y")
	t.Setenv("GOTUNNELS_DEV", "0")
	// no pepper / totp key
	if _, err := Load(); err == nil {
		t.Fatal("expected error when secrets are missing and not in dev mode")
	}
}

func TestLoad_DevGeneratesSecrets(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/gotunnels")
	t.Setenv("GOTUNNELS_DEV", "1")
	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.IPHashPepper()) == 0 {
		t.Error("expected a generated IP hash pepper in dev mode")
	}
	var zero [32]byte
	if c.TOTPAESKey() == zero {
		t.Error("expected a non-zero TOTP AES key in dev mode")
	}
	if c.CSPMode != "report-only" {
		t.Errorf("expected default CSP mode report-only, got %q", c.CSPMode)
	}
	if !strings.Contains(c.CSPPolicy, "default-src 'self'") {
		t.Errorf("default CSP policy should lock to self, got %q", c.CSPPolicy)
	}
}

func TestLoad_TOTPKeyDerivedDeterministically(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/gotunnels")
	t.Setenv("GOTUNNELS_IP_HASH_PEPPER", "pepper")
	t.Setenv("GOTUNNELS_TOTP_ENCRYPTION_KEY", "the-same-secret")

	c1, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	c2, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c1.TOTPAESKey() != c2.TOTPAESKey() {
		t.Error("TOTP key derivation must be deterministic for a given secret")
	}
}

func TestValidate_CSPMode(t *testing.T) {
	c := &Config{
		DatabaseURL: "x",
		HTTPAddr:    ":8080",
		SessionTTL:  time.Hour,
		RPOrigins:   []string{"http://localhost"},
		CSPMode:     "nonsense",
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected invalid CSP mode to fail validation")
	}
	c.CSPMode = "enforce"
	if err := c.Validate(); err != nil {
		t.Fatalf("enforce should be valid: %v", err)
	}
}

func TestSplitList(t *testing.T) {
	cases := map[string][]string{
		"a,b,c":            {"a", "b", "c"},
		"a b c":            {"a", "b", "c"},
		" a , b ,, c ":     {"a", "b", "c"},
		"https://x https:": {"https://x", "https:"},
		"":                 {},
	}
	for in, want := range cases {
		got := splitList(in)
		if len(got) != len(want) {
			t.Errorf("splitList(%q) len = %d, want %d (%v)", in, len(got), len(want), got)
			continue
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("splitList(%q)[%d] = %q, want %q", in, i, got[i], want[i])
			}
		}
	}
}

func TestParseHeaders(t *testing.T) {
	h := parseHeaders("uptrace-dsn=https://tok@api.uptrace.dev,x-extra = val ")
	if h["uptrace-dsn"] != "https://tok@api.uptrace.dev" {
		t.Errorf("unexpected dsn header: %q", h["uptrace-dsn"])
	}
	if h["x-extra"] != "val" {
		t.Errorf("unexpected x-extra header: %q", h["x-extra"])
	}
}

func TestEndpointFromDSN(t *testing.T) {
	cases := []struct {
		dsn          string
		wantEndpoint string
		wantInsecure bool
	}{
		{"https://TOKEN@api.uptrace.dev?grpc=4317", "https://api.uptrace.dev", false},
		{"https://TOKEN@api.uptrace.dev:443/v1", "https://api.uptrace.dev:443", false},
		{"http://token@localhost:14318", "http://localhost:14318", true},
	}
	for _, tc := range cases {
		ep, insecure := endpointFromDSN(tc.dsn)
		if ep != tc.wantEndpoint {
			t.Errorf("endpointFromDSN(%q) endpoint = %q, want %q", tc.dsn, ep, tc.wantEndpoint)
		}
		if insecure != tc.wantInsecure {
			t.Errorf("endpointFromDSN(%q) insecure = %v, want %v", tc.dsn, insecure, tc.wantInsecure)
		}
	}
}

func TestResolveTelemetry_DisabledByDefault(t *testing.T) {
	// Ensure a clean env for the OTLP vars.
	t.Setenv("UPTRACE_DSN", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	c := &Config{}
	c.resolveTelemetry()
	if c.Telemetry.Enabled {
		t.Error("telemetry should be disabled when no endpoint/DSN is configured")
	}
}

func TestResolveTelemetry_UptraceDSN(t *testing.T) {
	t.Setenv("UPTRACE_DSN", "https://secret@api.uptrace.dev?grpc=4317")
	c := &Config{}
	c.resolveTelemetry()
	if !c.Telemetry.Enabled {
		t.Fatal("telemetry should be enabled with an Uptrace DSN")
	}
	if c.Telemetry.EndpointURL != "https://api.uptrace.dev" {
		t.Errorf("endpoint = %q", c.Telemetry.EndpointURL)
	}
	if c.Telemetry.Headers["uptrace-dsn"] == "" {
		t.Error("expected uptrace-dsn header to be set")
	}
}

func TestGetHelpers(t *testing.T) {
	t.Setenv("X_INT", "42")
	if getint("X_INT", 0) != 42 {
		t.Error("getint failed")
	}
	if getint("X_MISSING", 7) != 7 {
		t.Error("getint default failed")
	}
	t.Setenv("X_BOOL", "true")
	if !getbool("X_BOOL", false) {
		t.Error("getbool failed")
	}
	t.Setenv("X_DUR", "250ms")
	if getdur("X_DUR", 0) != 250*time.Millisecond {
		t.Error("getdur failed")
	}
}

func TestLoad_SignupLimitDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/gotunnels")
	t.Setenv("GOTUNNELS_DEV", "1")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.SignupIPInterval != 5*time.Minute || c.SignupIPBurst != 1 {
		t.Fatalf("ip defaults = (%v, %d), want (5m, 1)", c.SignupIPInterval, c.SignupIPBurst)
	}
	if c.SignupGlobalInterval != time.Minute || c.SignupGlobalBurst != 1 {
		t.Fatalf("global defaults = (%v, %d), want (1m, 1)", c.SignupGlobalInterval, c.SignupGlobalBurst)
	}
}

func TestLoad_SignupLimitOverridesAndDisable(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/gotunnels")
	t.Setenv("GOTUNNELS_DEV", "1")
	t.Setenv("GOTUNNELS_SIGNUP_IP_INTERVAL", "30s")
	t.Setenv("GOTUNNELS_SIGNUP_IP_BURST", "3")
	t.Setenv("GOTUNNELS_SIGNUP_GLOBAL_INTERVAL", "0")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.SignupIPInterval != 30*time.Second || c.SignupIPBurst != 3 {
		t.Fatalf("ip override = (%v, %d)", c.SignupIPInterval, c.SignupIPBurst)
	}
	if c.SignupGlobalInterval != 0 {
		t.Fatalf("global interval = %v, want 0 (disabled)", c.SignupGlobalInterval)
	}
}

func TestValidate_SignupBurstMustBePositiveWhenEnabled(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/gotunnels")
	t.Setenv("GOTUNNELS_DEV", "1")
	t.Setenv("GOTUNNELS_SIGNUP_IP_BURST", "0")
	if _, err := Load(); err == nil {
		t.Fatal("burst 0 with a non-zero interval should fail validation")
	}
}

func TestLoad_RPOriginPatternsDefault(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/gotunnels")
	t.Setenv("GOTUNNELS_DEV", "1")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(c.RPOriginPatterns) != 1 || c.RPOriginPatterns[0] != "https://*.trycloudflare.com" {
		t.Fatalf("patterns = %v, want the trycloudflare default", c.RPOriginPatterns)
	}
}

func TestLoad_RPOriginPatternsOverrideAndClear(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/gotunnels")
	t.Setenv("GOTUNNELS_DEV", "1")
	t.Setenv("GOTUNNELS_RP_ORIGIN_PATTERNS", "https://*.a.example, https://*.b.example")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(c.RPOriginPatterns) != 2 || c.RPOriginPatterns[0] != "https://*.a.example" || c.RPOriginPatterns[1] != "https://*.b.example" {
		t.Fatalf("patterns = %v", c.RPOriginPatterns)
	}

	t.Setenv("GOTUNNELS_RP_ORIGIN_PATTERNS", " ")
	c2, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(c2.RPOriginPatterns) != 0 {
		t.Fatalf("blank override should clear patterns, got %v", c2.RPOriginPatterns)
	}
}

func TestLoad_HostName(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/gotunnels")
	t.Setenv("GOTUNNELS_DEV", "1")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.HostName != "" {
		t.Fatalf("default host name = %q, want empty (SDK fallback)", c.HostName)
	}
	t.Setenv("GOTUNNELS_HOST_NAME", "virginia")
	c2, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c2.HostName != "virginia" {
		t.Fatalf("host name = %q, want virginia", c2.HostName)
	}
}

func TestPerInterval(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want float64
	}{
		{time.Second, 1},
		{time.Minute, 1.0 / 60.0},
		{5 * time.Minute, 1.0 / 300.0},
		{0, 0},
		{-time.Second, 0},
	}
	for _, tc := range cases {
		if got := PerInterval(tc.in); got != tc.want {
			t.Errorf("PerInterval(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestValidate_SessionTTL(t *testing.T) {
	// A session TTL of zero is the new default and means "never expires":
	// sessions live until an explicit logout. It must pass validation.
	base := func() *Config {
		return &Config{
			DatabaseURL: "x",
			HTTPAddr:    ":8080",
			RPOrigins:   []string{"http://localhost"},
			CSPMode:     "report-only",
		}
	}

	zero := base()
	zero.SessionTTL = 0
	if err := zero.Validate(); err != nil {
		t.Fatalf("SessionTTL=0 (never expires) must be valid: %v", err)
	}

	positive := base()
	positive.SessionTTL = 48 * time.Hour
	if err := positive.Validate(); err != nil {
		t.Fatalf("SessionTTL=48h must be valid: %v", err)
	}

	negative := base()
	negative.SessionTTL = -time.Second
	if err := negative.Validate(); err == nil {
		t.Fatal("negative SessionTTL must fail validation")
	}
}

func TestLoad_SessionTTLDefaultsToNeverExpire(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/gotunnels")
	t.Setenv("GOTUNNELS_DEV", "1")
	// Deliberately do NOT set GOTUNNELS_SESSION_TTL.
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.SessionTTL != 0 {
		t.Fatalf("default SessionTTL = %v, want 0 (never expires)", c.SessionTTL)
	}
}

func TestLoad_SessionTTLOverride(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/gotunnels")
	t.Setenv("GOTUNNELS_DEV", "1")
	t.Setenv("GOTUNNELS_SESSION_TTL", "48h")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.SessionTTL != 48*time.Hour {
		t.Fatalf("SessionTTL override = %v, want 48h", c.SessionTTL)
	}
}
