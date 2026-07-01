package auth

import (
	"strings"
	"testing"
)

func TestHashPasswordRoundTrip(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$") {
		t.Fatalf("unexpected PHC prefix: %q", hash)
	}

	ok, err := VerifyPassword(hash, "correct horse battery staple")
	if err != nil {
		t.Fatalf("VerifyPassword: %v", err)
	}
	if !ok {
		t.Fatal("expected password to verify")
	}
}

func TestVerifyPasswordWrong(t *testing.T) {
	hash, err := HashPassword("hunter2")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	ok, err := VerifyPassword(hash, "hunter3")
	if err != nil {
		t.Fatalf("VerifyPassword: %v", err)
	}
	if ok {
		t.Fatal("expected wrong password to fail")
	}
}

func TestHashPasswordUniqueSalts(t *testing.T) {
	a, _ := HashPassword("same")
	b, _ := HashPassword("same")
	if a == b {
		t.Fatal("expected different hashes due to random salt")
	}
}

func TestHashPasswordEmpty(t *testing.T) {
	if _, err := HashPassword(""); err == nil {
		t.Fatal("expected error for empty password")
	}
}

func TestVerifyPasswordBadFormats(t *testing.T) {
	cases := []string{
		"",
		"not-a-hash",
		"$argon2i$v=19$m=65536,t=1,p=4$c2FsdA$aGFzaA",  // wrong variant
		"$argon2id$v=19$m=65536,t=1$c2FsdA$aGFzaA",     // missing p
		"$argon2id$v=99$m=65536,t=1,p=4$c2FsdA$aGFzaA", // bad version
	}
	for _, c := range cases {
		if _, err := VerifyPassword(c, "x"); err == nil {
			t.Fatalf("expected error for %q", c)
		}
	}
}

func TestDecodePHCFields(t *testing.T) {
	hash, err := HashPassword("abc")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	p, salt, digest, err := decodePHC(hash)
	if err != nil {
		t.Fatalf("decodePHC: %v", err)
	}
	if p.memory != argonMemory || p.time != argonTime || p.threads != argonThreads {
		t.Fatalf("params mismatch: %+v", p)
	}
	if len(salt) != argonSaltLen {
		t.Fatalf("salt len = %d, want %d", len(salt), argonSaltLen)
	}
	if len(digest) != argonKeyLen {
		t.Fatalf("digest len = %d, want %d", len(digest), argonKeyLen)
	}
}
