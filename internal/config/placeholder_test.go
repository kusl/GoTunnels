package config

import (
	"strings"
	"testing"
)

func TestLoad_RejectsUnexpandedComposePlaceholders(t *testing.T) {
	cases := map[string]string{
		"GOTUNNELS_SIGNUP_IP_INTERVAL": "${GOTUNNELS_SIGNUP_IP_INTERVAL:-5m}",
		"GOTUNNELS_AUTH_ATTEMPT_BURST": "${GOTUNNELS_AUTH_ATTEMPT_BURST:-10}",
		"GOTUNNELS_RP_ORIGIN_PATTERNS": "${GOTUNNELS_RP_ORIGIN_PATTERNS:-https://*.trycloudflare.com}",
		"GOTUNNELS_CSP_MODE":           "${GOTUNNELS_CSP_MODE:-report-only}",
		"GOTUNNELS_DEV":                "${GOTUNNELS_DEV:-false}",
	}
	for key, val := range cases {
		t.Run(key, func(t *testing.T) {
			t.Setenv("DATABASE_URL", "postgres://localhost/gotunnels")
			t.Setenv("GOTUNNELS_IP_HASH_PEPPER", "p")
			t.Setenv("GOTUNNELS_TOTP_ENCRYPTION_KEY", "k")
			t.Setenv(key, val)
			_, err := Load()
			if err == nil {
				t.Fatalf("Load accepted %s=%q", key, val)
			}
			msg := err.Error()
			if !strings.Contains(msg, key) || !strings.Contains(msg, "unexpanded compose placeholder") {
				t.Fatalf("error does not name the placeholder problem for %s: %v", key, err)
			}
		})
	}
}

func TestEnvReader_PlaceholderFallsBackAndRecordsError(t *testing.T) {
	t.Setenv("X_PLACEHOLDER", "${X_PLACEHOLDER:-1h}")
	e := &env{}
	if got := e.str("X_PLACEHOLDER", "def"); got != "def" {
		t.Fatalf("str = %q, want default", got)
	}
	if e.err() == nil {
		t.Fatal("placeholder must be recorded as an error")
	}
}

func TestEnvReader_DollarWithoutBraceIsAllowed(t *testing.T) {
	t.Setenv("X_DOLLAR", "pa$$word")
	e := &env{}
	if got := e.str("X_DOLLAR", ""); got != "pa$$word" {
		t.Fatalf("str = %q", got)
	}
	if err := e.err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
