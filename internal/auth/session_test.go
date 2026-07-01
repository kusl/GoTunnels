package auth

import "testing"

func TestNewSessionToken(t *testing.T) {
	token, id, err := NewSessionToken()
	if err != nil {
		t.Fatalf("NewSessionToken: %v", err)
	}
	if token == "" || id == "" {
		t.Fatal("expected non-empty token and id")
	}
	if token == id {
		t.Fatal("token and id must differ (id is a hash of token)")
	}
	if len(id) != 64 { // sha256 hex
		t.Fatalf("id length = %d, want 64", len(id))
	}
	if HashSessionToken(token) != id {
		t.Fatal("HashSessionToken not consistent with NewSessionToken")
	}
}

func TestSessionTokensUnique(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		token, _, err := NewSessionToken()
		if err != nil {
			t.Fatalf("NewSessionToken: %v", err)
		}
		if seen[token] {
			t.Fatal("duplicate session token generated")
		}
		seen[token] = true
	}
}

func TestHashSessionTokenDeterministic(t *testing.T) {
	if HashSessionToken("abc") != HashSessionToken("abc") {
		t.Fatal("hashing must be deterministic")
	}
	if HashSessionToken("abc") == HashSessionToken("abd") {
		t.Fatal("different tokens must hash differently")
	}
}
