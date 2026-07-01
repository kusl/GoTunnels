package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTPSecret bundles a freshly generated shared secret and its provisioning
// URL (which the frontend renders as a QR code / manual key).
type TOTPSecret struct {
	Secret string // base32, no padding
	URL    string // otpauth:// provisioning URI
}

// GenerateTOTP creates a new TOTP secret for the given issuer/account.
func GenerateTOTP(issuer, account string) (TOTPSecret, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: account,
		Algorithm:   otp.AlgorithmSHA1, // widest authenticator-app compatibility
		Digits:      otp.DigitsSix,
		Period:      30,
	})
	if err != nil {
		return TOTPSecret{}, fmt.Errorf("auth: generate totp: %w", err)
	}
	return TOTPSecret{Secret: key.Secret(), URL: key.URL()}, nil
}

// ValidateTOTP reports whether code is currently valid for secret.
func ValidateTOTP(code, secret string) bool {
	return totp.Validate(strings.TrimSpace(code), secret)
}

// ---------------------------------------------------------------------------
// Encryption of the secret at rest (AES-256-GCM)
// ---------------------------------------------------------------------------

// EncryptSecret seals a TOTP secret with AES-256-GCM. Output layout is
// nonce || ciphertext, so DecryptSecret is self-contained.
func EncryptSecret(key [32]byte, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	// Seal appends ciphertext to dst (nonce), returning nonce||ciphertext.
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// DecryptSecret opens a value produced by EncryptSecret.
func DecryptSecret(key [32]byte, sealed []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ns := gcm.NonceSize()
	if len(sealed) < ns {
		return nil, errors.New("auth: ciphertext too short")
	}
	nonce, ct := sealed[:ns], sealed[ns:]
	return gcm.Open(nil, nonce, ct, nil)
}

// ---------------------------------------------------------------------------
// Recovery codes
// ---------------------------------------------------------------------------

// GenerateRecoveryCodes returns n human-friendly one-time codes together with
// their hashes (what gets stored). Codes are shown to the user exactly once.
func GenerateRecoveryCodes(n int) (codes []string, hashes []string, err error) {
	if n <= 0 {
		n = 10
	}
	for i := 0; i < n; i++ {
		code, err := randomRecoveryCode()
		if err != nil {
			return nil, nil, err
		}
		codes = append(codes, code)
		hashes = append(hashes, HashRecoveryCode(code))
	}
	return codes, hashes, nil
}

// HashRecoveryCode hashes a recovery code for storage/comparison. Recovery
// codes are high-entropy, so a fast SHA-256 is acceptable here (unlike
// passwords). Normalisation strips formatting dashes and lowercases.
func HashRecoveryCode(code string) string {
	sum := sha256.Sum256([]byte(NormalizeRecoveryCode(code)))
	return hex.EncodeToString(sum[:])
}

// NormalizeRecoveryCode canonicalises user input (dashes/spaces/case).
func NormalizeRecoveryCode(code string) string {
	var b strings.Builder
	for _, r := range code {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
		default:
			// drop dashes, spaces, etc.
		}
	}
	return b.String()
}

// CompareRecoveryHash is a constant-time equality check for two hex hashes.
func CompareRecoveryHash(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

var recoveryAlphabet = []byte("abcdefghijkmnpqrstuvwxyz23456789") // no ambiguous chars

func randomRecoveryCode() (string, error) {
	const groups, perGroup = 2, 5
	buf := make([]byte, groups*perGroup)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	var b strings.Builder
	for g := 0; g < groups; g++ {
		if g > 0 {
			b.WriteByte('-')
		}
		for i := 0; i < perGroup; i++ {
			idx := int(buf[g*perGroup+i]) % len(recoveryAlphabet)
			b.WriteByte(recoveryAlphabet[idx])
		}
	}
	return b.String(), nil
}

// base32NoPad is exposed for callers that need to sanity-check a secret.
var base32NoPad = base32.StdEncoding.WithPadding(base32.NoPadding)
