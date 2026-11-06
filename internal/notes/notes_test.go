package notes

import (
	"fmt"
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
		{"c1 csi escape", "sneaky \u009b31m red"},
		{"next line control", "a\u0085b"},
		{"bidi override", "abc\u202Edef"},
		{"bidi isolate", "abc\u2066def"},
		{"zero width space", "a\u200Bb"},
		{"byte order mark", "a\uFEFFb"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got, problem := ValidateBody(tc.in); problem == "" {
				t.Fatalf("ValidateBody(%q) accepted as %q, want rejection", tc.in, got)
			}
		})
	}
}

func TestValidateBodyKeepsEmojiJoiners(t *testing.T) {
	family := "family \U0001F468\u200D\U0001F469\u200D\U0001F467"
	if got, problem := ValidateBody(family); problem != "" || got != family {
		t.Fatalf("ZWJ emoji sequence rejected: %q", problem)
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
	t.Run("skips empty segments and duplicates", func(t *testing.T) {
		ids, problem := parseAuthors(u1 + ",," + u1 + "," + strings.ToUpper(u1))
		if problem != "" || len(ids) != 1 || ids[0] != u1 {
			t.Fatalf("got (%v, %q)", ids, problem)
		}
	})
	t.Run("canonicalises accepted uuid spellings", func(t *testing.T) {
		raw := "{" + u1 + "},urn:uuid:" + strings.ToUpper(u2) + "," + strings.ReplaceAll(u1, "-", "")
		ids, problem := parseAuthors(raw)
		if problem != "" || len(ids) != 2 || ids[0] != u1 || ids[1] != strings.ToLower(u2) {
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
}

func TestParseAuthorsCapCountsDistinctIDs(t *testing.T) {
	parts := make([]string, 0, maxAuthorFilter+1)
	for i := range maxAuthorFilter + 1 {
		parts = append(parts, fmt.Sprintf("00000000-0000-4000-8000-%012d", i))
	}
	if _, problem := parseAuthors(strings.Join(parts[:maxAuthorFilter], ",")); problem != "" {
		t.Fatalf("exactly-at-cap list rejected: %s", problem)
	}
	if _, problem := parseAuthors(strings.Join(parts, ",")); problem == "" {
		t.Fatal("over-cap list of distinct ids accepted, want rejection")
	}
}
