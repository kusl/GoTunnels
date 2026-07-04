package prefs

import "testing"

func TestValidKey(t *testing.T) {
	valid := []string{
		"captcha.leaderboard.open",
		"theme",
		"a",
		"notes.feed-order",
		"k1_2-3.four",
	}
	for _, k := range valid {
		if !ValidKey(k) {
			t.Errorf("ValidKey(%q) = false, want true", k)
		}
	}

	invalid := []string{
		"",                                // empty
		".starts-with-dot",                // must start alphanumeric
		"-starts-with-dash",               // must start alphanumeric
		"_starts_with_under",              // must start alphanumeric
		"Has.Uppercase",                   // uppercase not allowed
		"has space",                       // whitespace not allowed
		"has/slash",                       // slash not allowed
		"emoji.😀",                         // non-ASCII not allowed
		"trailing.newline\n",              // control char
		string(make([]byte, MaxKeyLen+1)), // too long
	}
	for _, k := range invalid {
		if ValidKey(k) {
			t.Errorf("ValidKey(%q) = true, want false", k)
		}
	}
}
