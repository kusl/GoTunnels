package notes

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestValidateBodyAccepts(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "hello world", "hello world"},
		{"trims surrounding whitespace", "  hi there \n", "hi there"},
		{"keeps interior newlines", "line one\nline two", "line one\nline two"},
		{"normalises crlf", "a\r\nb", "a\nb"},
		{"normalises lone cr", "a\rb", "a\nb"},
		{"keeps tabs", "col1\tcol2", "col1\tcol2"},
		{"multibyte ok", "héllø 世界 🚀", "héllø 世界 🚀"},
		{"exactly max runes", strings.Repeat("界", MaxBodyChars), strings.Repeat("界", MaxBodyChars)},
		{"url stays plain text", "see https://example.com for more", "see https://example.com for more"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, problem := ValidateBody(tc.in)
			if problem != "" {
				t.Fatalf("ValidateBody(%q) rejected: %s", tc.in, problem)
			}
			if got != tc.want {
				t.Fatalf("ValidateBody(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestValidateBodyRejects(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"whitespace only", "   \n\t  "},
		{"over max runes", strings.Repeat("a", MaxBodyChars+1)},
		{"over max multibyte runes", strings.Repeat("界", MaxBodyChars+1)},
		{"escape sequence", "sneaky \x1b[31m red"},
		{"null byte", "a\x00b"},
		{"delete char", "a\x7fb"},
		{"invalid utf8", string([]byte{0xff, 0xfe, 0x41})},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got, problem := ValidateBody(tc.in); problem == "" {
				t.Fatalf("ValidateBody(%q) accepted as %q, want rejection", tc.in, got)
			}
		})
	}
}

func TestValidateBodyMaxCountsRunesNotBytes(t *testing.T) {
	// 500 multibyte runes is far more than 500 bytes but must be accepted.
	in := strings.Repeat("é", MaxBodyChars)
	if utf8.RuneCountInString(in) != MaxBodyChars {
		t.Fatal("test setup wrong")
	}
	if len(in) <= MaxBodyChars {
		t.Fatal("test setup wrong: want multibyte input")
	}
	if _, problem := ValidateBody(in); problem != "" {
		t.Fatalf("multibyte body at limit rejected: %s", problem)
	}
}
