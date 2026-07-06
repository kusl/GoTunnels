package auth

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/kusl/GoTunnels/internal/httpx"
	"github.com/kusl/GoTunnels/internal/store"
)

// NewWebAuthn builds a WebAuthn relying party from configuration. RPID is the
// frontend's registrable domain and RPOrigins are the full origins the browser
// will present.
func NewWebAuthn(rpID, rpDisplayName string, rpOrigins []string) (*webauthn.WebAuthn, error) {
	return webauthn.New(&webauthn.Config{
		RPID:          rpID,
		RPDisplayName: rpDisplayName,
		RPOrigins:     rpOrigins,
	})
}

// ---------------------------------------------------------------------------
// Per-request relying party
// ---------------------------------------------------------------------------

// RPConfig configures an RPProvider.
type RPConfig struct {
	// RPID / RPDisplayName / RPOrigins are the statically configured relying
	// party (what scripts/up.sh discovers at startup). They remain the
	// fallback for requests that carry no Origin header.
	RPID          string
	RPDisplayName string
	RPOrigins     []string
	// OriginPatterns are wildcard origins (see httpx.MatchOriginPattern) that
	// are additionally allowed to act as the relying party, with the RP ID
	// derived from the request's Origin host. This is what keeps passkeys
	// working when a Cloudflare Quick Tunnel hostname rotates or a container
	// booted before tunnel discovery: the browser tells us its origin, and if
	// that origin matches a trusted pattern we build the RP around it.
	//
	// Security note: deriving the RP from the Origin is safe here because
	// WebAuthn credentials are scoped to the RP ID by the authenticator
	// itself. A page on attacker.trycloudflare.com can only ever exercise
	// credentials created for attacker.trycloudflare.com — it cannot assert
	// a credential registered for your-frontend.trycloudflare.com. The same
	// is NOT true of CORS-with-credentials, which is why these patterns are
	// not fed into the CORS allow-list.
	OriginPatterns []string
}

// maxDerivedRPs bounds the per-origin cache so a client spraying many
// pattern-matching Origin values cannot grow memory without limit.
const maxDerivedRPs = 256

// RPProvider hands out *webauthn.WebAuthn instances appropriate for a given
// request origin: the static instance for configured origins, or a derived,
// cached instance for origins matching a trusted wildcard pattern.
type RPProvider struct {
	displayName   string
	static        *webauthn.WebAuthn
	staticDefault string
	staticOrigins map[string]struct{}
	patterns      []string

	mu       sync.Mutex
	byOrigin map[string]*webauthn.WebAuthn
}

// NewRPProvider validates the static configuration and prepares the provider.
func NewRPProvider(cfg RPConfig) (*RPProvider, error) {
	static, err := NewWebAuthn(cfg.RPID, cfg.RPDisplayName, cfg.RPOrigins)
	if err != nil {
		return nil, err
	}
	statics := make(map[string]struct{}, len(cfg.RPOrigins))
	for _, o := range cfg.RPOrigins {
		statics[strings.ToLower(strings.TrimSpace(o))] = struct{}{}
	}
	def := ""
	if len(cfg.RPOrigins) > 0 {
		def = cfg.RPOrigins[0]
	}
	return &RPProvider{
		displayName:   cfg.RPDisplayName,
		static:        static,
		staticDefault: def,
		staticOrigins: statics,
		patterns:      cfg.OriginPatterns,
		byOrigin:      map[string]*webauthn.WebAuthn{},
	}, nil
}

// Static returns the boot-time configured relying party.
func (p *RPProvider) Static() *webauthn.WebAuthn { return p.static }

// ForRequest resolves the relying party for an incoming HTTP request from its
// Origin header. Requests without an Origin header (curl, tests) fall back to
// the static configuration.
func (p *RPProvider) ForRequest(r *http.Request) (*webauthn.WebAuthn, string, error) {
	return p.ForOrigin(r.Header.Get("Origin"))
}

// ForOrigin resolves the relying party for a browser origin. It returns the
// instance, the canonical origin it is bound to (persist this with the flow so
// begin and finish agree even if the header changes between them), and an
// error when the origin is neither configured nor pattern-trusted.
func (p *RPProvider) ForOrigin(origin string) (*webauthn.WebAuthn, string, error) {
	origin = strings.ToLower(strings.TrimSpace(origin))
	if origin == "" {
		return p.static, p.staticDefault, nil
	}
	if _, ok := p.staticOrigins[origin]; ok {
		return p.static, origin, nil
	}
	for _, pat := range p.patterns {
		if httpx.MatchOriginPattern(pat, origin) {
			wa, err := p.derived(origin)
			if err != nil {
				return nil, "", err
			}
			return wa, origin, nil
		}
	}
	return nil, "", fmt.Errorf("auth: origin %q is not an allowed WebAuthn relying party", origin)
}

func (p *RPProvider) derived(origin string) (*webauthn.WebAuthn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if wa, ok := p.byOrigin[origin]; ok {
		return wa, nil
	}
	rpID := hostOfOrigin(origin)
	if rpID == "" {
		return nil, fmt.Errorf("auth: cannot derive RP ID from origin %q", origin)
	}
	wa, err := NewWebAuthn(rpID, p.displayName, []string{origin})
	if err != nil {
		return nil, err
	}
	if len(p.byOrigin) >= maxDerivedRPs {
		p.byOrigin = map[string]*webauthn.WebAuthn{}
	}
	p.byOrigin[origin] = wa
	return wa, nil
}

// hostOfOrigin extracts the bare host (no scheme, no port) from an origin.
// The RP ID must be a domain, never host:port.
func hostOfOrigin(origin string) string {
	if i := strings.Index(origin, "://"); i >= 0 {
		origin = origin[i+3:]
	}
	if i := strings.IndexByte(origin, '/'); i >= 0 {
		origin = origin[:i]
	}
	if i := strings.LastIndexByte(origin, ':'); i >= 0 {
		origin = origin[:i]
	}
	return origin
}

// ---------------------------------------------------------------------------
// webauthn.User adapter
// ---------------------------------------------------------------------------

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
