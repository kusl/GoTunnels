package csp

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestNormalizeLegacy(t *testing.T) {
	body := []byte(`{"csp-report":{
		"document-uri":"https://f.example/",
		"referrer":"",
		"blocked-uri":"https://evil.example/x.js",
		"violated-directive":"script-src 'self'",
		"effective-directive":"script-src",
		"original-policy":"default-src 'self'",
		"disposition":"report",
		"source-file":"https://f.example/app.js",
		"line-number":12,
		"column-number":34,
		"status-code":200,
		"script-sample":"eval(...)"
	}}`)
	reports, err := Normalize("application/csp-report", body)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	r := reports[0]
	if r.BlockedURI != "https://evil.example/x.js" {
		t.Fatalf("blocked uri = %q", r.BlockedURI)
	}
	if r.EffectiveDirective != "script-src" {
		t.Fatalf("effective directive = %q", r.EffectiveDirective)
	}
	if r.LineNumber != 12 || r.ColumnNumber != 34 || r.StatusCode != 200 {
		t.Fatalf("numeric fields wrong: %+v", r)
	}
	if r.ScriptSample != "eval(...)" {
		t.Fatalf("script sample = %q", r.ScriptSample)
	}
	if len(r.Raw) == 0 {
		t.Fatal("expected raw payload retained")
	}
}

func TestNormalizeReportingAPIArray(t *testing.T) {
	body := []byte(`[
	  {"type":"csp-violation","age":10,"url":"https://f.example/","user_agent":"UA",
	   "body":{"documentURL":"https://f.example/","blockedURL":"inline",
	           "effectiveDirective":"style-src-attr","originalPolicy":"default-src 'self'",
	           "disposition":"report","sourceFile":"https://f.example/","lineNumber":1,
	           "columnNumber":2,"statusCode":200,"sample":"color:red"}},
	  {"type":"deprecation","body":{"id":"x"}}
	]`)
	reports, err := Normalize("application/reports+json", body)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("expected only the csp-violation report, got %d", len(reports))
	}
	r := reports[0]
	if r.BlockedURI != "inline" {
		t.Fatalf("blocked uri = %q", r.BlockedURI)
	}
	if r.EffectiveDirective != "style-src-attr" {
		t.Fatalf("effective directive = %q", r.EffectiveDirective)
	}
	// ViolatedDirective falls back to effective when not provided.
	if r.ViolatedDirective != "style-src-attr" {
		t.Fatalf("violated directive fallback = %q", r.ViolatedDirective)
	}
	if r.ScriptSample != "color:red" {
		t.Fatalf("sample = %q", r.ScriptSample)
	}
}

func TestNormalizeCustom(t *testing.T) {
	body := []byte(`{
		"documentURI":"https://f.example/page",
		"blockedURI":"https://cdn.evil/x.js",
		"violatedDirective":"script-src-elem",
		"effectiveDirective":"script-src-elem",
		"originalPolicy":"default-src 'self'",
		"disposition":"report",
		"sourceFile":"https://f.example/page",
		"lineNumber":5,"columnNumber":9,"statusCode":0,"sample":""
	}`)
	reports, err := Normalize("application/json", body)
	if err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if reports[0].ViolatedDirective != "script-src-elem" {
		t.Fatalf("violated directive = %q", reports[0].ViolatedDirective)
	}
	if reports[0].DocumentURI != "https://f.example/page" {
		t.Fatalf("document uri = %q", reports[0].DocumentURI)
	}
}

func TestNormalizeEmpty(t *testing.T) {
	if _, err := Normalize("application/json", []byte("   ")); err == nil {
		t.Fatal("expected error for empty body")
	}
}

func TestNormalizeGarbage(t *testing.T) {
	if _, err := Normalize("application/json", []byte(`{"unrelated":true}`)); err == nil {
		t.Fatal("expected error for unrecognised object")
	}
}

func TestNormalizeReportingAPINoCSP(t *testing.T) {
	body := []byte(`[{"type":"deprecation","body":{"id":"x"}}]`)
	if _, err := Normalize("application/reports+json", body); err == nil {
		t.Fatal("expected error when no csp reports present")
	}
}

func TestMediaType(t *testing.T) {
	if got := mediaType("application/csp-report; charset=utf-8"); got != "application/csp-report" {
		t.Fatalf("mediaType = %q", got)
	}
	if got := mediaType("  Application/JSON "); got != "application/json" {
		t.Fatalf("mediaType = %q", got)
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if firstNonEmpty("", "", "c") != "c" {
		t.Fatal("expected first non-empty to be c")
	}
	if firstNonEmpty("", "") != "" {
		t.Fatal("expected empty when all empty")
	}
}

func TestTruncateRunes(t *testing.T) {
	cases := []struct {
		name string
		in   string
		n    int
		want string
	}{
		{"shorter passes through", "hello", 10, "hello"},
		{"exactly n passes through", "hello", 5, "hello"},
		{"ascii truncates with ellipsis", "hello world", 5, "hello…"},
		{"multibyte counts runes not bytes", "ééééé", 5, "ééééé"},
		{"multibyte truncates cleanly", "日本語テキスト", 3, "日本語…"},
		{"zero yields empty", "anything", 0, ""},
		{"negative yields empty", "anything", -1, ""},
		{"empty stays empty", "", 5, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := truncateRunes(tc.in, tc.n); got != tc.want {
				t.Fatalf("truncateRunes(%q, %d) = %q, want %q", tc.in, tc.n, got, tc.want)
			}
		})
	}
}

func TestTruncateRunesNeverSplitsUTF8(t *testing.T) {
	in := strings.Repeat("界", 300)
	got := truncateRunes(in, maxURILen)
	if !utf8.ValidString(got) {
		t.Fatal("truncation produced invalid UTF-8")
	}
	if utf8.RuneCountInString(got) != maxURILen+1 { // +1 for the ellipsis
		t.Fatalf("rune count = %d, want %d", utf8.RuneCountInString(got), maxURILen+1)
	}
}

func TestRecentRejectsBadLimit(t *testing.T) {
	// Limit validation runs before any store access, so a nil store is safe
	// here and keeps this a pure unit test.
	h := NewHandler(nil, slog.New(slog.NewTextHandler(io.Discard, nil)), nil)
	for _, q := range []string{"limit=0", "limit=-1", "limit=201", "limit=abc"} {
		req := httptest.NewRequest(http.MethodGet, "/api/csp-reports/recent?"+q, nil)
		rec := httptest.NewRecorder()
		h.Recent(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", q, rec.Code)
		}
	}
}
