package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
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

func TestFlowEnvelopeRoundTrip(t *testing.T) {
	sess := json.RawMessage(`{"challenge":"abc","user_id":"aWQ="}`)
	env := flowEnvelope{
		V:       1,
		Origin:  "https://tart-panda.trycloudflare.com",
		Signup:  &signupFlowData{UserID: "11111111-2222-4333-8444-555555555555", Username: "alice", DisplayName: "Alice"},
		Session: sess,
	}
	blob, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := decodeFlowEnvelope(blob)
	if got.V != 1 || got.Origin != env.Origin {
		t.Fatalf("roundtrip lost header fields: %+v", got)
	}
	if got.Signup == nil || got.Signup.Username != "alice" || got.Signup.UserID != env.Signup.UserID {
		t.Fatalf("roundtrip lost signup data: %+v", got.Signup)
	}
	if string(got.Session) != string(sess) {
		t.Fatalf("roundtrip changed session blob: %s", got.Session)
	}
}

func TestDecodeFlowEnvelopeLegacyBlob(t *testing.T) {
	// Flows stored before the envelope existed were the bare
	// webauthn.SessionData JSON. They must wrap on the fly with no pinned
	// origin (which resolves to the static relying party).
	legacy := []byte(`{"challenge":"xyz","user_id":"aWQ=","expires":"0001-01-01T00:00:00Z"}`)
	env := decodeFlowEnvelope(legacy)
	if env.Origin != "" {
		t.Fatalf("legacy blob must carry no origin, got %q", env.Origin)
	}
	if env.Signup != nil {
		t.Fatal("legacy blob must carry no signup data")
	}
	if string(env.Session) != string(legacy) {
		t.Fatalf("legacy blob must become the session verbatim, got %s", env.Session)
	}
	if _, err := env.sessionData(); err != nil {
		t.Fatalf("legacy session must still unpack: %v", err)
	}
}

func TestDecodeFlowEnvelopeGarbage(t *testing.T) {
	env := decodeFlowEnvelope([]byte("not json at all"))
	if string(env.Session) != "not json at all" {
		t.Fatalf("garbage should be preserved as the session blob, got %q", env.Session)
	}
	if _, err := env.sessionData(); err == nil {
		t.Fatal("unpacking a garbage session must error")
	}
}

func TestNewUUIDv4(t *testing.T) {
	seen := map[string]bool{}
	re := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	for i := 0; i < 64; i++ {
		id, err := newUUIDv4()
		if err != nil {
			t.Fatalf("newUUIDv4: %v", err)
		}
		if !re.MatchString(id) {
			t.Fatalf("id %q is not a canonical v4 uuid", id)
		}
		if seen[id] {
			t.Fatalf("duplicate uuid generated: %s", id)
		}
		seen[id] = true
	}
}
