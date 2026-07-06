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

func TestParseAuthors(t *testing.T) {
	u1 := "11111111-2222-4333-8444-555555555555"
	u2 := "AAAAAAAA-BBBB-4CCC-8DDD-EEEEEEEEEEEE"

	t.Run("empty means no filter", func(t *testing.T) {
		ids, problem := parseAuthors("")
		if problem != "" || ids != nil {
			t.Fatalf("got (%v, %q), want (nil, \"\")", ids, problem)
		}
	})
	t.Run("whitespace only means no filter", func(t *testing.T) {
		ids, problem := parseAuthors("   ")
		if problem != "" || ids != nil {
			t.Fatalf("got (%v, %q), want (nil, \"\")", ids, problem)
		}
	})
	t.Run("parses, trims, lowercases", func(t *testing.T) {
		ids, problem := parseAuthors(" " + u1 + " , " + u2 + " ")
		if problem != "" {
			t.Fatalf("unexpected problem: %s", problem)
		}
		if len(ids) != 2 || ids[0] != u1 || ids[1] != strings.ToLower(u2) {
			t.Fatalf("ids = %v", ids)
		}
	})
	t.Run("skips empty segments", func(t *testing.T) {
		ids, problem := parseAuthors(u1 + ",," + u1)
		if problem != "" || len(ids) != 2 {
			t.Fatalf("got (%v, %q)", ids, problem)
		}
	})
	t.Run("rejects non-uuid entries", func(t *testing.T) {
		for _, raw := range []string{"alice", u1 + ",alice", "1; DROP TABLE notes", u1 + "x"} {
			if _, problem := parseAuthors(raw); problem == "" {
				t.Errorf("parseAuthors(%q) accepted, want rejection", raw)
			}
		}
	})
	t.Run("caps the list", func(t *testing.T) {
		parts := make([]string, maxAuthorFilter+1)
		for i := range parts {
			parts[i] = u1
		}
		if _, problem := parseAuthors(strings.Join(parts, ",")); problem == "" {
			t.Fatal("over-cap list accepted, want rejection")
		}
	})
}

func TestIsUUIDString(t *testing.T) {
	good := []string{
		"11111111-2222-4333-8444-555555555555",
		"AAAAAAAA-BBBB-4CCC-8DDD-EEEEEEEEEEEE",
		"00000000-0000-0000-0000-000000000000",
	}
	for _, s := range good {
		if !isUUIDString(s) {
			t.Errorf("expected %q to be a uuid", s)
		}
	}
	bad := []string{
		"",
		"11111111-2222-4333-8444-55555555555",   // too short
		"11111111-2222-4333-8444-5555555555556", // too long
		"11111111x2222-4333-8444-555555555555",  // wrong separator
		"1111111g-2222-4333-8444-555555555555",  // non-hex
		"11111111-2222-4333-8444_555555555555",  // underscore
	}
	for _, s := range bad {
		if isUUIDString(s) {
			t.Errorf("expected %q to be rejected", s)
		}
	}
}
