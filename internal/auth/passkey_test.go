package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kusl/GoTunnels/internal/store"
)

func TestWebAuthnUserAdapter(t *testing.T) {
	u := &WebAuthnUser{User: store.User{ID: "abc-123", Username: "alice", DisplayName: "Alice"}}
	if string(u.WebAuthnID()) != "abc-123" {
		t.Fatalf("WebAuthnID = %q", u.WebAuthnID())
	}
	if u.WebAuthnName() != "alice" {
		t.Fatalf("WebAuthnName = %q", u.WebAuthnName())
	}
	if u.WebAuthnDisplayName() != "Alice" {
		t.Fatalf("WebAuthnDisplayName = %q", u.WebAuthnDisplayName())
	}
	if len(u.WebAuthnCredentials()) != 0 {
		t.Fatal("expected no credentials")
	}
}

func TestWebAuthnUserDisplayNameFallback(t *testing.T) {
	u := &WebAuthnUser{User: store.User{ID: "x", Username: "bob"}}
	if u.WebAuthnDisplayName() != "bob" {
		t.Fatalf("expected fallback to username, got %q", u.WebAuthnDisplayName())
	}
}

func TestNewWebAuthnValid(t *testing.T) {
	w, err := NewWebAuthn("example.com", "GoTunnels", []string{"https://example.com"})
	if err != nil {
		t.Fatalf("NewWebAuthn: %v", err)
	}
	if w == nil {
		t.Fatal("expected non-nil WebAuthn")
	}
}

func testRPProvider(t *testing.T) *RPProvider {
	t.Helper()
	p, err := NewRPProvider(RPConfig{
		RPID:           "localhost",
		RPDisplayName:  "GoTunnels",
		RPOrigins:      []string{"http://localhost:8080"},
		OriginPatterns: []string{"https://*.trycloudflare.com"},
	})
	if err != nil {
		t.Fatalf("NewRPProvider: %v", err)
	}
	return p
}

func TestRPProviderEmptyOriginFallsBackToStatic(t *testing.T) {
	p := testRPProvider(t)
	wa, origin, err := p.ForOrigin("")
	if err != nil {
		t.Fatalf("ForOrigin(\"\"): %v", err)
	}
	if wa != p.Static() {
		t.Fatal("empty origin should return the static relying party")
	}
	if origin != "http://localhost:8080" {
		t.Fatalf("canonical origin = %q, want the first static origin", origin)
	}
}

func TestRPProviderStaticOriginUsesStaticInstance(t *testing.T) {
	p := testRPProvider(t)
	wa, origin, err := p.ForOrigin("HTTP://LOCALHOST:8080")
	if err != nil {
		t.Fatalf("ForOrigin(static): %v", err)
	}
	if wa != p.Static() {
		t.Fatal("configured origin should return the static relying party")
	}
	if origin != "http://localhost:8080" {
		t.Fatalf("canonical origin = %q, want lowercased static origin", origin)
	}
}

func TestRPProviderDerivesForPatternMatch(t *testing.T) {
	p := testRPProvider(t)
	wa, origin, err := p.ForOrigin("https://Tart-Panda.TryCloudflare.com")
	if err != nil {
		t.Fatalf("ForOrigin(pattern): %v", err)
	}
	if wa == p.Static() {
		t.Fatal("pattern-matched origin should get a derived relying party, not the static one")
	}
	if origin != "https://tart-panda.trycloudflare.com" {
		t.Fatalf("canonical origin = %q, want lowercased request origin", origin)
	}
	if got := wa.Config.RPID; got != "tart-panda.trycloudflare.com" {
		t.Fatalf("derived RPID = %q, want the origin host", got)
	}
	if len(wa.Config.RPOrigins) != 1 || wa.Config.RPOrigins[0] != "https://tart-panda.trycloudflare.com" {
		t.Fatalf("derived RPOrigins = %v, want exactly the canonical origin", wa.Config.RPOrigins)
	}
}

func TestRPProviderCachesDerivedInstances(t *testing.T) {
	p := testRPProvider(t)
	a, _, err := p.ForOrigin("https://same.trycloudflare.com")
	if err != nil {
		t.Fatal(err)
	}
	b, _, err := p.ForOrigin("https://same.trycloudflare.com")
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatal("same origin should reuse the cached derived instance")
	}
	c, _, err := p.ForOrigin("https://other.trycloudflare.com")
	if err != nil {
		t.Fatal(err)
	}
	if c == a {
		t.Fatal("different origins must not share a derived instance")
	}
}

func TestRPProviderRejectsUntrustedOrigin(t *testing.T) {
	p := testRPProvider(t)
	for _, origin := range []string{
		"https://evil.example",
		"http://a.trycloudflare.com",          // scheme downgrade
		"https://a.b.trycloudflare.com",       // two labels
		"https://trycloudflare.com",           // zero labels
		"https://a.trycloudflare.com.evil.io", // suffix trick
	} {
		if _, _, err := p.ForOrigin(origin); err == nil {
			t.Errorf("ForOrigin(%q) accepted, want rejection", origin)
		}
	}
}

func TestRPProviderForRequestReadsOriginHeader(t *testing.T) {
	p := testRPProvider(t)
	req := httptest.NewRequest(http.MethodPost, "/api/passkey/login/begin", nil)
	req.Header.Set("Origin", "https://from-header.trycloudflare.com")
	wa, origin, err := p.ForRequest(req)
	if err != nil {
		t.Fatalf("ForRequest: %v", err)
	}
	if origin != "https://from-header.trycloudflare.com" {
		t.Fatalf("canonical origin = %q", origin)
	}
	if wa.Config.RPID != "from-header.trycloudflare.com" {
		t.Fatalf("RPID = %q", wa.Config.RPID)
	}
}

func TestHostOfOrigin(t *testing.T) {
	cases := []struct{ in, want string }{
		{"https://a.trycloudflare.com", "a.trycloudflare.com"},
		{"http://localhost:8080", "localhost"},
		{"https://x.example:8443", "x.example"},
		{"", ""},
		// hostOfOrigin is a plain extractor by design: validation happens
		// earlier (MatchOriginPattern), so a scheme-less string passes through.
		{"not-an-origin", "not-an-origin"},
	}
	for _, tc := range cases {
		if got := hostOfOrigin(tc.in); got != tc.want {
			t.Errorf("hostOfOrigin(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
