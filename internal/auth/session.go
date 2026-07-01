package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// sessionTokenBytes is the entropy of an opaque session token (256 bits).
const sessionTokenBytes = 32

// NewSessionToken returns a fresh opaque session token (given to the client)
// and its storage id, which is sha256(token) in hex. The database only ever
// stores the id, so a database compromise does not yield usable tokens.
func NewSessionToken() (token string, id string, err error) {
	buf := make([]byte, sessionTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("auth: read session entropy: %w", err)
	}
	token = base64.RawURLEncoding.EncodeToString(buf)
	return token, HashSessionToken(token), nil
}

// HashSessionToken maps a token to its storage id.
func HashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
