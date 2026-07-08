package config

// These tests pin the cross-file contract for the "sessions never expire until
// the user logs out" behaviour and the "log out everywhere" control. None of
// this can be exercised by a plain unit test in a single package: the promise
// spans the Go config default, the store's SQL, the auth cookie logic, the HTTP
// route table, the frontend token store, and the settings UI. Reading the real
// files keeps those layers honest so a well-meaning refactor in any one of them
// cannot silently reintroduce inactivity logout, per-tab token loss, or a dead
// "log out everywhere" button.
//
// repoRoot and mustRead are defined in csp_deployment_test.go (same package).
//
// stdlib only, no network, no containers: safe under `go test ./...` both in CI
// and in the offline sandbox.

import (
	"regexp"
	"strings"
	"testing"
)

// TestTokenStoredInLocalStorage pins the fix for the new-tab logout defect:
// the Bearer token must live in localStorage (shared across tabs, survives
// browser restart), NOT sessionStorage (per-tab, cleared on close). A single
// mention of sessionStorage is allowed — the one-time legacy migration — but
// the getter/setter/clearer must all use localStorage.
func TestTokenStoredInLocalStorage(t *testing.T) {
	api := mustRead(t, "frontend/js/api.js")

	for _, want := range []string{
		`localStorage.getItem(TOKEN_KEY)`,
		`localStorage.setItem(TOKEN_KEY`,
		`localStorage.removeItem(TOKEN_KEY)`,
	} {
		if !strings.Contains(api, want) {
			t.Errorf("frontend/js/api.js: token store must use %q — a new tab would otherwise bounce to /login", want)
		}
	}

	// Guard against a regression to a sessionStorage-backed getter. We allow
	// sessionStorage in the one-time migration/cleanup code (read the legacy
	// value once, remove it, remove any stray copy on clear = 3 code uses) but
	// nothing more. Count only executable lines so the explanatory comments
	// above the code do not inflate the total.
	codeUses := 0
	for _, line := range strings.Split(api, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "//") {
			continue
		}
		codeUses += strings.Count(line, "sessionStorage")
	}
	if codeUses > 3 {
		t.Errorf("frontend/js/api.js: %d sessionStorage uses in code — the primary token store must be localStorage, not sessionStorage", codeUses)
	}
}

// TestSessionTTLDocumentedInEnvExample pins that operators can discover the
// knob and that its default is the never-expire value.
func TestSessionTTLDocumentedInEnvExample(t *testing.T) {
	env := mustRead(t, ".env.example")
	re := regexp.MustCompile(`(?m)^GOTUNNELS_SESSION_TTL=(\S+)`)
	m := re.FindStringSubmatch(env)
	if m == nil {
		t.Fatal(".env.example: GOTUNNELS_SESSION_TTL is not documented")
	}
	if m[1] != "0" {
		t.Errorf(".env.example: GOTUNNELS_SESSION_TTL default = %q, want 0 (never expires)", m[1])
	}
}

// TestLogoutEverywhereRouteRegistered pins the server route that backs the new
// settings button.
func TestLogoutEverywhereRouteRegistered(t *testing.T) {
	server := mustRead(t, "internal/server/server.go")
	if !strings.Contains(server, `"POST /api/logout-all"`) {
		t.Error(`internal/server/server.go: route "POST /api/logout-all" not found — the "log out everywhere" button would 404`)
	}
}

// TestLogoutEverywhereWiredEndToEnd pins the rest of the chain: the store
// method that revokes every session, the handler that calls it, the frontend
// API wrapper, the settings-page click handler, and the button itself.
func TestLogoutEverywhereWiredEndToEnd(t *testing.T) {
	cases := []struct {
		rel  string
		want string
		why  string
	}{
		{
			"internal/store/store.go",
			"func (s *Store) RevokeAllSessionsForUser(",
			"the store must expose a bulk-revoke used by logout-all",
		},
		{
			"internal/auth/handlers.go",
			"func (h *Handlers) LogoutAll(",
			"the auth package must handle the logout-all request",
		},
		{
			"internal/auth/handlers.go",
			"RevokeAllSessionsForUser(",
			"the LogoutAll handler must actually revoke every session",
		},
		{
			"frontend/js/api.js",
			`logoutAll:`,
			"the frontend API client must expose logoutAll()",
		},
		{
			"frontend/js/page-settings.js",
			"Api.logoutAll(",
			"the settings page must call logoutAll on click",
		},
		{
			"frontend/settings.html",
			`id="logout-all"`,
			"the settings page must render the log-out-everywhere button",
		},
	}
	for _, tc := range cases {
		if !strings.Contains(mustRead(t, tc.rel), tc.want) {
			t.Errorf("%s: missing %q — %s", tc.rel, tc.want, tc.why)
		}
	}
}

// TestClientNeverLogsOutOnTransientError pins that the frontend only treats a
// definitive 401 (or an absent token) as logged out. A network blip or 5xx
// must NOT clear the token or redirect to /login — otherwise a hiccup becomes
// an involuntary logout, which this app refuses to do. Catches a regression to
// the old catch-all `catch { return null }`.
func TestClientNeverLogsOutOnTransientError(t *testing.T) {
	common := mustRead(t, "frontend/js/common.js")
	if !strings.Contains(common, "status === 401") {
		t.Error("frontend/js/common.js: session check must gate logout on a definitive 401, not any error")
	}
	if !strings.Contains(common, "uncertain") {
		t.Error("frontend/js/common.js: a transient (uncertain) session check must keep the user on the page rather than redirect")
	}
}

// never-expiring sessions possible: CreateSession takes a nullable expiry and
// GetSession treats a NULL expires_at as "valid forever".
func TestSessionStoreHonoursNullExpiry(t *testing.T) {
	store := mustRead(t, "internal/store/store.go")

	if !strings.Contains(store, "expiresAt *time.Time") {
		t.Error("internal/store/store.go: CreateSession must accept a *time.Time expiry so NULL (never expires) is representable")
	}
	if !strings.Contains(store, "expires_at IS NULL OR expires_at > now()") {
		t.Error("internal/store/store.go: GetSession must treat NULL expires_at as non-expiring (expires_at IS NULL OR expires_at > now())")
	}
}

// TestOptionalExpiryMigrationIsReversible pins that migration 0008 drops the
// NOT NULL constraint and ships a matching down-migration, so the schema can
// actually hold a NULL expiry. (database_test.go independently enforces the
// up/down pairing for every migration; this adds an intent check on 0008.)
func TestOptionalExpiryMigrationIsReversible(t *testing.T) {
	up := mustRead(t, "migrations/0008_sessions_optional_expiry.up.sql")
	if !regexp.MustCompile(`(?i)alter\s+column\s+expires_at\s+drop\s+not\s+null`).MatchString(up) {
		t.Error("migrations/0008 up: expected `ALTER COLUMN expires_at DROP NOT NULL`")
	}
	down := mustRead(t, "migrations/0008_sessions_optional_expiry.down.sql")
	if !regexp.MustCompile(`(?i)set\s+not\s+null`).MatchString(down) {
		t.Error("migrations/0008 down: expected the NOT NULL constraint to be restored")
	}
}
