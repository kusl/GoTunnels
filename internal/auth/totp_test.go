package auth

import (
	"bytes"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

func TestGenerateAndValidateTOTP(t *testing.T) {
	sec, err := GenerateTOTP("GoTunnels", "alice")
	if err != nil {
		t.Fatalf("GenerateTOTP: %v", err)
	}
	if sec.Secret == "" || sec.URL == "" {
		t.Fatal("expected non-empty secret and URL")
	}

	code, err := totp.GenerateCode(sec.Secret, time.Now())
	if err != nil {
		t.Fatalf("GenerateCode: %v", err)
	}
	if !ValidateTOTP(code, sec.Secret) {
		t.Fatal("expected freshly generated code to validate")
	}
	if ValidateTOTP("000000", sec.Secret) && ValidateTOTP("123456", sec.Secret) {
		t.Fatal("two arbitrary codes both validated; extremely unlikely, logic suspect")
	}
}

func TestEncryptDecryptSecretRoundTrip(t *testing.T) {
	key := sha256.Sum256([]byte("unit-test-key"))
	plaintext := []byte("JBSWY3DPEHPK3PXP")

	sealed, err := EncryptSecret(key, plaintext)
	if err != nil {
		t.Fatalf("EncryptSecret: %v", err)
	}
	if bytes.Contains(sealed, plaintext) {
		t.Fatal("ciphertext should not contain plaintext")
	}

	opened, err := DecryptSecret(key, sealed)
	if err != nil {
		t.Fatalf("DecryptSecret: %v", err)
	}
	if !bytes.Equal(opened, plaintext) {
		t.Fatalf("round trip mismatch: got %q want %q", opened, plaintext)
	}
}

func TestDecryptSecretWrongKey(t *testing.T) {
	k1 := sha256.Sum256([]byte("key-one"))
	k2 := sha256.Sum256([]byte("key-two"))
	sealed, err := EncryptSecret(k1, []byte("secret"))
	if err != nil {
		t.Fatalf("EncryptSecret: %v", err)
	}
	if _, err := DecryptSecret(k2, sealed); err == nil {
		t.Fatal("expected authentication failure with wrong key")
	}
}

func TestDecryptSecretTooShort(t *testing.T) {
	key := sha256.Sum256([]byte("k"))
	if _, err := DecryptSecret(key, []byte("x")); err == nil {
		t.Fatal("expected error for short ciphertext")
	}
}

func TestEncryptSecretUniqueNonce(t *testing.T) {
	key := sha256.Sum256([]byte("k"))
	a, _ := EncryptSecret(key, []byte("same"))
	b, _ := EncryptSecret(key, []byte("same"))
	if bytes.Equal(a, b) {
		t.Fatal("expected distinct ciphertexts due to random nonce")
	}
}

func TestNormalizeRecoveryCode(t *testing.T) {
	cases := map[string]string{
		"ABCDE-FGHIJ": "abcdefghij",
		"abc de fghi": "abcdefghi",
		"ab-cd-ef":    "abcdef",
		"XY23-4Z":     "xy234z",
	}
	for in, want := range cases {
		if got := NormalizeRecoveryCode(in); got != want {
			t.Fatalf("NormalizeRecoveryCode(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestGenerateRecoveryCodes(t *testing.T) {
	codes, hashes, err := GenerateRecoveryCodes(8)
	if err != nil {
		t.Fatalf("GenerateRecoveryCodes: %v", err)
	}
	if len(codes) != 8 || len(hashes) != 8 {
		t.Fatalf("expected 8 codes and hashes, got %d/%d", len(codes), len(hashes))
	}
	seen := map[string]bool{}
	for i, c := range codes {
		if seen[c] {
			t.Fatalf("duplicate recovery code %q", c)
		}
		seen[c] = true
		if HashRecoveryCode(c) != hashes[i] {
			t.Fatalf("hash mismatch for code %d", i)
		}
	}
}

func TestHashRecoveryCodeStableAcrossFormatting(t *testing.T) {
	if HashRecoveryCode("ABCDE-FGHIJ") != HashRecoveryCode("abcdefghij") {
		t.Fatal("expected formatting-insensitive recovery hashing")
	}
}

func TestCompareRecoveryHash(t *testing.T) {
	h := HashRecoveryCode("abcde-fghij")
	if !CompareRecoveryHash(h, h) {
		t.Fatal("expected equal hashes to compare true")
	}
	if CompareRecoveryHash(h, HashRecoveryCode("zzzzz-zzzzz")) {
		t.Fatal("expected different hashes to compare false")
	}
}

func TestGenerateRecoveryCodesDefaultCount(t *testing.T) {
	codes, _, err := GenerateRecoveryCodes(0)
	if err != nil {
		t.Fatalf("GenerateRecoveryCodes: %v", err)
	}
	if len(codes) != 10 {
		t.Fatalf("expected default of 10 codes, got %d", len(codes))
	}
}
