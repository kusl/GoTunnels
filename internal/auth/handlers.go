package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/kusl/GoTunnels/internal/activity"
	"github.com/kusl/GoTunnels/internal/httpx"
	"github.com/kusl/GoTunnels/internal/store"
)

const (
	minPasswordLen     = 8
	maxPasswordBytes   = 1024
	maxDisplayNameRune = 64
)

// Settings are the auth-related knobs handlers need, sourced from config.
type Settings struct {
	SessionTTL   time.Duration
	FlowTTL      time.Duration
	TOTPKey      [32]byte
	Issuer       string
	CookieName   string
	CookieSecure bool
}

// Handlers bundles dependencies for the auth HTTP endpoints.
type Handlers struct {
	store *store.Store
	rp    *RPProvider
	rec   *activity.Recorder
	log   *slog.Logger
	set   Settings
}

// NewHandlers builds the auth handler set. The RPProvider (rather than a
// single *webauthn.WebAuthn) lets passkey ceremonies bind to the origin the
// browser actually presents instead of whatever origin was configured when
// the process booted.
func NewHandlers(s *store.Store, rp *RPProvider, rec *activity.Recorder, log *slog.Logger, set Settings) *Handlers {
	if set.FlowTTL <= 0 {
		set.FlowTTL = 10 * time.Minute
	}
	if set.CookieName == "" {
		set.CookieName = "gotunnels_session"
	}
	return &Handlers{store: s, rp: rp, rec: rec, log: log, set: set}
}

// ---------------------------------------------------------------------------
// context / current user
// ---------------------------------------------------------------------------

type ctxKey int

const (
	userKey ctxKey = iota
	sessionKey
)

// CurrentUser returns the authenticated user, or false if unauthenticated.
func CurrentUser(ctx context.Context) (store.User, bool) {
	u, ok := ctx.Value(userKey).(store.User)
	return u, ok
}

// CurrentSession returns the active session id, or "".
func CurrentSession(ctx context.Context) string {
	if s, ok := ctx.Value(sessionKey).(string); ok {
		return s
	}
	return ""
}

// RequireAuth is middleware that loads the session referenced by the Bearer
// token (preferred) or session cookie (fallback). A missing, revoked, or
// expired session is a definitive 401. Any other failure (database down,
// timeout) is a 503: the frontend treats 401 as "logged out" and clears the
// token, so a transient fault must never masquerade as one.
func (h *Handlers) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := h.extractToken(r)
		if token == "" {
			httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		id := HashSessionToken(token)
		sess, err := h.store.GetSession(r.Context(), id)
		if err != nil {
			h.authLookupFailed(w, r, "invalid or expired session", err)
			return
		}
		user, err := h.store.GetUserByID(r.Context(), sess.UserID)
		if err != nil {
			h.authLookupFailed(w, r, "session user not found", err)
			return
		}
		if err := h.store.TouchSession(r.Context(), id); err != nil {
			h.log.WarnContext(r.Context(), "touch session", slog.String("error", err.Error()))
		}
		ctx := context.WithValue(r.Context(), userKey, user)
		ctx = context.WithValue(ctx, sessionKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handlers) authLookupFailed(w http.ResponseWriter, r *http.Request, msg string, err error) {
	if errors.Is(err, store.ErrNotFound) {
		httpx.WriteError(w, http.StatusUnauthorized, msg)
		return
	}
	h.log.ErrorContext(r.Context(), "auth: session lookup failed",
		slog.String("error", err.Error()),
		slog.String("request_id", httpx.RequestIDFromContext(r.Context())),
	)
	httpx.WriteError(w, http.StatusServiceUnavailable, "session check temporarily unavailable")
}

func (h *Handlers) extractToken(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); auth != "" {
		if v, ok := strings.CutPrefix(auth, "Bearer "); ok {
			return strings.TrimSpace(v)
		}
	}
	if c, err := r.Cookie(h.set.CookieName); err == nil {
		return c.Value
	}
	return ""
}

// ---------------------------------------------------------------------------
// request/response payloads
// ---------------------------------------------------------------------------

type signupRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	TOTP     string `json:"totp"`
}

type userResponse struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Roles       []string `json:"roles"`
	TOTPEnabled bool     `json:"totp_enabled"`
	Passkeys    int      `json:"passkeys"`
}

type sessionResponse struct {
	Token     string       `json:"token"`
	ExpiresAt *time.Time   `json:"expires_at,omitempty"`
	User      userResponse `json:"user"`
}

// ---------------------------------------------------------------------------
// signup / login / logout / me / activity
// ---------------------------------------------------------------------------

// Signup creates a password account and immediately establishes a session.
func (h *Handlers) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	username, displayName, problem := normalizeIdentity(req.Username, req.DisplayName)
	if problem != "" {
		httpx.WriteError(w, http.StatusBadRequest, problem)
		return
	}
	if problem := validatePassword(req.Password); problem != "" {
		httpx.WriteError(w, http.StatusBadRequest, problem)
		return
	}

	exists, err := h.store.UsernameExists(r.Context(), username)
	if err != nil {
		h.serverError(w, r, "signup: check username", err)
		return
	}
	if exists {
		httpx.WriteError(w, http.StatusConflict, "username already taken")
		return
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		h.serverError(w, r, "signup: hash password", err)
		return
	}
	user, err := h.store.CreateUserWithPassword(r.Context(), username, displayName, hash)
	if err != nil {
		if store.IsUniqueViolation(err) {
			httpx.WriteError(w, http.StatusConflict, "username already taken")
			return
		}
		h.serverError(w, r, "signup: create user", err)
		return
	}

	h.record(r, &user.ID, user.Username, "signup", "password", "success", nil)
	h.issueSession(w, r, user, "password")
}

// Login authenticates with a password and, when enabled, a second factor.
// Unknown users still pay for one argon2 derivation so response timing does
// not reveal which usernames exist.
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if len(req.Password) > maxPasswordBytes {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	user, err := h.store.GetUserByUsername(r.Context(), strings.TrimSpace(req.Username))
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		h.serverError(w, r, "login: load user", err)
		return
	}
	if errors.Is(err, store.ErrNotFound) {
		burnPasswordCheck(req.Password)
		h.record(r, nil, req.Username, "login", "password", "failure", map[string]any{"reason": "unknown_user"})
		httpx.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	hash, err := h.store.GetPasswordHash(r.Context(), user.ID)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		h.serverError(w, r, "login: load password", err)
		return
	}
	if errors.Is(err, store.ErrNotFound) {
		burnPasswordCheck(req.Password)
		h.record(r, &user.ID, user.Username, "login", "password", "failure", map[string]any{"reason": "no_password"})
		httpx.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	ok, err := VerifyPassword(hash, req.Password)
	if err != nil || !ok {
		h.record(r, &user.ID, user.Username, "login", "password", "failure", map[string]any{"reason": "bad_password"})
		httpx.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	enabled, err := h.totpEnabled(r.Context(), user.ID)
	if err != nil {
		h.serverError(w, r, "login: check totp", err)
		return
	}
	if !enabled {
		h.record(r, &user.ID, user.Username, "login", "password", "success", nil)
		h.issueSession(w, r, user, "password")
		return
	}
	passed, err := h.verifySecondFactor(r.Context(), user.ID, req.TOTP)
	if err != nil {
		h.serverError(w, r, "login: verify second factor", err)
		return
	}
	if !passed {
		h.record(r, &user.ID, user.Username, "login", "password+totp", "failure", map[string]any{"reason": "totp"})
		httpx.WriteError(w, http.StatusUnauthorized, "second factor required or invalid")
		return
	}
	h.record(r, &user.ID, user.Username, "login", "password+totp", "success", nil)
	h.issueSession(w, r, user, "password+totp")
}

// Logout revokes the current session.
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	if id := CurrentSession(r.Context()); id != "" {
		if err := h.store.RevokeSession(r.Context(), id); err != nil {
			h.serverError(w, r, "logout: revoke session", err)
			return
		}
	}
	if user, ok := CurrentUser(r.Context()); ok {
		h.record(r, &user.ID, user.Username, "logout", "", "success", nil)
	}
	http.SetCookie(w, h.expiredCookie())
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

// LogoutAll revokes every active session for the current user across all
// devices — the user-initiated "log out everywhere" action. The current
// session is part of that set.
func (h *Handlers) LogoutAll(w http.ResponseWriter, r *http.Request) {
	user, ok := CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	n, err := h.store.RevokeAllSessionsForUser(r.Context(), user.ID)
	if err != nil {
		h.serverError(w, r, "logout_all: revoke sessions", err)
		return
	}
	h.record(r, &user.ID, user.Username, "logout_all", "", "success", map[string]any{"sessions_revoked": n})
	http.SetCookie(w, h.expiredCookie())
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":           "logged_out_everywhere",
		"sessions_revoked": n,
	})
}

// Me returns the current user's profile.
func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	resp, err := h.userResponse(r.Context(), user)
	if err != nil {
		h.serverError(w, r, "me: load profile", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// Activity lists the current user's audit events.
func (h *Handlers) Activity(w http.ResponseWriter, r *http.Request) {
	user, ok := CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	events, err := h.store.ListActivityForUser(r.Context(), user.ID, 200)
	if err != nil {
		h.serverError(w, r, "activity: list", err)
		return
	}
	if events == nil {
		events = []store.Activity{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"activity": events})
}

// ---------------------------------------------------------------------------
// passkey registration (authenticated)
// ---------------------------------------------------------------------------

// PasskeyRegisterBegin starts adding a passkey to the current user.
func (h *Handlers) PasskeyRegisterBegin(w http.ResponseWriter, r *http.Request) {
	user, ok := CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	wa, origin, err := h.rp.ForRequest(r)
	if err != nil {
		h.record(r, &user.ID, user.Username, "passkey_register", "passkey", "failure", map[string]any{"error": err.Error()})
		httpx.WriteError(w, http.StatusBadRequest, "this origin is not allowed to use passkeys")
		return
	}
	waUser, err := h.buildWAUser(r.Context(), user)
	if err != nil {
		h.serverError(w, r, "passkey: load user", err)
		return
	}
	creation, sessionData, err := wa.BeginRegistration(waUser,
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementPreferred),
		webauthn.WithExclusions(webauthn.Credentials(waUser.Credentials).CredentialDescriptors()))
	if err != nil {
		h.serverError(w, r, "passkey: begin registration", err)
		return
	}
	flowID, err := h.saveFlow(r.Context(), &user.ID, flowRegister, origin, nil, sessionData)
	if err != nil {
		h.serverError(w, r, "passkey: save flow", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"flow_id": flowID, "options": creation})
}

// PasskeyRegisterFinish completes adding a passkey.
func (h *Handlers) PasskeyRegisterFinish(w http.ResponseWriter, r *http.Request) {
	user, ok := CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	flow, ok := h.consumeFlow(w, r, flowRegister)
	if !ok {
		return
	}
	if flow.UserID != user.ID {
		httpx.WriteError(w, http.StatusBadRequest, "invalid or expired flow")
		return
	}
	waUser, err := h.buildWAUser(r.Context(), user)
	if err != nil {
		h.serverError(w, r, "passkey: load user", err)
		return
	}
	cred, err := flow.wa.FinishRegistration(waUser, *flow.session, r)
	if err != nil {
		h.record(r, &user.ID, user.Username, "passkey_register", "passkey", "failure", map[string]any{"error": err.Error()})
		httpx.WriteError(w, http.StatusBadRequest, "passkey registration failed")
		return
	}
	if err := h.store.AddWebAuthnCredential(r.Context(), user.ID, cred); err != nil {
		if store.IsUniqueViolation(err) {
			httpx.WriteError(w, http.StatusConflict, "this passkey is already registered")
			return
		}
		h.serverError(w, r, "passkey: store credential", err)
		return
	}
	h.record(r, &user.ID, user.Username, "passkey_register", "passkey", "success", nil)
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "registered"})
}

// ---------------------------------------------------------------------------
// passkey login (unauthenticated)
// ---------------------------------------------------------------------------

type passkeyLoginBeginRequest struct {
	Username string `json:"username"`
}

// PasskeyLoginBegin starts a passkey assertion for a named user.
func (h *Handlers) PasskeyLoginBegin(w http.ResponseWriter, r *http.Request) {
	var req passkeyLoginBeginRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	wa, origin, err := h.rp.ForRequest(r)
	if err != nil {
		h.record(r, nil, req.Username, "login", "passkey", "failure", map[string]any{"error": err.Error()})
		httpx.WriteError(w, http.StatusBadRequest, "this origin is not allowed to use passkeys")
		return
	}
	user, err := h.store.GetUserByUsername(r.Context(), strings.TrimSpace(req.Username))
	if errors.Is(err, store.ErrNotFound) {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		h.serverError(w, r, "passkey: load user", err)
		return
	}
	waUser, err := h.buildWAUser(r.Context(), user)
	if err != nil {
		h.serverError(w, r, "passkey: load credentials", err)
		return
	}
	if len(waUser.Credentials) == 0 {
		httpx.WriteError(w, http.StatusBadRequest, "no passkeys registered for this account")
		return
	}
	assertion, sessionData, err := wa.BeginLogin(waUser)
	if err != nil {
		h.serverError(w, r, "passkey: begin login", err)
		return
	}
	flowID, err := h.saveFlow(r.Context(), &user.ID, flowLogin, origin, nil, sessionData)
	if err != nil {
		h.serverError(w, r, "passkey: save flow", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"flow_id": flowID, "options": assertion})
}

// PasskeyLoginFinish completes a passkey assertion and starts a session.
func (h *Handlers) PasskeyLoginFinish(w http.ResponseWriter, r *http.Request) {
	flow, ok := h.consumeFlow(w, r, flowLogin)
	if !ok {
		return
	}
	if flow.UserID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "invalid or expired flow")
		return
	}
	user, err := h.store.GetUserByID(r.Context(), flow.UserID)
	if errors.Is(err, store.ErrNotFound) {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		h.serverError(w, r, "passkey: load user", err)
		return
	}
	waUser, err := h.buildWAUser(r.Context(), user)
	if err != nil {
		h.serverError(w, r, "passkey: load credentials", err)
		return
	}
	cred, err := flow.wa.FinishLogin(waUser, *flow.session, r)
	if err != nil {
		h.record(r, &user.ID, user.Username, "login", "passkey", "failure", map[string]any{"error": err.Error()})
		httpx.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err := h.store.UpdateWebAuthnCredential(r.Context(), user.ID, cred); err != nil {
		h.log.WarnContext(r.Context(), "update credential sign count", slog.String("error", err.Error()))
	}
	h.record(r, &user.ID, user.Username, "login", "passkey", "success", nil)
	h.issueSession(w, r, user, "passkey")
}

// ---------------------------------------------------------------------------
// passkey-first signup (unauthenticated)
// ---------------------------------------------------------------------------

type passkeySignupBeginRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

// PasskeySignupBegin starts creating an account whose only credential will be
// a passkey. No user row is written yet: a candidate id travels inside the
// flow envelope and the account is only created in PasskeySignupFinish, once
// the authenticator has produced a credential.
func (h *Handlers) PasskeySignupBegin(w http.ResponseWriter, r *http.Request) {
	var req passkeySignupBeginRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	username, displayName, problem := normalizeIdentity(req.Username, req.DisplayName)
	if problem != "" {
		httpx.WriteError(w, http.StatusBadRequest, problem)
		return
	}
	exists, err := h.store.UsernameExists(r.Context(), username)
	if err != nil {
		h.serverError(w, r, "passkey signup: check username", err)
		return
	}
	if exists {
		httpx.WriteError(w, http.StatusConflict, "username already taken")
		return
	}

	wa, origin, err := h.rp.ForRequest(r)
	if err != nil {
		h.record(r, nil, username, "signup", "passkey", "failure", map[string]any{"error": err.Error()})
		httpx.WriteError(w, http.StatusBadRequest, "this origin is not allowed to use passkeys")
		return
	}

	candidateID, err := newUUIDv4()
	if err != nil {
		h.serverError(w, r, "passkey signup: candidate id", err)
		return
	}
	waUser := &WebAuthnUser{User: store.User{ID: candidateID, Username: username, DisplayName: displayName}}
	creation, sessionData, err := wa.BeginRegistration(waUser,
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementPreferred))
	if err != nil {
		h.serverError(w, r, "passkey signup: begin registration", err)
		return
	}
	flowID, err := h.saveFlow(r.Context(), nil, flowSignup, origin, &signupFlowData{
		UserID:      candidateID,
		Username:    username,
		DisplayName: displayName,
	}, sessionData)
	if err != nil {
		h.serverError(w, r, "passkey signup: save flow", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"flow_id": flowID, "options": creation})
}

// PasskeySignupFinish verifies the new credential and only then creates the
// passwordless account (user, role, and credential atomically) and starts a
// session.
func (h *Handlers) PasskeySignupFinish(w http.ResponseWriter, r *http.Request) {
	flow, ok := h.consumeFlow(w, r, flowSignup)
	if !ok {
		return
	}
	if flow.signup == nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid or expired flow")
		return
	}
	candidate := store.User{
		ID:          flow.signup.UserID,
		Username:    flow.signup.Username,
		DisplayName: flow.signup.DisplayName,
	}
	cred, err := flow.wa.FinishRegistration(&WebAuthnUser{User: candidate}, *flow.session, r)
	if err != nil {
		h.record(r, nil, candidate.Username, "signup", "passkey", "failure", map[string]any{"error": err.Error()})
		httpx.WriteError(w, http.StatusBadRequest, "passkey signup failed")
		return
	}
	user, err := h.store.CreateUserWithPasskey(r.Context(), candidate.ID, candidate.Username, candidate.DisplayName, cred)
	if err != nil {
		if store.IsUniqueViolation(err) {
			h.record(r, nil, candidate.Username, "signup", "passkey", "failure", map[string]any{"reason": "conflict"})
			httpx.WriteError(w, http.StatusConflict, "username already taken")
			return
		}
		h.serverError(w, r, "passkey signup: create user", err)
		return
	}
	h.record(r, &user.ID, user.Username, "signup", "passkey", "success", nil)
	h.issueSession(w, r, user, "passkey")
}

// ---------------------------------------------------------------------------
// TOTP (authenticated)
// ---------------------------------------------------------------------------

type totpCodeRequest struct {
	Code string `json:"code"`
}

// TOTPEnroll generates a new TOTP secret and recovery codes. The secret is
// stored encrypted and unconfirmed until TOTPConfirm. An already-active
// second factor is never replaced here: it must be disabled with a valid code
// first, so a hijacked session cannot quietly swap or strip it.
func (h *Handlers) TOTPEnroll(w http.ResponseWriter, r *http.Request) {
	user, ok := CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	sec, err := GenerateTOTP(h.issuer(), user.Username)
	if err != nil {
		h.serverError(w, r, "totp: generate", err)
		return
	}
	sealed, err := EncryptSecret(h.set.TOTPKey, []byte(sec.Secret))
	if err != nil {
		h.serverError(w, r, "totp: encrypt", err)
		return
	}
	codes, hashes, err := GenerateRecoveryCodes(10)
	if err != nil {
		h.serverError(w, r, "totp: recovery codes", err)
		return
	}
	started, err := h.store.StartTOTPEnrollment(r.Context(), user.ID, sealed, hashes)
	if err != nil {
		h.serverError(w, r, "totp: store enrollment", err)
		return
	}
	if !started {
		httpx.WriteError(w, http.StatusConflict, "two-factor authentication is already enabled; disable it first")
		return
	}
	h.record(r, &user.ID, user.Username, "totp_enroll", "totp", "success", nil)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"secret":         sec.Secret,
		"otpauth_url":    sec.URL,
		"recovery_codes": codes,
	})
}

// TOTPConfirm validates the first code and marks TOTP active.
func (h *Handlers) TOTPConfirm(w http.ResponseWriter, r *http.Request) {
	user, ok := CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var req totpCodeRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	sealed, confirmed, err := h.store.GetTOTPSecret(r.Context(), user.ID)
	if errors.Is(err, store.ErrNotFound) {
		httpx.WriteError(w, http.StatusBadRequest, "no pending TOTP enrollment")
		return
	}
	if err != nil {
		h.serverError(w, r, "totp: load secret", err)
		return
	}
	if confirmed {
		httpx.WriteError(w, http.StatusConflict, "two-factor authentication is already enabled")
		return
	}
	secret, err := DecryptSecret(h.set.TOTPKey, sealed)
	if err != nil {
		h.serverError(w, r, "totp: decrypt secret", err)
		return
	}
	step, ok := MatchTOTPStep(req.Code, string(secret), time.Now())
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, "invalid code")
		return
	}
	done, err := h.store.ConfirmTOTP(r.Context(), user.ID, step)
	if err != nil {
		h.serverError(w, r, "totp: confirm", err)
		return
	}
	if !done {
		httpx.WriteError(w, http.StatusConflict, "no pending TOTP enrollment")
		return
	}
	h.record(r, &user.ID, user.Username, "totp_confirm", "totp", "success", nil)
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "totp_enabled"})
}

// TOTPDisable turns off TOTP after verifying a current code or recovery code.
func (h *Handlers) TOTPDisable(w http.ResponseWriter, r *http.Request) {
	user, ok := CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var req totpCodeRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	enabled, err := h.totpEnabled(r.Context(), user.ID)
	if err != nil {
		h.serverError(w, r, "totp: check enabled", err)
		return
	}
	if !enabled {
		httpx.WriteError(w, http.StatusBadRequest, "two-factor authentication is not enabled")
		return
	}
	passed, err := h.verifySecondFactor(r.Context(), user.ID, req.Code)
	if err != nil {
		h.serverError(w, r, "totp: verify", err)
		return
	}
	if !passed {
		h.record(r, &user.ID, user.Username, "totp_disable", "totp", "failure", nil)
		httpx.WriteError(w, http.StatusBadRequest, "invalid code")
		return
	}
	if err := h.store.DeleteTOTP(r.Context(), user.ID); err != nil {
		h.serverError(w, r, "totp: disable", err)
		return
	}
	h.record(r, &user.ID, user.Username, "totp_disable", "totp", "success", nil)
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "totp_disabled"})
}

// ---------------------------------------------------------------------------
// flows — what goes into webauthn_flows.session_data
// ---------------------------------------------------------------------------

const (
	flowRegister = "register"
	flowLogin    = "login"
	flowSignup   = "signup"
)

// flowEnvelope wraps the go-webauthn ceremony state with the origin the flow
// was created for (so begin and finish agree on the relying party) and, for
// passkey-first signup, the candidate account created only on success.
type flowEnvelope struct {
	Origin  string          `json:"origin,omitempty"`
	Signup  *signupFlowData `json:"signup,omitempty"`
	Session json.RawMessage `json:"session"`
}

type signupFlowData struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

// liveFlow is a consumed flow, decoded and bound to its relying party.
type liveFlow struct {
	UserID  string
	signup  *signupFlowData
	session *webauthn.SessionData
	wa      *webauthn.WebAuthn
}

func decodeFlowEnvelope(blob []byte) (flowEnvelope, *webauthn.SessionData, error) {
	var env flowEnvelope
	if err := json.Unmarshal(blob, &env); err != nil {
		return flowEnvelope{}, nil, err
	}
	if len(env.Session) == 0 {
		return flowEnvelope{}, nil, errors.New("auth: flow has no session data")
	}
	var sd webauthn.SessionData
	if err := json.Unmarshal(env.Session, &sd); err != nil {
		return flowEnvelope{}, nil, err
	}
	return env, &sd, nil
}

func (h *Handlers) saveFlow(ctx context.Context, userID *string, kind, origin string, signup *signupFlowData, sd *webauthn.SessionData) (string, error) {
	sess, err := json.Marshal(sd)
	if err != nil {
		return "", err
	}
	blob, err := json.Marshal(flowEnvelope{Origin: origin, Signup: signup, Session: sess})
	if err != nil {
		return "", err
	}
	id, err := newFlowID()
	if err != nil {
		return "", err
	}
	err = h.store.SaveFlow(ctx, store.Flow{
		ID:          id,
		UserID:      userID,
		Kind:        kind,
		SessionData: blob,
		ExpiresAt:   time.Now().Add(h.set.FlowTTL),
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

// consumeFlow atomically takes the flow named by ?flow= out of the store, so
// each ceremony can be finished at most once, then decodes it and resolves
// the relying party pinned when it began. On failure it has already written
// the response.
func (h *Handlers) consumeFlow(w http.ResponseWriter, r *http.Request, kind string) (liveFlow, bool) {
	id := r.URL.Query().Get("flow")
	if id == "" {
		httpx.WriteError(w, http.StatusBadRequest, "invalid or expired flow")
		return liveFlow{}, false
	}
	flow, err := h.store.ConsumeFlow(r.Context(), id, kind)
	if errors.Is(err, store.ErrNotFound) {
		httpx.WriteError(w, http.StatusBadRequest, "invalid or expired flow")
		return liveFlow{}, false
	}
	if err != nil {
		h.serverError(w, r, "flow: consume", err)
		return liveFlow{}, false
	}
	env, sd, err := decodeFlowEnvelope(flow.SessionData)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid or expired flow")
		return liveFlow{}, false
	}
	wa, _, err := h.rp.ForOrigin(env.Origin)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "this origin is not allowed to use passkeys")
		return liveFlow{}, false
	}
	lf := liveFlow{signup: env.Signup, session: sd, wa: wa}
	if flow.UserID != nil {
		lf.UserID = *flow.UserID
	}
	return lf, true
}

// ---------------------------------------------------------------------------
// sessions & cookies
// ---------------------------------------------------------------------------

func (h *Handlers) issueSession(w http.ResponseWriter, r *http.Request, user store.User, method string) {
	token, id, err := NewSessionToken()
	if err != nil {
		h.serverError(w, r, "session: token", err)
		return
	}
	dbExpiry, cookieExpiry, cookieMaxAge := computeSessionExpiry(time.Now(), h.set.SessionTTL)
	if err := h.store.CreateSession(r.Context(), id, user.ID, method, dbExpiry); err != nil {
		h.serverError(w, r, "session: create", err)
		return
	}
	profile, err := h.userResponse(r.Context(), user)
	if err != nil {
		h.log.WarnContext(r.Context(), "session: load profile", slog.String("error", err.Error()))
		profile = userResponse{ID: user.ID, Username: user.Username, DisplayName: user.DisplayName, Roles: []string{}}
	}
	http.SetCookie(w, h.sessionCookie(token, cookieExpiry, cookieMaxAge))
	httpx.WriteJSON(w, http.StatusOK, sessionResponse{Token: token, ExpiresAt: dbExpiry, User: profile})
}

// persistentSessionCookieTTL dates the secondary session cookie when the
// session itself never expires. Browsers clamp it further (~400 days); it is
// re-set on every login.
const persistentSessionCookieTTL = 365 * 24 * time.Hour

// computeSessionExpiry maps the configured session TTL to the database expiry
// (nil = never expires) and the cookie's Expires/Max-Age. A TTL of zero or
// less means only an explicit logout ends the session.
func computeSessionExpiry(now time.Time, ttl time.Duration) (dbExpiry *time.Time, cookieExpiry time.Time, cookieMaxAge int) {
	if ttl > 0 {
		exp := now.Add(ttl)
		return &exp, exp, int(ttl.Seconds())
	}
	return nil, now.Add(persistentSessionCookieTTL), int(persistentSessionCookieTTL.Seconds())
}

func (h *Handlers) sessionCookie(token string, expires time.Time, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     h.set.CookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.set.CookieSecure,
		SameSite: http.SameSiteNoneMode,
	}
}

func (h *Handlers) expiredCookie() *http.Cookie {
	return &http.Cookie{
		Name:     h.set.CookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.set.CookieSecure,
		SameSite: http.SameSiteNoneMode,
	}
}

// ---------------------------------------------------------------------------
// internal helpers
// ---------------------------------------------------------------------------

func (h *Handlers) userResponse(ctx context.Context, u store.User) (userResponse, error) {
	roles, err := h.store.UserRoles(ctx, u.ID)
	if err != nil {
		return userResponse{}, err
	}
	if roles == nil {
		roles = []string{}
	}
	totp, err := h.totpEnabled(ctx, u.ID)
	if err != nil {
		return userResponse{}, err
	}
	pk, err := h.store.CountWebAuthnCredentials(ctx, u.ID)
	if err != nil {
		return userResponse{}, err
	}
	return userResponse{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		Roles:       roles,
		TOTPEnabled: totp,
		Passkeys:    pk,
	}, nil
}

func (h *Handlers) buildWAUser(ctx context.Context, u store.User) (*WebAuthnUser, error) {
	creds, err := h.store.GetWebAuthnCredentials(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	return &WebAuthnUser{User: u, Credentials: creds}, nil
}

func (h *Handlers) totpEnabled(ctx context.Context, userID string) (bool, error) {
	_, confirmed, err := h.store.GetTOTPSecret(ctx, userID)
	if errors.Is(err, store.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return confirmed, nil
}

// verifySecondFactor accepts a TOTP code whose time step is newer than the
// last one accepted (so every code is single-use) or an unused recovery code
// (consumed on success). Storage failures are returned as errors, never as a
// silent pass or fail.
func (h *Handlers) verifySecondFactor(ctx context.Context, userID, code string) (bool, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return false, nil
	}
	sealed, confirmed, err := h.store.GetTOTPSecret(ctx, userID)
	if errors.Is(err, store.ErrNotFound) || (err == nil && !confirmed) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	secret, err := DecryptSecret(h.set.TOTPKey, sealed)
	if err != nil {
		return false, err
	}
	if step, ok := MatchTOTPStep(code, string(secret), time.Now()); ok {
		return h.store.MarkTOTPStepUsed(ctx, userID, step)
	}
	return h.store.UseRecoveryCode(ctx, userID, HashRecoveryCode(code))
}

func (h *Handlers) record(r *http.Request, userID *string, username, eventType, method, outcome string, detail map[string]any) {
	err := h.rec.Record(r.Context(), r, activity.Event{
		UserID:     userID,
		Username:   username,
		EventType:  eventType,
		AuthMethod: method,
		Outcome:    outcome,
		Detail:     detail,
	})
	if err != nil {
		h.log.WarnContext(r.Context(), "record activity", slog.String("error", err.Error()), slog.String("event", eventType))
	}
}

func (h *Handlers) serverError(w http.ResponseWriter, r *http.Request, msg string, err error) {
	httpx.ServerError(w, r, h.log, msg, err)
}

func (h *Handlers) issuer() string {
	if h.set.Issuer != "" {
		return h.set.Issuer
	}
	return "GoTunnels"
}

// ---------------------------------------------------------------------------
// free helpers
// ---------------------------------------------------------------------------

var (
	dummyHashOnce sync.Once
	dummyHash     string
)

// burnPasswordCheck performs one argon2 verification against a throwaway hash
// so failed logins for unknown or passwordless accounts take as long as a
// wrong password for a real one.
func burnPasswordCheck(password string) {
	dummyHashOnce.Do(func() {
		dummyHash, _ = HashPassword("gotunnels-timing-equaliser")
	})
	if dummyHash != "" {
		_, _ = VerifyPassword(dummyHash, password)
	}
}

func newFlowID() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// newUUIDv4 generates an RFC 9562 version 4 UUID string for the passkey
// signup candidate, which must exist as a WebAuthn user handle before any row
// exists in the database.
func newUUIDv4() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	return httpx.DecodeJSON(w, r, dst, 64<<10)
}

// normalizeIdentity trims and validates a username and display name, falling
// back to the username when the display name is blank.
func normalizeIdentity(username, displayName string) (string, string, string) {
	username = strings.TrimSpace(username)
	if !validUsername(username) {
		return "", "", "username must be 3-32 chars: letters, digits, dot, dash, underscore"
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		displayName = username
	}
	if !utf8.ValidString(displayName) || utf8.RuneCountInString(displayName) > maxDisplayNameRune {
		return "", "", fmt.Sprintf("display name must be at most %d characters", maxDisplayNameRune)
	}
	for _, r := range displayName {
		if r < 0x20 || (r >= 0x7f && r < 0xa0) {
			return "", "", "display name must not contain control characters"
		}
	}
	return username, displayName, ""
}

func validatePassword(p string) string {
	switch {
	case len(p) < minPasswordLen:
		return fmt.Sprintf("password must be at least %d characters", minPasswordLen)
	case len(p) > maxPasswordBytes:
		return fmt.Sprintf("password must be at most %d bytes", maxPasswordBytes)
	}
	return ""
}

// validUsername enforces a conservative username policy.
func validUsername(u string) bool {
	if len(u) < 3 || len(u) > 32 {
		return false
	}
	for _, r := range u {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '.' || r == '-' || r == '_':
		default:
			return false
		}
	}
	return true
}
