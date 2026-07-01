package auth

import (
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
