package httpx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOriginAllowed(t *testing.T) {
	allowed := []string{"https://a.example", "https://b.example"}

	if echo, ok := OriginAllowed(allowed, false, "https://a.example"); !ok || echo != "https://a.example" {
		t.Fatalf("exact match failed: %q %v", echo, ok)
	}
	if _, ok := OriginAllowed(allowed, false, "https://evil.example"); ok {
		t.Fatal("non-listed origin should be rejected")
	}
	if _, ok := OriginAllowed(allowed, false, ""); ok {
		t.Fatal("empty origin should be rejected")
	}
	if echo, ok := OriginAllowed(nil, true, "https://anything.example"); !ok || echo != "https://anything.example" {
		t.Fatalf("allowAny should echo origin: %q %v", echo, ok)
	}
	if _, ok := OriginAllowed(nil, true, ""); ok {
		t.Fatal("allowAny with empty origin should still be rejected")
	}
}

func TestCORSPreflight(t *testing.T) {
	h := CORS([]string{"https://a.example"}, false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("preflight should not reach the next handler")
	}))
	req := httptest.NewRequest(http.MethodOptions, "/api/x", nil)
	req.Header.Set("Origin", "https://a.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://a.example" {
		t.Fatalf("ACAO = %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("ACAC = %q", got)
	}
}

func TestCORSDisallowedOriginNoHeaders(t *testing.T) {
	reached := false
	h := CORS([]string{"https://a.example"}, false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
	req.Header.Set("Origin", "https://evil.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !reached {
		t.Fatal("non-preflight request should still reach handler")
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("must not set ACAO for disallowed origin")
	}
}

func TestRateLimiterBurstAndRefill(t *testing.T) {
	rl := NewRateLimiter(1, 2) // 1 tok/sec, burst 2
	current := time.Unix(0, 0)
	rl.now = func() time.Time { return current }

	if !rl.Allow("k") {
		t.Fatal("first request should pass")
	}
	if !rl.Allow("k") {
		t.Fatal("second request within burst should pass")
	}
	if rl.Allow("k") {
		t.Fatal("third request should be limited")
	}
	// advance 1 second -> 1 token back
	current = current.Add(time.Second)
	if !rl.Allow("k") {
		t.Fatal("request after refill should pass")
	}
	if rl.Allow("k") {
		t.Fatal("no tokens left after single refill")
	}
}

func TestRateLimiterPerKeyIsolation(t *testing.T) {
	rl := NewRateLimiter(0, 1)
	if !rl.Allow("a") {
		t.Fatal("key a first call should pass")
	}
	if rl.Allow("a") {
		t.Fatal("key a second call should be limited")
	}
	if !rl.Allow("b") {
		t.Fatal("key b should have its own bucket")
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	var seen string
	h := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if seen == "" {
		t.Fatal("expected a generated request id in context")
	}
	if rec.Header().Get("X-Request-Id") != seen {
		t.Fatal("response header should match context id")
	}
}

func TestRequestIDMiddlewareReusesInbound(t *testing.T) {
	var seen string
	h := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = RequestIDFromContext(r.Context())
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Id", "abc123")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if seen != "abc123" {
		t.Fatalf("expected inbound id reused, got %q", seen)
	}
}

func TestRequestIDFromContextEmpty(t *testing.T) {
	if RequestIDFromContext(context.Background()) != "" {
		t.Fatal("expected empty id for bare context")
	}
}

func TestWriteJSONAndError(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSON(rec, http.StatusTeapot, map[string]int{"n": 7})
	if rec.Code != http.StatusTeapot {
		t.Fatalf("status = %d", rec.Code)
	}
	var got map[string]int
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("body not json: %v", err)
	}
	if got["n"] != 7 {
		t.Fatalf("payload = %v", got)
	}

	rec2 := httptest.NewRecorder()
	WriteError(rec2, http.StatusBadRequest, "nope")
	var e map[string]string
	_ = json.Unmarshal(rec2.Body.Bytes(), &e)
	if e["error"] != "nope" {
		t.Fatalf("error body = %v", e)
	}
}
