package config

// These tests pin the deployment files to the canonical DefaultCSPPolicy and
// to the CSP reporting wiring the Caddyfile must carry. The policy string is
// duplicated by necessity — the compose default, the Caddyfile fallback, the
// .env template in scripts/lib.sh, .env.example, and the docs all repeat it —
// and the copies HAD drifted: every one of them was missing media-src and
// frame-src relative to this constant, so /api/info advertised a stricter
// policy than the frontend actually emitted. Reading the real files from the
// repo keeps every copy honest, and pins the report-uri/report-to/
// Reporting-Endpoints plumbing so native browser reporting cannot silently
// regress.
//
// stdlib only, no network, no containers: safe under `go test ./...` both in
// CI and in the offline sandbox.

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// repoRoot resolves the repository root relative to this source file
// (internal/config/ -> ../..) so `go test ./...` works from any directory.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
}

func mustRead(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(b)
}

// extractOne returns the single capture group of re in text, failing the test
// when the pattern is absent or ambiguous.
func extractOne(t *testing.T, rel, text string, re *regexp.Regexp) string {
	t.Helper()
	m := re.FindAllStringSubmatch(text, -1)
	if len(m) == 0 {
		t.Fatalf("%s: pattern %q not found", rel, re)
	}
	if len(m) > 1 {
		t.Fatalf("%s: pattern %q matched %d times, want exactly 1", rel, re, len(m))
	}
	return m[0][1]
}

func TestDefaultCSPPolicyCopiesAgree(t *testing.T) {
	cases := []struct {
		rel string
		re  *regexp.Regexp
	}{
		// Caddyfile fallback inside the env placeholder:
		//   {$GOTUNNELS_CSP_POLICY:<default>}
		{"frontend/Caddyfile", regexp.MustCompile(`\{\$GOTUNNELS_CSP_POLICY:([^}]*)\}`)},
		// compose.yaml default for the FRONTEND service:
		//   GOTUNNELS_CSP_POLICY: "${GOTUNNELS_CSP_POLICY:-<default>}"
		// [^}]+ (non-empty) so the api service's deliberately empty mirror
		// (asserted separately below) is not caught here.
		{"compose.yaml", regexp.MustCompile(`GOTUNNELS_CSP_POLICY: "\$\{GOTUNNELS_CSP_POLICY:-([^}]+)\}"`)},
		// .env template written by scripts/lib.sh ensure_env:
		{"scripts/lib.sh", regexp.MustCompile(`(?m)^GOTUNNELS_CSP_POLICY="([^"]*)"$`)},
		// documented example env file:
		{".env.example", regexp.MustCompile(`(?m)^GOTUNNELS_CSP_POLICY="([^"]*)"$`)},
	}
	for _, tc := range cases {
		got := extractOne(t, tc.rel, mustRead(t, tc.rel), tc.re)
		if got != DefaultCSPPolicy {
			t.Errorf("%s: CSP policy default drifted from config.DefaultCSPPolicy\n got: %s\nwant: %s",
				tc.rel, got, DefaultCSPPolicy)
		}
	}
}

func TestComposeMirrorsCSPPolicyToAPI(t *testing.T) {
	compose := mustRead(t, "compose.yaml")
	// The api service must receive the SAME variable with an EMPTY default:
	// config.getenv treats empty as unset, so with nothing in .env the API
	// falls back to DefaultCSPPolicy (== the frontend default per the test
	// above), and with a customised .env both services see the same value —
	// either way /api/info stays honest.
	if !strings.Contains(compose, `GOTUNNELS_CSP_POLICY: "${GOTUNNELS_CSP_POLICY:-}"`) {
		t.Error(`compose.yaml: api service must pass GOTUNNELS_CSP_POLICY through with an empty default ("${GOTUNNELS_CSP_POLICY:-}")`)
	}
}

func TestCaddyfileWiresCSPReporting(t *testing.T) {
	caddy := mustRead(t, "frontend/Caddyfile")

	for _, want := range []string{
		// Reporting directives appended OUTSIDE the policy placeholder, so a
		// customised (or stale, pre-reporting) GOTUNNELS_CSP_POLICY in .env
		// still reports.
		`}; report-uri /csp-report; report-to csp-endpoint"`,
		// Reporting API endpoint group matching `report-to csp-endpoint`.
		"Reporting-Endpoints `csp-endpoint=\"/csp-report\"`",
		// Same-origin ingest path proxied to the API on the compose network.
		"handle /csp-report",
		"rewrite * /api/csp-reports",
		"reverse_proxy {$GOTUNNELS_API_UPSTREAM:api:8080}",
	} {
		if !strings.Contains(caddy, want) {
			t.Errorf("frontend/Caddyfile: missing %q", want)
		}
	}

	// The proxy target must actually exist on the API side, or /csp-report
	// would forward into a 404.
	server := mustRead(t, "internal/server/server.go")
	if !strings.Contains(server, `"POST /api/csp-reports"`) {
		t.Error(`internal/server/server.go: route "POST /api/csp-reports" not found — the /csp-report proxy would 404`)
	}
}

func TestConfigurationDocsShowCanonicalPolicy(t *testing.T) {
	doc := mustRead(t, "docs/CONFIGURATION.md")
	re := regexp.MustCompile("(?s)```\n(default-src 'self';.*?)\n```")
	m := re.FindStringSubmatch(doc)
	if m == nil {
		t.Fatal("docs/CONFIGURATION.md: fenced default-policy block not found")
	}
	got := strings.Join(strings.Fields(m[1]), " ")
	if got != DefaultCSPPolicy {
		t.Errorf("docs/CONFIGURATION.md policy block drifted from config.DefaultCSPPolicy\n got: %s\nwant: %s",
			got, DefaultCSPPolicy)
	}
}
