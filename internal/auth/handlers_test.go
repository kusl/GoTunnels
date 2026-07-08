package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"
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

// TestComputeSessionExpiry pins the mapping from configured TTL to the DB
// expiry and cookie lifetime. The zero/negative cases are the "never expires"
// contract: a nil DB expiry (stored as SQL NULL) and a far-future,
// positively-aged cookie so the session survives a browser restart.
func TestComputeSessionExpiry(t *testing.T) {
	now := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)

	t.Run("never expires when ttl is zero", func(t *testing.T) {
		db, cookieExp, maxAge := computeSessionExpiry(now, 0)
		if db != nil {
			t.Errorf("dbExpiry = %v, want nil (NULL / never expires)", db)
		}
		if maxAge <= 0 {
			t.Errorf("cookie MaxAge = %d, want a large positive value", maxAge)
		}
		if !cookieExp.After(now.AddDate(0, 6, 0)) {
			t.Errorf("cookie expiry %v should be many months in the future", cookieExp)
		}
		if want := int(persistentSessionCookieTTL.Seconds()); maxAge != want {
			t.Errorf("cookie MaxAge = %d, want %d (persistent)", maxAge, want)
		}
	})

	t.Run("never expires when ttl is negative", func(t *testing.T) {
		db, _, maxAge := computeSessionExpiry(now, -time.Hour)
		if db != nil {
			t.Errorf("dbExpiry = %v, want nil for a non-positive ttl", db)
		}
		if maxAge <= 0 {
			t.Errorf("cookie MaxAge = %d, want positive", maxAge)
		}
	})

	t.Run("absolute expiry when ttl is positive", func(t *testing.T) {
		ttl := 48 * time.Hour
		db, cookieExp, maxAge := computeSessionExpiry(now, ttl)
		if db == nil {
			t.Fatal("dbExpiry = nil, want an absolute time for a positive ttl")
		}
		if !db.Equal(now.Add(ttl)) {
			t.Errorf("dbExpiry = %v, want %v", *db, now.Add(ttl))
		}
		if !cookieExp.Equal(now.Add(ttl)) {
			t.Errorf("cookie expiry = %v, want %v (aligned with DB expiry)", cookieExp, now.Add(ttl))
		}
		if maxAge != int(ttl.Seconds()) {
			t.Errorf("cookie MaxAge = %d, want %d", maxAge, int(ttl.Seconds()))
		}
	})
}

// TestSessionCookieAttributes pins the security-relevant cookie attributes.
// The frontend and API live on different origins, so the cookie must be
// SameSite=None; that in turn requires Secure, which tracks CookieSecure.
func TestSessionCookieAttributes(t *testing.T) {
	h := &Handlers{set: Settings{CookieName: "gotunnels_session", CookieSecure: true}}
	exp := time.Now().Add(persistentSessionCookieTTL)
	c := h.sessionCookie("tok-123", exp, int(persistentSessionCookieTTL.Seconds()))

	if c.Name != "gotunnels_session" {
		t.Errorf("cookie name = %q", c.Name)
	}
	if c.Value != "tok-123" {
		t.Errorf("cookie value = %q", c.Value)
	}
	if c.Path != "/" {
		t.Errorf("cookie path = %q, want /", c.Path)
	}
	if !c.HttpOnly {
		t.Error("session cookie must be HttpOnly")
	}
	if c.SameSite != http.SameSiteNoneMode {
		t.Errorf("SameSite = %v, want None (cross-site frontend/API)", c.SameSite)
	}
	if !c.Secure {
		t.Error("Secure must be true when CookieSecure is set (required for SameSite=None)")
	}
	if c.MaxAge <= 0 {
		t.Errorf("MaxAge = %d, want positive for a persistent cookie", c.MaxAge)
	}

	// When CookieSecure is false (local dev over http), Secure must follow.
	insecure := &Handlers{set: Settings{CookieName: "gotunnels_session", CookieSecure: false}}
	if insecure.sessionCookie("t", exp, 100).Secure {
		t.Error("Secure should be false when CookieSecure is false")
	}
}

// TestExpiredCookieClears pins that logout / logout-all emit a cookie the
// browser will delete immediately: empty value and a negative Max-Age.
func TestExpiredCookieClears(t *testing.T) {
	h := &Handlers{set: Settings{CookieName: "gotunnels_session", CookieSecure: true}}
	c := h.expiredCookie()

	if c.Value != "" {
		t.Errorf("cleared cookie value = %q, want empty", c.Value)
	}
	if c.MaxAge >= 0 {
		t.Errorf("cleared cookie MaxAge = %d, want negative (delete now)", c.MaxAge)
	}
	if c.Name != "gotunnels_session" {
		t.Errorf("cleared cookie name = %q", c.Name)
	}
	if c.Path != "/" {
		t.Errorf("cleared cookie path = %q, want / so it overwrites the original", c.Path)
	}
}
