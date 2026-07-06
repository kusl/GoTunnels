// Package notes implements a deliberately plain public microblog: short
// plain-text posts, visible to every signed-in user, hard-deletable only by
// their author, and never editable.
//
// "Plain text" is enforced on both sides of the wire. The server normalises
// line endings and rejects control characters (other than newline and tab) so
// stored bodies are clean, copyable text; the client renders bodies with
// textContent, so nothing is ever parsed as HTML and URLs stay inert strings.
// There is no edit endpoint on purpose — a post is either exactly what its
// author wrote, or gone.
package notes

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/kusl/GoTunnels/internal/auth"
	"github.com/kusl/GoTunnels/internal/httpx"
	"github.com/kusl/GoTunnels/internal/store"
)

// scopeName is the instrumentation scope for telemetry emitted here.
const scopeName = "github.com/kusl/GoTunnels/internal/notes"

const (
	// MaxBodyChars is the maximum note length in characters (code points),
	// mirrored by the notes_body_len CHECK constraint in the database.
	MaxBodyChars = 500
	// defaultPageSize is how many notes one list request returns by default.
	defaultPageSize = 50
	// maxPageSize caps a caller-supplied limit.
	maxPageSize = 200
)

// Handlers bundles dependencies for the notes endpoints.
type Handlers struct {
	store *store.Store
	log   *slog.Logger

	created metric.Int64Counter
	deleted metric.Int64Counter
}

// NewHandlers builds the handler set and registers its OTel instruments.
func NewHandlers(s *store.Store, log *slog.Logger) *Handlers {
	h := &Handlers{store: s, log: log}
	m := otel.Meter(scopeName)
	var err error
	if h.created, err = m.Int64Counter("gotunnels.notes.created",
		metric.WithDescription("Notes created"),
		metric.WithUnit("{note}")); err != nil {
		log.Warn("notes: register created counter", slog.String("error", err.Error()))
	}
	if h.deleted, err = m.Int64Counter("gotunnels.notes.deleted",
		metric.WithDescription("Notes deleted by their author"),
		metric.WithUnit("{note}")); err != nil {
		log.Warn("notes: register deleted counter", slog.String("error", err.Error()))
	}
	return h
}

type createRequest struct {
	Body string `json:"body"`
}

// List returns notes newest-first: the latest page by default, or — with
// ?before=<id> — the page strictly older than that id ("load older"). The
// same endpoint serves the page's poll-based auto-refresh, which simply
// re-fetches the newest window and reconciles client-side.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.CurrentUser(r.Context()); !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var beforeID int64
	if v := r.URL.Query().Get("before"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n < 0 {
			httpx.WriteError(w, http.StatusBadRequest, "invalid before cursor")
			return
		}
		beforeID = n
	}
	limit := defaultPageSize
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > maxPageSize {
			httpx.WriteError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = n
	}
	authorIDs, problem := parseAuthors(r.URL.Query().Get("authors"))
	if problem != "" {
		httpx.WriteError(w, http.StatusBadRequest, problem)
		return
	}

	rows, err := h.store.ListNotes(r.Context(), beforeID, limit, authorIDs)
	if err != nil {
		h.serverError(w, r, "notes: list", err)
		return
	}
	if rows == nil {
		rows = []store.Note{}
	}
	trace.SpanFromContext(r.Context()).SetAttributes(
		attribute.Int("notes.returned", len(rows)),
		attribute.Int64("notes.before", beforeID),
		attribute.Int("notes.author_filter", len(authorIDs)),
	)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"notes": rows})
}

// Authors lists everyone who currently has notes, for the author-filter
// dropdown on the notes page.
func (h *Handlers) Authors(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.CurrentUser(r.Context()); !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	authors, err := h.store.ListNoteAuthors(r.Context())
	if err != nil {
		h.serverError(w, r, "notes: authors", err)
		return
	}
	if authors == nil {
		authors = []store.NoteAuthor{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"authors": authors})
}

// Create posts a new note for the caller.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var req createRequest
	if !httpx.DecodeJSON(w, r, &req, 64<<10) {
		return
	}
	body, problem := ValidateBody(req.Body)
	if problem != "" {
		httpx.WriteError(w, http.StatusBadRequest, problem)
		return
	}

	n, err := h.store.CreateNote(r.Context(), user.ID, body)
	if err != nil {
		h.serverError(w, r, "notes: create", err)
		return
	}
	if h.created != nil {
		h.created.Add(r.Context(), 1)
	}
	trace.SpanFromContext(r.Context()).SetAttributes(
		attribute.Int64("note.id", n.ID),
		attribute.Int("note.chars", utf8.RuneCountInString(n.Body)),
	)
	httpx.WriteJSON(w, http.StatusCreated, map[string]any{"note": n})
}

// Delete removes the caller's own note. Ownership is enforced in the store's
// single DELETE statement; both "no such note" and "someone else's note"
// uniformly produce 404 so the endpoint is not an existence oracle.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}
	removed, err := h.store.DeleteNote(r.Context(), id, user.ID)
	if err != nil {
		h.serverError(w, r, "notes: delete", err)
		return
	}
	if !removed {
		httpx.WriteError(w, http.StatusNotFound, "note not found")
		return
	}
	if h.deleted != nil {
		h.deleted.Add(r.Context(), 1)
	}
	trace.SpanFromContext(r.Context()).SetAttributes(attribute.Int64("note.id", id))
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ValidateBody normalises and validates a raw note body. It returns the
// cleaned body and an empty string on success, or a zero body and a
// human-readable problem on failure.
//
// Normalisation: CRLF and lone CR become LF, then surrounding whitespace is
// trimmed. Validation: valid UTF-8, 1..MaxBodyChars characters, and no control
// characters other than newline and tab (keeps posts as honest, copyable plain
// text — no zero-width or terminal-escape mischief).
func ValidateBody(raw string) (string, string) {
	if !utf8.ValidString(raw) {
		return "", "note body must be valid UTF-8"
	}
	body := strings.ReplaceAll(raw, "\r\n", "\n")
	body = strings.ReplaceAll(body, "\r", "\n")
	body = strings.TrimSpace(body)
	if body == "" {
		return "", "note body must not be empty"
	}
	if n := utf8.RuneCountInString(body); n > MaxBodyChars {
		return "", fmt.Sprintf("note body must be at most %d characters (got %d)", MaxBodyChars, n)
	}
	for _, r := range body {
		if r < 0x20 && r != '\n' && r != '\t' {
			return "", "note body must not contain control characters"
		}
		if r == 0x7f {
			return "", "note body must not contain control characters"
		}
	}
	return body, ""
}

// maxAuthorFilter caps how many author ids one request may filter by.
const maxAuthorFilter = 50

// parseAuthors parses the ?authors= query parameter: a comma-separated list
// of user uuids. Empty input means "no filter". Every entry must look like a
// uuid — the ids go into a `$n::uuid[]` bind, and a malformed value would
// otherwise surface as an opaque database cast error instead of a 400.
func parseAuthors(raw string) ([]string, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, ""
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !isUUIDString(p) {
			return nil, "invalid author id"
		}
		out = append(out, strings.ToLower(p))
	}
	if len(out) > maxAuthorFilter {
		return nil, fmt.Sprintf("at most %d authors may be selected", maxAuthorFilter)
	}
	return out, ""
}

// isUUIDString reports whether s is a canonical 8-4-4-4-12 hex uuid.
func isUUIDString(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			isHex := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
			if !isHex {
				return false
			}
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
