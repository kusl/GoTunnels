package httpx

import (
	"context"
	"encoding/json/v2"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
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

func TestMatchOriginPattern(t *testing.T) {
	cases := []struct {
		pattern, origin string
		want            bool
	}{
		// The shape this exists for: Quick Tunnel hostnames.
		{"https://*.trycloudflare.com", "https://tart-panda-tea-cup.trycloudflare.com", true},
		{"https://*.trycloudflare.com", "https://A-B-C.TRYCLOUDFLARE.COM", true}, // case-insensitive
		// Exactly one label: zero or two labels must not match.
		{"https://*.trycloudflare.com", "https://trycloudflare.com", false},
		{"https://*.trycloudflare.com", "https://a.b.trycloudflare.com", false},
		// Scheme must match; http downgrade must not.
		{"https://*.trycloudflare.com", "http://a.trycloudflare.com", false},
		// Suffix tricks: the pattern's dot must be a real boundary.
		{"https://*.trycloudflare.com", "https://eviltrycloudflare.com", false},
		{"https://*.trycloudflare.com", "https://a.trycloudflare.com.evil.example", false},
		// No path/port/userinfo/query in an Origin.
		{"https://*.trycloudflare.com", "https://a.trycloudflare.com/path", false},
		{"https://*.trycloudflare.com", "https://a.trycloudflare.com:8443", false},
		{"https://*.trycloudflare.com", "https://user@a.trycloudflare.com", false},
		// Non-wildcard patterns match exactly (case-insensitively) only.
		{"https://app.example", "https://app.example", true},
		{"https://app.example", "https://APP.EXAMPLE", true},
		{"https://app.example", "https://other.example", false},
		// Garbage patterns/origins never match.
		{"", "https://a.trycloudflare.com", false},
		{"https://*.trycloudflare.com", "", false},
		{"not-an-origin", "https://a.trycloudflare.com", false},
	}
	for _, tc := range cases {
		if got := MatchOriginPattern(tc.pattern, tc.origin); got != tc.want {
			t.Errorf("MatchOriginPattern(%q, %q) = %v, want %v", tc.pattern, tc.origin, got, tc.want)
		}
	}
}

func TestOriginAllowedWildcardEntries(t *testing.T) {
	allowed := []string{"https://exact.example", "https://*.trycloudflare.com"}
	if echo, ok := OriginAllowed(allowed, false, "https://a-b.trycloudflare.com"); !ok || echo != "https://a-b.trycloudflare.com" {
		t.Fatalf("wildcard entry should match and echo the request origin: %q %v", echo, ok)
	}
	if _, ok := OriginAllowed(allowed, false, "https://trycloudflare.com"); ok {
		t.Fatal("apex must not match a *. pattern")
	}
	if echo, ok := OriginAllowed(allowed, false, "https://exact.example"); !ok || echo != "https://exact.example" {
		t.Fatalf("exact entry should still match: %q %v", echo, ok)
	}
}

func TestLimitWithMessageAndStatus(t *testing.T) {
	rl := NewRateLimiter(0, 1) // no refill, single token
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := rl.LimitWith(func(*http.Request) string { return "k" }, "custom limit message")(next)

	req := httptest.NewRequest(http.MethodPost, "/api/signup", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("first request status = %d, want 200", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request status = %d, want 429", rec2.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec2.Body.Bytes(), &body); err != nil {
		t.Fatalf("429 body not json: %v", err)
	}
	if body["error"] != "custom limit message" {
		t.Fatalf("429 message = %q, want the caller-supplied one", body["error"])
	}
}

func TestChainOrderOutermostFirst(t *testing.T) {
	var order []string
	mk := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	}), mk("outer"), mk("inner"))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if len(order) != 3 || order[0] != "outer" || order[1] != "inner" || order[2] != "handler" {
		t.Fatalf("execution order = %v, want [outer inner handler]", order)
	}
}

func TestRequestIDRejectsUnsafeInbound(t *testing.T) {
	for _, bad := range []string{"has space", "line\nbreak", strings.Repeat("a", maxRequestIDLen+1), "quote\""} {
		var seen string
		h := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			seen = RequestIDFromContext(r.Context())
		}))
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Request-Id", bad)
		h.ServeHTTP(httptest.NewRecorder(), req)
		if seen == bad || seen == "" {
			t.Errorf("inbound id %q must be replaced, got %q", bad, seen)
		}
	}
}

func TestRateLimiterSweepsFullBuckets(t *testing.T) {
	now := time.Unix(1_000, 0)
	rl := NewRateLimiter(1, 2)
	rl.now = func() time.Time { return now }
	for i := range 100 {
		rl.Allow(strconv.Itoa(i))
	}
	if rl.Len() != 100 {
		t.Fatalf("Len = %d, want 100", rl.Len())
	}
	now = now.Add(2 * sweepInterval)
	rl.Allow("fresh")
	if rl.Len() != 1 {
		t.Fatalf("after sweep Len = %d, want 1 (only the fresh key)", rl.Len())
	}
}

func TestRateLimiterSweepKeepsDrainedBuckets(t *testing.T) {
	now := time.Unix(1_000, 0)
	rl := NewRateLimiter(0.001, 1)
	rl.now = func() time.Time { return now }
	rl.Allow("hot")
	now = now.Add(2 * sweepInterval)
	if rl.Allow("hot") {
		t.Fatal("a drained bucket must survive the sweep and keep limiting")
	}
}

func TestSecurityHeaders(t *testing.T) {
	h := SecurityHeaders()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	for k, v := range map[string]string{
		"X-Content-Type-Options": "nosniff",
		"Cache-Control":          "no-store",
		"Referrer-Policy":        "no-referrer",
	} {
		if got := rec.Header().Get(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
}

func TestRecovererRepanicsAbortHandler(t *testing.T) {
	h := Recoverer(slog.New(slog.DiscardHandler))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(http.ErrAbortHandler)
	}))
	defer func() {
		if rec := recover(); rec != http.ErrAbortHandler {
			t.Fatalf("expected ErrAbortHandler to propagate, got %v", rec)
		}
	}()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
}

func TestCORSAlwaysVaryOrigin(t *testing.T) {
	h := CORS([]string{"https://ok.example"}, false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Header().Get("Vary") != "Origin" {
		t.Fatalf("Vary = %q, want Origin even for disallowed origins", rec.Header().Get("Vary"))
	}
}

func TestDecodeJSONRejectsTrailingData(t *testing.T) {
	var dst map[string]any
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"a":1}{"b":2}`))
	if DecodeJSON(rec, req, &dst, 0) {
		t.Fatal("trailing JSON must be rejected")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestWriteJSONDeterministicAndEmptySlices(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSON(rec, http.StatusOK, map[string]any{"b": 1, "a": []string(nil), "c": map[string]int(nil)})
	if got, want := rec.Body.String(), `{"a":[],"b":1,"c":{}}`+"\n"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestWriteJSONEncodingFailureIs500(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSON(rec, http.StatusOK, map[string]any{"d": time.Second})
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	var e map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil || e["error"] == "" {
		t.Fatalf("body = %q (%v), want a JSON error", rec.Body.String(), err)
	}
}

func decodeInto(t *testing.T, body string, maxBytes int64) (*httptest.ResponseRecorder, bool) {
	t.Helper()
	var dst struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	return rec, DecodeJSON(rec, req, &dst, maxBytes)
}

func errorMessage(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var e map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil {
		t.Fatalf("error body not json: %v", err)
	}
	return e["error"]
}

func TestDecodeJSONStrictness(t *testing.T) {
	cases := []struct {
		name, body string
		status     int
		msg        string
	}{
		{"unknown member", `{"name":"x","nmae":"y"}`, http.StatusBadRequest, `unknown field "nmae"`},
		{"case mismatch is unknown", `{"Name":"x"}`, http.StatusBadRequest, `unknown field "Name"`},
		{"duplicate name", `{"name":"x","name":"y"}`, http.StatusBadRequest, "invalid JSON body"},
		{"invalid utf-8", "{\"name\":\"\xff\"}", http.StatusBadRequest, "invalid JSON body"},
		{"wrong type", `{"count":"seven"}`, http.StatusBadRequest, `invalid value for "count"`},
		{"empty body", ``, http.StatusBadRequest, "invalid JSON body"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec, ok := decodeInto(t, tc.body, 0)
			if ok {
				t.Fatal("decode succeeded, want failure")
			}
			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d", rec.Code, tc.status)
			}
			if got := errorMessage(t, rec); got != tc.msg {
				t.Fatalf("message = %q, want %q", got, tc.msg)
			}
		})
	}
}

func TestDecodeJSONAcceptsExactPayload(t *testing.T) {
	if rec, ok := decodeInto(t, " {\"name\":\"x\",\"count\":3}\n", 0); !ok {
		t.Fatalf("decode failed: %d %s", rec.Code, rec.Body.String())
	}
}

func TestDecodeJSONTooLargeIs413(t *testing.T) {
	rec, ok := decodeInto(t, `{"name":"`+strings.Repeat("a", 100)+`"}`, 32)
	if ok {
		t.Fatal("oversized body accepted")
	}
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", rec.Code)
	}
}

func TestMaxBodyCapsReads(t *testing.T) {
	var readErr error
	h := MaxBody(8)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, readErr = io.ReadAll(r.Body)
	}))
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("x", 64)))
	h.ServeHTTP(httptest.NewRecorder(), req)
	if _, ok := errors.AsType[*http.MaxBytesError](readErr); !ok {
		t.Fatalf("read error = %v, want *http.MaxBytesError", readErr)
	}
}
