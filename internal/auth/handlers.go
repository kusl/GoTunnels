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
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/kusl/GoTunnels/internal/activity"
	"github.com/kusl/GoTunnels/internal/httpx"
	"github.com/kusl/GoTunnels/internal/store"
)

// Settings are the auth-related knobs handlers need, sourced from config.
type Settings struct {
	SessionTTL   time.Duration
	FlowTTL      time.Duration
	TOTPKey      [32]byte
	Issuer       string // TOTP issuer / RP display name
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
// single *webauthn.WebAuthn) is what lets passkey ceremonies bind to the
// origin the browser actually presents instead of whatever origin happened to
// be configured when the process booted — the root cause of the classic
// "requested RPID did not match the origin" error behind a rotating tunnel.
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
// token (preferred) or session cookie (fallback) and rejects with 401 when
// absent or invalid.
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
			httpx.WriteError(w, http.StatusUnauthorized, "invalid or expired session")
			return
		}
		user, err := h.store.GetUserByID(r.Context(), sess.UserID)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "session user not found")
			return
		}
		_ = h.store.TouchSession(r.Context(), id)
		ctx := context.WithValue(r.Context(), userKey, user)
		ctx = context.WithValue(ctx, sessionKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
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
	Token string `json:"token"`
	// ExpiresAt is omitted entirely for a persistent (never-expiring) session,
	// so clients can distinguish "no expiry" from a concrete time.
	ExpiresAt *time.Time   `json:"expires_at,omitempty"`
	User      userResponse `json:"user"`
}

// ---------------------------------------------------------------------------
// signup / login / logout / me / activity
// ---------------------------------------------------------------------------

// Signup creates an account with a password and immediately establishes a
// session. (Password is one of two ways in — see PasskeySignupBegin/Finish
// for the passwordless path.)
func (h *Handlers) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if !validUsername(req.Username) {
		httpx.WriteError(w, http.StatusBadRequest, "username must be 3-32 chars: letters, digits, dot, dash, underscore")
		return
	}
	if len(req.Password) < 8 {
		httpx.WriteError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	exists, err := h.store.UsernameExists(r.Context(), req.Username)
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
	user, err := h.store.CreateUser(r.Context(), req.Username, req.DisplayName)
	if err != nil {
		if store.IsUniqueViolation(err) {
			httpx.WriteError(w, http.StatusConflict, "username already taken")
			return
		}
		h.serverError(w, r, "signup: create user", err)
		return
	}
	if err := h.store.SetPassword(r.Context(), user.ID, hash); err != nil {
		h.serverError(w, r, "signup: set password", err)
		return
	}

	h.record(r, &user.ID, user.Username, "signup", "password", "success", nil)
	h.issueSession(w, r, user, "password")
}

// Login authenticates with a password and, when enabled, a second factor.
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	user, err := h.store.GetUserByUsername(r.Context(), strings.TrimSpace(req.Username))
	if err != nil {
		// Do not distinguish unknown user from bad password.
		h.record(r, nil, req.Username, "login", "password", "failure", map[string]any{"reason": "unknown_user"})
		httpx.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	hash, err := h.store.GetPasswordHash(r.Context(), user.ID)
	if err != nil {
		// Passkey-only accounts have no password row; same generic answer.
		httpx.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	ok, err := VerifyPassword(hash, req.Password)
	if err != nil || !ok {
		h.record(r, &user.ID, user.Username, "login", "password", "failure", map[string]any{"reason": "bad_password"})
		httpx.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Second factor, if the user has a confirmed TOTP secret.
	if enabled, _ := h.totpEnabled(r.Context(), user.ID); enabled {
		if !h.verifySecondFactor(r.Context(), user.ID, req.TOTP) {
			h.record(r, &user.ID, user.Username, "login", "password+totp", "failure", map[string]any{"reason": "totp"})
			httpx.WriteError(w, http.StatusUnauthorized, "second factor required or invalid")
			return
		}
		h.record(r, &user.ID, user.Username, "login", "password+totp", "success", nil)
		h.issueSession(w, r, user, "password+totp")
		return
	}

	h.record(r, &user.ID, user.Username, "login", "password", "success", nil)
	h.issueSession(w, r, user, "password")
}

// Logout revokes the current session.
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	id := CurrentSession(r.Context())
	if id != "" {
		_ = h.store.RevokeSession(r.Context(), id)
	}
	if user, ok := CurrentUser(r.Context()); ok {
		h.record(r, &user.ID, user.Username, "logout", "", "success", nil)
	}
	http.SetCookie(w, h.expiredCookie())
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

// LogoutAll revokes every currently-active session for the current user across
// all their devices and browsers — the "log out everywhere" action exposed on
// the settings page. This is the ONLY logout GoTunnels performs beyond a
// single explicit Logout, and it too is entirely user-initiated: the app never
// ends a session on its own. The response reports how many sessions were
// revoked. The current session is included in that set, so this also logs the
// caller out here; the frontend drops its local token and the expired cookie
// clears the secondary transport.
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
	httpx.WriteJSON(w, http.StatusOK, h.userResponse(r.Context(), user))
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
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementPreferred))
	if err != nil {
		h.serverError(w, r, "passkey: begin registration", err)
		return
	}
	flowID, err := h.saveFlow(r.Context(), &user.ID, "register", origin, nil, sessionData)
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
	flowID := r.URL.Query().Get("flow")
	env, err := h.loadFlow(r.Context(), flowID, "register", &user.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid or expired registration flow")
		return
	}
	sd, err := env.sessionData()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "corrupt registration flow")
		return
	}
	// Finish against the relying party pinned when the flow began, so begin
	// and finish agree even if the Origin header were to change between them.
	wa, err := h.rpForFlow(env)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "this origin is not allowed to use passkeys")
		return
	}
	waUser, err := h.buildWAUser(r.Context(), user)
	if err != nil {
		h.serverError(w, r, "passkey: load user", err)
		return
	}
	cred, err := wa.FinishRegistration(waUser, *sd, r)
	if err != nil {
		h.record(r, &user.ID, user.Username, "passkey_register", "passkey", "failure", map[string]any{"error": err.Error()})
		httpx.WriteError(w, http.StatusBadRequest, "passkey registration failed")
		return
	}
	if err := h.store.AddWebAuthnCredential(r.Context(), user.ID, cred); err != nil {
		h.serverError(w, r, "passkey: store credential", err)
		return
	}
	_ = h.store.DeleteFlow(r.Context(), flowID)
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
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	waUser, err := h.buildWAUser(r.Context(), user)
	if err != nil {
		h.serverError(w, r, "passkey: load user", err)
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
	flowID, err := h.saveFlow(r.Context(), &user.ID, "login", origin, nil, sessionData)
	if err != nil {
		h.serverError(w, r, "passkey: save flow", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"flow_id": flowID, "options": assertion})
}

// PasskeyLoginFinish completes a passkey assertion and starts a session.
func (h *Handlers) PasskeyLoginFinish(w http.ResponseWriter, r *http.Request) {
	flowID := r.URL.Query().Get("flow")
	flow, err := h.loadFlowAny(r.Context(), flowID, "login")
	if err != nil || flow.UserID == nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid or expired login flow")
		return
	}
	user, err := h.store.GetUserByID(r.Context(), *flow.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	env := decodeFlowEnvelope(flow.SessionData)
	sd, err := env.sessionData()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "corrupt login flow")
		return
	}
	wa, err := h.rpForFlow(env)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "this origin is not allowed to use passkeys")
		return
	}
	waUser, err := h.buildWAUser(r.Context(), user)
	if err != nil {
		h.serverError(w, r, "passkey: load user", err)
		return
	}
	cred, err := wa.FinishLogin(waUser, *sd, r)
	if err != nil {
		h.record(r, &user.ID, user.Username, "login", "passkey", "failure", map[string]any{"error": err.Error()})
		httpx.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err := h.store.UpdateWebAuthnCredential(r.Context(), user.ID, cred); err != nil {
		h.log.WarnContext(r.Context(), "update credential sign count", slog.String("error", err.Error()))
	}
	_ = h.store.DeleteFlow(r.Context(), flowID)
	h.record(r, &user.ID, user.Username, "login", "passkey", "success", nil)
	h.issueSession(w, r, user, "passkey")
}

// ---------------------------------------------------------------------------
// passkey-first signup (unauthenticated) — create an account with no password
// ---------------------------------------------------------------------------

type passkeySignupBeginRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

// PasskeySignupBegin starts creating a brand-new account whose only credential
// will be a passkey. No user row is written yet: a candidate user id is
// generated here and carried inside the flow envelope, and the account only
// comes into existence in PasskeySignupFinish after the authenticator has
// actually produced a credential. Abandoned ceremonies therefore leave nothing
// behind but an expiring row in webauthn_flows.
func (h *Handlers) PasskeySignupBegin(w http.ResponseWriter, r *http.Request) {
	var req passkeySignupBeginRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if !validUsername(req.Username) {
		httpx.WriteError(w, http.StatusBadRequest, "username must be 3-32 chars: letters, digits, dot, dash, underscore")
		return
	}
	// Fast-fail on an obviously taken name. The authoritative check is the
	// unique index at Finish time, which also closes the begin/finish race.
	exists, err := h.store.UsernameExists(r.Context(), req.Username)
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
		h.record(r, nil, req.Username, "signup", "passkey", "failure", map[string]any{"error": err.Error()})
		httpx.WriteError(w, http.StatusBadRequest, "this origin is not allowed to use passkeys")
		return
	}

	candidateID, err := newUUIDv4()
	if err != nil {
		h.serverError(w, r, "passkey signup: candidate id", err)
		return
	}
	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		displayName = req.Username
	}

	// A transient user object: satisfies webauthn.User for the ceremony
	// without existing in the database.
	waUser := &WebAuthnUser{User: store.User{ID: candidateID, Username: req.Username, DisplayName: displayName}}
	creation, sessionData, err := wa.BeginRegistration(waUser,
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementPreferred))
	if err != nil {
		h.serverError(w, r, "passkey signup: begin registration", err)
		return
	}
	flowID, err := h.saveFlow(r.Context(), nil, "signup", origin, &signupFlowData{
		UserID:      candidateID,
		Username:    req.Username,
		DisplayName: displayName,
	}, sessionData)
	if err != nil {
		h.serverError(w, r, "passkey signup: save flow", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"flow_id": flowID, "options": creation})
}

// PasskeySignupFinish verifies the new credential and only then creates the
// account (passwordless) and starts a session.
func (h *Handlers) PasskeySignupFinish(w http.ResponseWriter, r *http.Request) {
	flowID := r.URL.Query().Get("flow")
	flow, err := h.loadFlowAny(r.Context(), flowID, "signup")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid or expired signup flow")
		return
	}
	env := decodeFlowEnvelope(flow.SessionData)
	if env.Signup == nil {
		httpx.WriteError(w, http.StatusBadRequest, "corrupt signup flow")
		return
	}
	sd, err := env.sessionData()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "corrupt signup flow")
		return
	}
	wa, err := h.rpForFlow(env)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "this origin is not allowed to use passkeys")
		return
	}

	candidate := store.User{
		ID:          env.Signup.UserID,
		Username:    env.Signup.Username,
		DisplayName: env.Signup.DisplayName,
	}
	waUser := &WebAuthnUser{User: candidate}
	cred, err := wa.FinishRegistration(waUser, *sd, r)
	if err != nil {
		h.record(r, nil, candidate.Username, "signup", "passkey", "failure", map[string]any{"error": err.Error()})
		httpx.WriteError(w, http.StatusBadRequest, "passkey signup failed")
		return
	}

	user, err := h.store.CreateUserWithID(r.Context(), candidate.ID, candidate.Username, candidate.DisplayName)
	if err != nil {
		if store.IsUniqueViolation(err) {
			h.record(r, nil, candidate.Username, "signup", "passkey", "failure", map[string]any{"reason": "username_taken"})
			httpx.WriteError(w, http.StatusConflict, "username already taken")
			return
		}
		h.serverError(w, r, "passkey signup: create user", err)
		return
	}
	if err := h.store.AddWebAuthnCredential(r.Context(), user.ID, cred); err != nil {
		h.serverError(w, r, "passkey signup: store credential", err)
		return
	}
	_ = h.store.DeleteFlow(r.Context(), flowID)
	h.record(r, &user.ID, user.Username, "signup", "passkey", "success", nil)
	h.issueSession(w, r, user, "passkey")
}

// ---------------------------------------------------------------------------
// TOTP (authenticated)
// ---------------------------------------------------------------------------

type totpConfirmRequest struct {
	Code string `json:"code"`
}

// TOTPEnroll generates a new TOTP secret and recovery codes for the current
// user. The secret is stored encrypted and unconfirmed until TOTPConfirm.
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
	if err := h.store.UpsertTOTPSecret(r.Context(), user.ID, sealed); err != nil {
		h.serverError(w, r, "totp: store secret", err)
		return
	}
	codes, hashes, err := GenerateRecoveryCodes(10)
	if err != nil {
		h.serverError(w, r, "totp: recovery codes", err)
		return
	}
	if err := h.store.AddRecoveryCodes(r.Context(), user.ID, hashes); err != nil {
		h.serverError(w, r, "totp: store recovery codes", err)
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
	var req totpConfirmRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	secret, err := h.decryptTOTPSecret(r.Context(), user.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "no pending TOTP enrollment")
		return
	}
	if !ValidateTOTP(req.Code, secret) {
		httpx.WriteError(w, http.StatusBadRequest, "invalid code")
		return
	}
	if err := h.store.ConfirmTOTP(r.Context(), user.ID); err != nil {
		h.serverError(w, r, "totp: confirm", err)
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
	var req totpConfirmRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if !h.verifySecondFactor(r.Context(), user.ID, req.Code) {
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
// flow envelope — what actually goes into webauthn_flows.session_data
// ---------------------------------------------------------------------------

// flowEnvelope wraps the go-webauthn ceremony state with the origin the flow
// was created for (so begin and finish always agree on the relying party) and,
// for passkey-first signup, the candidate account that will be created only if
// the ceremony succeeds.
type flowEnvelope struct {
	V       int             `json:"v"`
	Origin  string          `json:"origin,omitempty"`
	Signup  *signupFlowData `json:"signup,omitempty"`
	Session json.RawMessage `json:"session"`
}

// signupFlowData is the not-yet-created account carried through a passkey
// signup ceremony.
type signupFlowData struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

// decodeFlowEnvelope reads a stored flow blob. Flows written before the
// envelope existed stored the bare webauthn.SessionData; those are wrapped
// on the fly (with no pinned origin, which resolves to the static relying
// party) so in-flight ceremonies survive an upgrade.
func decodeFlowEnvelope(blob []byte) flowEnvelope {
	var env flowEnvelope
	if err := json.Unmarshal(blob, &env); err == nil && len(env.Session) > 0 {
		return env
	}
	return flowEnvelope{V: 0, Session: json.RawMessage(blob)}
}

// sessionData unpacks the library's ceremony state.
func (e flowEnvelope) sessionData() (*webauthn.SessionData, error) {
	var sd webauthn.SessionData
	if err := json.Unmarshal(e.Session, &sd); err != nil {
		return nil, err
	}
	return &sd, nil
}

// rpForFlow resolves the relying party a flow was pinned to when it began.
func (h *Handlers) rpForFlow(env flowEnvelope) (*webauthn.WebAuthn, error) {
	wa, _, err := h.rp.ForOrigin(env.Origin)
	return wa, err
}

// ---------------------------------------------------------------------------
// internal helpers
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
	http.SetCookie(w, h.sessionCookie(token, cookieExpiry, cookieMaxAge))
	httpx.WriteJSON(w, http.StatusOK, sessionResponse{
		Token:     token,
		ExpiresAt: dbExpiry, // nil (omitted) when the session never expires
		User:      h.userResponse(r.Context(), user),
	})
}

// persistentSessionCookieTTL is how far in the future the secondary session
// cookie is dated when the session itself never expires. The server-side
// session (the source of truth) has no expiry and the frontend's Bearer token
// lives in localStorage indefinitely; this only affects the cookie, which
// browsers additionally clamp to their own maximum (commonly ~400 days). It is
// re-set on every login, so a returning user's cookie stays fresh.
const persistentSessionCookieTTL = 365 * 24 * time.Hour

// computeSessionExpiry maps the configured session TTL to the database expiry
// (nil = never expires, stored as SQL NULL) and the cookie's Expires/Max-Age.
// A TTL of zero or less means the session never expires on its own: no logout
// for inactivity, no logout on browser close — only an explicit logout ends
// it. A positive TTL opts into an absolute expiry. Pure and time-injected so
// it is trivially unit-testable.
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
		MaxAge:   maxAge, // Max-Age wins over Expires in modern browsers; both set for reach
		HttpOnly: true,
		Secure:   h.set.CookieSecure,
		SameSite: http.SameSiteNoneMode, // cross-site: frontend and API differ
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

func (h *Handlers) userResponse(ctx context.Context, u store.User) userResponse {
	roles, _ := h.store.UserRoles(ctx, u.ID)
	totp, _ := h.totpEnabled(ctx, u.ID)
	pk, _ := h.store.CountWebAuthnCredentials(ctx, u.ID)
	return userResponse{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		Roles:       roles,
		TOTPEnabled: totp,
		Passkeys:    pk,
	}
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

func (h *Handlers) decryptTOTPSecret(ctx context.Context, userID string) (string, error) {
	sealed, _, err := h.store.GetTOTPSecret(ctx, userID)
	if err != nil {
		return "", err
	}
	plain, err := DecryptSecret(h.set.TOTPKey, sealed)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// verifySecondFactor accepts either a valid TOTP code or an unused recovery
// code. It returns true on success and consumes a recovery code when used.
func (h *Handlers) verifySecondFactor(ctx context.Context, userID, code string) bool {
	code = strings.TrimSpace(code)
	if code == "" {
		return false
	}
	if secret, err := h.decryptTOTPSecret(ctx, userID); err == nil {
		if ValidateTOTP(code, secret) {
			return true
		}
	}
	// Fall back to a one-time recovery code.
	used, err := h.store.UseRecoveryCode(ctx, userID, HashRecoveryCode(code))
	if err != nil {
		h.log.WarnContext(ctx, "recovery code check", slog.String("error", err.Error()))
		return false
	}
	return used
}

func (h *Handlers) saveFlow(ctx context.Context, userID *string, kind, origin string, signup *signupFlowData, sd *webauthn.SessionData) (string, error) {
	sess, err := json.Marshal(sd)
	if err != nil {
		return "", err
	}
	blob, err := json.Marshal(flowEnvelope{V: 1, Origin: origin, Signup: signup, Session: sess})
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

func (h *Handlers) loadFlow(ctx context.Context, id, kind string, expectUser *string) (flowEnvelope, error) {
	flow, err := h.loadFlowAny(ctx, id, kind)
	if err != nil {
		return flowEnvelope{}, err
	}
	if expectUser != nil {
		if flow.UserID == nil || *flow.UserID != *expectUser {
			return flowEnvelope{}, errors.New("auth: flow does not belong to user")
		}
	}
	return decodeFlowEnvelope(flow.SessionData), nil
}

func (h *Handlers) loadFlowAny(ctx context.Context, id, kind string) (store.Flow, error) {
	if id == "" {
		return store.Flow{}, errors.New("auth: missing flow id")
	}
	flow, err := h.store.GetFlow(ctx, id)
	if err != nil {
		return store.Flow{}, err
	}
	if flow.Kind != kind {
		return store.Flow{}, errors.New("auth: flow kind mismatch")
	}
	return flow, nil
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
	h.log.ErrorContext(r.Context(), msg,
		slog.String("error", err.Error()),
		slog.String("request_id", httpx.RequestIDFromContext(r.Context())),
	)
	httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
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

func newFlowID() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// newUUIDv4 generates an RFC 4122 version 4 UUID string. Kept local (like the
// store's id::text convention) to avoid pulling in a UUID dependency for the
// one place the application, rather than Postgres, must mint an id: the
// passkey-signup candidate user, which has to exist as a WebAuthn user handle
// before any row exists in the database.
func newUUIDv4() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // RFC 4122 variant
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// decodeJSON reads a size-limited JSON body into dst, writing a 400 on failure.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MiB
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
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
