// Package activity records audit events and enforces the privacy rule that IP
// addresses are never stored in the clear. Addresses are hashed as
// sha256(pepper || ip) and only the lowercase hex digest is persisted.
package activity

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"strings"

	"github.com/kusl/GoTunnels/internal/store"
)

// HashIP returns sha256(pepper || ip) as lowercase hex. An empty ip yields an
// empty string so "unknown" is distinguishable from a real hash.
func HashIP(pepper []byte, ip string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return ""
	}
	h := sha256.New()
	h.Write(pepper)
	h.Write([]byte(ip))
	return hex.EncodeToString(h.Sum(nil))
}

// ClientIP extracts the best-guess client IP from a request. Because the API
// only ever receives traffic through the Cloudflare tunnel, Cf-Connecting-Ip
// is authoritative when present; otherwise fall back to the first
// X-Forwarded-For hop and finally the transport remote address.
func ClientIP(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("Cf-Connecting-Ip")); v != "" {
		return v
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if first, _, ok := strings.Cut(xff, ","); ok {
			return strings.TrimSpace(first)
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return host
}

// Recorder writes audit events, hashing request metadata on the way in.
type Recorder struct {
	store  *store.Store
	pepper []byte
}

// NewRecorder builds a Recorder.
func NewRecorder(s *store.Store, pepper []byte) *Recorder {
	return &Recorder{store: s, pepper: pepper}
}

// Event is a partially-filled audit event; request-derived fields (ip hash,
// user agent) are populated by Record.
type Event struct {
	UserID     *string
	Username   string
	EventType  string
	AuthMethod string
	Outcome    string
	Detail     map[string]any
}

// Record hashes the request's client IP, captures the user agent, and persists
// the event. Recording failures are returned but are typically logged and
// swallowed by callers so audit problems never block an auth flow.
func (rec *Recorder) Record(ctx context.Context, r *http.Request, e Event) error {
	ipHash := HashIP(rec.pepper, ClientIP(r))
	return rec.store.InsertActivity(ctx, store.ActivityInput{
		UserID:     e.UserID,
		Username:   e.Username,
		EventType:  e.EventType,
		AuthMethod: e.AuthMethod,
		Outcome:    e.Outcome,
		IPHash:     ipHash,
		UserAgent:  r.UserAgent(),
		Detail:     e.Detail,
	})
}
