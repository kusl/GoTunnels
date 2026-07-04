// Package prefs is a tiny per-user key/value preference store, exposed as
// GET/PUT /api/prefs/{key}. It exists so small UI choices (like whether the
// CAPTCHA leaderboard is expanded) follow the *account* rather than the
// browser. The dividing line used across the app: user preferences live here,
// on the server; device-specific settings (like the Magic Solve speed, which
// depends on the device's refresh rate) stay in localStorage.
//
// Keys are constrained to a conservative charset and values to a small size,
// so this can never quietly become a document store.
package prefs

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/kusl/GoTunnels/internal/auth"
	"github.com/kusl/GoTunnels/internal/httpx"
	"github.com/kusl/GoTunnels/internal/store"
)

const (
	// MaxKeyLen is the maximum preference key length.
	MaxKeyLen = 64
	// MaxValueBytes is the maximum preference value size, mirrored by the
	// CHECK constraint on user_prefs.value.
	MaxValueBytes = 4096
)

// Handlers bundles dependencies for the preference endpoints.
type Handlers struct {
	store *store.Store
	log   *slog.Logger
}

// NewHandlers builds the handler set.
func NewHandlers(s *store.Store, log *slog.Logger) *Handlers {
	return &Handlers{store: s, log: log}
}

type setRequest struct {
	Value string `json:"value"`
}

// Get returns the caller's stored value for a key. A missing key is a normal
// state, not an error: the response carries "exists": false so clients can
// fall back to defaults without special-casing 404s.
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	key := r.PathValue("key")
	if !ValidKey(key) {
		httpx.WriteError(w, http.StatusBadRequest, "invalid preference key")
		return
	}
	value, err := h.store.GetUserPref(r.Context(), user.ID, key)
	if errors.Is(err, store.ErrNotFound) {
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"key": key, "value": "", "exists": false})
		return
	}
	if err != nil {
		h.serverError(w, r, "prefs: get", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"key": key, "value": value, "exists": true})
}

// Set upserts the caller's value for a key.
func (h *Handlers) Set(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	key := r.PathValue("key")
	if !ValidKey(key) {
		httpx.WriteError(w, http.StatusBadRequest, "invalid preference key")
		return
	}
	var req setRequest
	if !httpx.DecodeJSON(w, r, &req, 16<<10) {
		return
	}
	if len(req.Value) > MaxValueBytes {
		httpx.WriteError(w, http.StatusBadRequest, "preference value too large")
		return
	}
	if err := h.store.SetUserPref(r.Context(), user.ID, key, req.Value); err != nil {
		h.serverError(w, r, "prefs: set", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

// ValidKey enforces the preference-key policy: 1..MaxKeyLen characters from
// [a-z0-9._-], and it must start with a letter or digit. Kept as a manual loop
// (matching validUsername in the auth package) rather than a regexp.
func ValidKey(k string) bool {
	if len(k) < 1 || len(k) > MaxKeyLen {
		return false
	}
	for i, r := range k {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '.' || r == '-' || r == '_':
			if i == 0 {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func (h *Handlers) serverError(w http.ResponseWriter, r *http.Request, msg string, err error) {
	h.log.ErrorContext(r.Context(), msg,
		slog.String("error", err.Error()),
		slog.String("request_id", httpx.RequestIDFromContext(r.Context())),
	)
	httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
}
