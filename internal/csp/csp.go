// Package csp accepts Content-Security-Policy violation reports in the three
// shapes browsers actually send and normalises them into a single struct that
// is both stored in Postgres and logged through the OpenTelemetry-backed slog
// logger. The reporting endpoint is unauthenticated (browsers post it without
// credentials), so callers rate-limit it at the middleware layer.
package csp

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/kusl/GoTunnels/internal/activity"
	"github.com/kusl/GoTunnels/internal/httpx"
	"github.com/kusl/GoTunnels/internal/store"
)

// Normalize parses a CSP report request body into zero or more normalised
// reports. It understands:
//
//   - legacy application/csp-report: {"csp-report": {...}} with hyphenated keys
//   - the Reporting API application/reports+json: an array of {type,url,body}
//   - a custom camelCase body posted by our in-page violation listener
//
// The raw JSON of each individual report is preserved in Raw.
func Normalize(contentType string, body []byte) ([]store.CSPReportInput, error) {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return nil, errors.New("csp: empty body")
	}

	// Reporting API sends a JSON array.
	if trimmed[0] == '[' {
		return parseReportingAPI(body)
	}

	// Object: could be the legacy wrapper or our custom shape.
	if strings.Contains(mediaType(contentType), "csp-report") || looksLegacy(body) {
		if r, ok := parseLegacy(body); ok {
			return []store.CSPReportInput{r}, nil
		}
	}
	if r, ok := parseCustom(body); ok {
		return []store.CSPReportInput{r}, nil
	}
	return nil, errors.New("csp: unrecognised report shape")
}

func mediaType(ct string) string {
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = ct[:i]
	}
	return strings.ToLower(strings.TrimSpace(ct))
}

func looksLegacy(body []byte) bool {
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(body, &probe); err != nil {
		return false
	}
	_, ok := probe["csp-report"]
	return ok
}

// --- legacy application/csp-report --------------------------------------

type legacyWrapper struct {
	CSPReport struct {
		DocumentURI        string `json:"document-uri"`
		Referrer           string `json:"referrer"`
		BlockedURI         string `json:"blocked-uri"`
		ViolatedDirective  string `json:"violated-directive"`
		EffectiveDirective string `json:"effective-directive"`
		OriginalPolicy     string `json:"original-policy"`
		Disposition        string `json:"disposition"`
		SourceFile         string `json:"source-file"`
		ScriptSample       string `json:"script-sample"`
		LineNumber         int    `json:"line-number"`
		ColumnNumber       int    `json:"column-number"`
		StatusCode         int    `json:"status-code"`
	} `json:"csp-report"`
}

func parseLegacy(body []byte) (store.CSPReportInput, bool) {
	var w legacyWrapper
	if err := json.Unmarshal(body, &w); err != nil {
		return store.CSPReportInput{}, false
	}
	c := w.CSPReport
	return store.CSPReportInput{
		DocumentURI:        c.DocumentURI,
		Referrer:           c.Referrer,
		BlockedURI:         c.BlockedURI,
		ViolatedDirective:  c.ViolatedDirective,
		EffectiveDirective: c.EffectiveDirective,
		OriginalPolicy:     c.OriginalPolicy,
		Disposition:        c.Disposition,
		SourceFile:         c.SourceFile,
		LineNumber:         c.LineNumber,
		ColumnNumber:       c.ColumnNumber,
		StatusCode:         c.StatusCode,
		ScriptSample:       c.ScriptSample,
		Raw:                json.RawMessage(body),
	}, true
}

// --- Reporting API application/reports+json -----------------------------

type reportingItem struct {
	Type string `json:"type"`
	URL  string `json:"url"`
	Body struct {
		DocumentURL        string `json:"documentURL"`
		Referrer           string `json:"referrer"`
		BlockedURL         string `json:"blockedURL"`
		ViolatedDirective  string `json:"violatedDirective"`
		EffectiveDirective string `json:"effectiveDirective"`
		OriginalPolicy     string `json:"originalPolicy"`
		Disposition        string `json:"disposition"`
		SourceFile         string `json:"sourceFile"`
		Sample             string `json:"sample"`
		LineNumber         int    `json:"lineNumber"`
		ColumnNumber       int    `json:"columnNumber"`
		StatusCode         int    `json:"statusCode"`
	} `json:"body"`
}

func parseReportingAPI(body []byte) ([]store.CSPReportInput, error) {
	var items []reportingItem
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, err
	}
	var out []store.CSPReportInput
	for _, it := range items {
		// Ignore non-CSP report types that may share the endpoint.
		if it.Type != "" && it.Type != "csp-violation" && it.Type != "csp" {
			continue
		}
		b := it.Body
		raw, _ := json.Marshal(it)
		doc := b.DocumentURL
		if doc == "" {
			doc = it.URL
		}
		out = append(out, store.CSPReportInput{
			DocumentURI:        doc,
			Referrer:           b.Referrer,
			BlockedURI:         b.BlockedURL,
			ViolatedDirective:  firstNonEmpty(b.ViolatedDirective, b.EffectiveDirective),
			EffectiveDirective: b.EffectiveDirective,
			OriginalPolicy:     b.OriginalPolicy,
			Disposition:        b.Disposition,
			SourceFile:         b.SourceFile,
			LineNumber:         b.LineNumber,
			ColumnNumber:       b.ColumnNumber,
			StatusCode:         b.StatusCode,
			ScriptSample:       b.Sample,
			Raw:                json.RawMessage(raw),
		})
	}
	if len(out) == 0 {
		return nil, errors.New("csp: reporting-api payload contained no csp reports")
	}
	return out, nil
}

// --- custom in-page listener shape --------------------------------------

type customReport struct {
	DocumentURI        string `json:"documentURI"`
	Referrer           string `json:"referrer"`
	BlockedURI         string `json:"blockedURI"`
	ViolatedDirective  string `json:"violatedDirective"`
	EffectiveDirective string `json:"effectiveDirective"`
	OriginalPolicy     string `json:"originalPolicy"`
	Disposition        string `json:"disposition"`
	SourceFile         string `json:"sourceFile"`
	Sample             string `json:"sample"`
	LineNumber         int    `json:"lineNumber"`
	ColumnNumber       int    `json:"columnNumber"`
	StatusCode         int    `json:"statusCode"`
}

func parseCustom(body []byte) (store.CSPReportInput, bool) {
	var c customReport
	dec := json.NewDecoder(strings.NewReader(string(body)))
	if err := dec.Decode(&c); err != nil {
		return store.CSPReportInput{}, false
	}
	// Require at least one meaningful field to avoid storing empty objects.
	if c.DocumentURI == "" && c.ViolatedDirective == "" && c.EffectiveDirective == "" && c.BlockedURI == "" {
		return store.CSPReportInput{}, false
	}
	return store.CSPReportInput{
		DocumentURI:        c.DocumentURI,
		Referrer:           c.Referrer,
		BlockedURI:         c.BlockedURI,
		ViolatedDirective:  firstNonEmpty(c.ViolatedDirective, c.EffectiveDirective),
		EffectiveDirective: c.EffectiveDirective,
		OriginalPolicy:     c.OriginalPolicy,
		Disposition:        c.Disposition,
		SourceFile:         c.SourceFile,
		LineNumber:         c.LineNumber,
		ColumnNumber:       c.ColumnNumber,
		StatusCode:         c.StatusCode,
		ScriptSample:       c.Sample,
		Raw:                json.RawMessage(body),
	}, true
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// HTTP handler
// ---------------------------------------------------------------------------

// Handler persists and logs CSP violation reports.
type Handler struct {
	store  *store.Store
	log    *slog.Logger
	pepper []byte
}

// NewHandler builds a CSP report handler.
func NewHandler(s *store.Store, log *slog.Logger, pepper []byte) *Handler {
	return &Handler{store: s, log: log, pepper: pepper}
}

// ServeHTTP accepts a report POST. It always responds 204 once the body is
// read (a report endpoint should never give an attacker signal), but still
// logs and stores what it can.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const maxBody = 64 * 1024
	body, err := readAll(r, maxBody)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	reports, perr := Normalize(r.Header.Get("Content-Type"), body)
	if perr != nil {
		h.log.WarnContext(r.Context(), "csp report parse failed",
			slog.String("error", perr.Error()))
		w.WriteHeader(http.StatusNoContent)
		return
	}

	ipHash := activity.HashIP(h.pepper, activity.ClientIP(r))
	ua := r.UserAgent()
	for i := range reports {
		reports[i].IPHash = ipHash
		reports[i].UserAgent = ua
		if err := h.store.InsertCSPReport(r.Context(), reports[i]); err != nil {
			h.log.ErrorContext(r.Context(), "store csp report", slog.String("error", err.Error()))
		}
		h.log.InfoContext(r.Context(), "csp violation",
			slog.String("violated_directive", reports[i].ViolatedDirective),
			slog.String("blocked_uri", reports[i].BlockedURI),
			slog.String("document_uri", reports[i].DocumentURI),
			slog.String("disposition", reports[i].Disposition),
		)
	}
	w.WriteHeader(http.StatusNoContent)
}

func readAll(r *http.Request, max int64) ([]byte, error) {
	defer func() { _ = r.Body.Close() }()
	return io.ReadAll(io.LimitReader(r.Body, max))
}

// ---------------------------------------------------------------------------
// public transparency feed
// ---------------------------------------------------------------------------

const (
	// recentDefaultLimit / recentMaxLimit bound the public feed page size.
	recentDefaultLimit = 50
	recentMaxLimit     = 200
	// maxURILen / maxSampleLen cap how much of any reported string the public
	// feed echoes back. Report fields are attacker-controlled input (anyone
	// can POST to the report endpoint), so the public view truncates them:
	// long enough to learn from, short enough that the feed cannot be abused
	// as free hosting for kilobytes of arbitrary text.
	maxURILen    = 200
	maxSampleLen = 100
)

// Recent serves the newest CSP violations as a public, read-only feed for the
// passkeys/security explainer page. Rows come from the store already stripped
// of ip_hash, user_agent, original_policy, and the raw payload (see
// store.CSPReportRow); this handler additionally truncates the free-text
// fields. Publishing what the browser blocked teaches CSP by example without
// revealing anything about who triggered a report.
func (h *Handler) Recent(w http.ResponseWriter, r *http.Request) {
	limit := recentDefaultLimit
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > recentMaxLimit {
			httpx.WriteError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = n
	}
	rows, err := h.store.ListRecentCSPReports(r.Context(), limit)
	if err != nil {
		h.log.ErrorContext(r.Context(), "csp: list recent", slog.String("error", err.Error()))
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if rows == nil {
		rows = []store.CSPReportRow{}
	}
	for i := range rows {
		rows[i].DocumentURI = truncateRunes(rows[i].DocumentURI, maxURILen)
		rows[i].BlockedURI = truncateRunes(rows[i].BlockedURI, maxURILen)
		rows[i].SourceFile = truncateRunes(rows[i].SourceFile, maxURILen)
		rows[i].ScriptSample = truncateRunes(rows[i].ScriptSample, maxSampleLen)
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"reports": rows})
}

// truncateRunes shortens s to at most n characters (code points), appending an
// ellipsis when it cut something. Rune-aware so it never splits a multi-byte
// UTF-8 sequence into mojibake.
func truncateRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}
