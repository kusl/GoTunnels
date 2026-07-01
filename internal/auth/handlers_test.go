package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidUsername(t *testing.T) {
	good := []string{"alice", "bob_99", "a.b-c", "ABC123", "___"}
	for _, u := range good {
		if !validUsername(u) {
			t.Errorf("expected %q to be valid", u)
		}
	}
	bad := []string{"", "ab", strings.Repeat("x", 33), "has space", "emoji😀", "semi;colon"}
	for _, u := range bad {
		if validUsername(u) {
			t.Errorf("expected %q to be invalid", u)
		}
	}
}

func TestNewFlowID(t *testing.T) {
	a, err := newFlowID()
	if err != nil {
		t.Fatalf("newFlowID: %v", err)
	}
	b, _ := newFlowID()
	if a == b {
		t.Fatal("flow ids should be unique")
	}
	if len(a) != 48 { // 24 bytes hex
		t.Fatalf("flow id length = %d, want 48", len(a))
	}
}

func TestDecodeJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"username":"x"}`))
	var dst struct {
		Username string `json:"username"`
	}
	if !decodeJSON(rec, req, &dst) {
		t.Fatal("expected decode success")
	}
	if dst.Username != "x" {
		t.Fatalf("username = %q", dst.Username)
	}

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`not json`))
	if decodeJSON(rec2, req2, &dst) {
		t.Fatal("expected decode failure")
	}
	if rec2.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec2.Code)
	}
}

func TestExtractToken(t *testing.T) {
	h := &Handlers{set: Settings{CookieName: "gotunnels_session"}}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer abc.def")
	if got := h.extractToken(req); got != "abc.def" {
		t.Fatalf("bearer token = %q", got)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.AddCookie(&http.Cookie{Name: "gotunnels_session", Value: "cookie-token"})
	if got := h.extractToken(req2); got != "cookie-token" {
		t.Fatalf("cookie token = %q", got)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	if got := h.extractToken(req3); got != "" {
		t.Fatalf("expected empty token, got %q", got)
	}
}
