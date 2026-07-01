package auth

import (
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/kusl/GoTunnels/internal/store"
)

// NewWebAuthn builds a WebAuthn relying party from configuration. RPID is the
// frontend's registrable domain and RPOrigins are the full origins the browser
// will present (both are discovered at startup for Cloudflare Quick Tunnels).
func NewWebAuthn(rpID, rpDisplayName string, rpOrigins []string) (*webauthn.WebAuthn, error) {
	return webauthn.New(&webauthn.Config{
		RPID:          rpID,
		RPDisplayName: rpDisplayName,
		RPOrigins:     rpOrigins,
	})
}

// WebAuthnUser adapts an application user and its registered credentials to the
// webauthn.User interface the go-webauthn library expects.
type WebAuthnUser struct {
	User        store.User
	Credentials []webauthn.Credential
}

// WebAuthnID returns a stable, opaque user handle (the UUID string; under the
// 64-byte limit the spec requires).
func (u *WebAuthnUser) WebAuthnID() []byte { return []byte(u.User.ID) }

// WebAuthnName returns the username.
func (u *WebAuthnUser) WebAuthnName() string { return u.User.Username }

// WebAuthnDisplayName returns the human-friendly display name.
func (u *WebAuthnUser) WebAuthnDisplayName() string {
	if u.User.DisplayName != "" {
		return u.User.DisplayName
	}
	return u.User.Username
}

// WebAuthnCredentials returns the user's registered credentials.
func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential { return u.Credentials }
