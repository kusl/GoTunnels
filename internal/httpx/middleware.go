// Package httpx holds transport-level helpers and middleware that do not
// depend on the data store: CORS, request identifiers, panic recovery, a small
// in-memory rate limiter, OpenTelemetry HTTP instrumentation, and JSON writing
// helpers. Keeping these store-agnostic makes them trivial to unit test.
package httpx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// ctxKey is a private context key type.
type ctxKey int

const requestIDKey ctxKey = iota

// Middleware is the standard http middleware shape.
type Middleware func(http.Handler) http.Handler

// Chain applies middleware in order (the first listed is outermost).
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// ---------------------------------------------------------------------------
// CORS
// ---------------------------------------------------------------------------

// OriginAllowed decides the Access-Control-Allow-Origin value to echo for a
// request Origin. Because the browser sends credentials (Bearer token flows
// still trigger CORS), the wildcard cannot be combined with credentials, so we
// echo a specific origin. When allowAny is true any non-empty Origin is echoed.
func OriginAllowed(allowed []string, allowAny bool, origin string) (echo string, ok bool) {
	if origin == "" {
		return "", false
	}
	if allowAny {
		return origin, true
	}
	for _, a := range allowed {
		if a == origin {
			return origin, true
		}
	}
	return "", false
}

// CORS returns middleware that applies the origin policy and answers
// preflight OPTIONS requests.
func CORS(allowed []string, allowAny bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if echo, ok := OriginAllowed(allowed, allowAny, origin); ok {
				h := w.Header()
				h.Set("Access-Control-Allow-Origin", echo)
				h.Set("Access-Control-Allow-Credentials", "true")
				h.Add("Vary", "Origin")
				h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
				h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				h.Set("Access-Control-Max-Age", "600")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ---------------------------------------------------------------------------
// Request ID
// ---------------------------------------------------------------------------

// RequestID attaches a request identifier (reusing an inbound X-Request-Id when
// present) to the context and response headers.
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-Id")
			if id == "" {
				id = newID()
			}
			w.Header().Set("X-Request-Id", id)
			ctx := context.WithValue(r.Context(), requestIDKey, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequestIDFromContext returns the request id, or "" if unset.
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}

// ---------------------------------------------------------------------------
// Recovery
// ---------------------------------------------------------------------------

// Recoverer converts panics into a 500 response and logs them.
func Recoverer(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.ErrorContext(r.Context(), "panic recovered",
						slog.Any("panic", rec),
						slog.String("path", r.URL.Path),
						slog.String("request_id", RequestIDFromContext(r.Context())),
					)
					WriteError(w, http.StatusInternalServerError, "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// ---------------------------------------------------------------------------
// OpenTelemetry HTTP instrumentation
// ---------------------------------------------------------------------------

// Instrument wraps a handler with otelhttp so every request produces a span and
// standard HTTP metrics. It is a no-op-friendly wrapper: when telemetry is
// disabled the global providers are no-ops, so this stays cheap.
func Instrument(h http.Handler, operation string) http.Handler {
	return otelhttp.NewHandler(h, operation)
}

// ---------------------------------------------------------------------------
// Rate limiting (token bucket per key)
// ---------------------------------------------------------------------------

// RateLimiter is a simple per-key token-bucket limiter safe for concurrent use.
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     float64 // tokens per second
	capacity float64
	now      func() time.Time // injectable clock for tests
}

type bucket struct {
	tokens float64
	last   time.Time
}

// NewRateLimiter builds a limiter allowing rate tokens/sec with the given burst.
func NewRateLimiter(ratePerSec, burst float64) *RateLimiter {
	if burst < 1 {
		burst = 1
	}
	return &RateLimiter{
		buckets:  map[string]*bucket{},
		rate:     ratePerSec,
		capacity: burst,
		now:      time.Now,
	}
}

// Allow reports whether an event for key is permitted right now.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := rl.now()
	b, ok := rl.buckets[key]
	if !ok {
		rl.buckets[key] = &bucket{tokens: rl.capacity - 1, last: now}
		return true
	}
	elapsed := now.Sub(b.last).Seconds()
	b.tokens += elapsed * rl.rate
	if b.tokens > rl.capacity {
		b.tokens = rl.capacity
	}
	b.last = now
	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// LimitByIP returns middleware that rate-limits using a key derived from the
// request (the provided keyFn, typically a hashed client IP). Rejected
// requests receive 429.
func (rl *RateLimiter) LimitByIP(keyFn func(*http.Request) string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.Allow(keyFn(r)) {
				WriteError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ---------------------------------------------------------------------------
// JSON helpers
// ---------------------------------------------------------------------------

// WriteJSON writes v as JSON with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(v)
}

// WriteError writes a JSON error body: {"error": "..."}.
func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

// DecodeJSON reads a size-limited JSON body into dst, writing a 400 and
// returning false on failure. maxBytes <= 0 defaults to 1 MiB.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any, maxBytes int64) bool {
	if maxBytes <= 0 {
		maxBytes = 1 << 20
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}
