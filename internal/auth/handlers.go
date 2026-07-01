package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

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
	wa    *webauthn.WebAuthn
	rec   *activity.Recorder
	log   *slog.Logger
	set   Settings
}

// NewHandlers builds the auth handler set.
func NewHandlers(s *store.Store, wa *webauthn.WebAuthn, rec *activity.Recorder, log *slog.Logger, set Settings) *Handlers {
	if set.FlowTTL <= 0 {
		set.FlowTTL = 10 * time.Minute
	}
	if set.CookieName == "" {
		set.CookieName = "gotunnels_session"
	}
	return &Handlers{store: s, wa: wa, rec: rec, log: log, set: set}
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
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      userResponse `json:"user"`
}

// ---------------------------------------------------------------------------
// signup / login / logout / me / activity
// ---------------------------------------------------------------------------

// Signup creates an account with a password (the always-present fallback
// credential) and immediately establishes a session.
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
	waUser, err := h.buildWAUser(r.Context(), user)
	if err != nil {
		h.serverError(w, r, "passkey: load user", err)
		return
	}
	creation, sessionData, err := h.wa.BeginRegistration(waUser)
	if err != nil {
		h.serverError(w, r, "passkey: begin registration", err)
		return
	}
	flowID, err := h.saveFlow(r.Context(), &user.ID, "register", sessionData)
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
	session, err := h.loadFlow(r.Context(), flowID, "register", &user.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid or expired registration flow")
		return
	}
	waUser, err := h.buildWAUser(r.Context(), user)
	if err != nil {
		h.serverError(w, r, "passkey: load user", err)
		return
	}
	cred, err := h.wa.FinishRegistration(waUser, *session, r)
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
	assertion, sessionData, err := h.wa.BeginLogin(waUser)
	if err != nil {
		h.serverError(w, r, "passkey: begin login", err)
		return
	}
	flowID, err := h.saveFlow(r.Context(), &user.ID, "login", sessionData)
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
	waUser, err := h.buildWAUser(r.Context(), user)
	if err != nil {
		h.serverError(w, r, "passkey: load user", err)
		return
	}
	var sd webauthn.SessionData
	if err := json.Unmarshal(flow.SessionData, &sd); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "corrupt login flow")
		return
	}
	cred, err := h.wa.FinishLogin(waUser, sd, r)
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
// internal helpers
// ---------------------------------------------------------------------------

func (h *Handlers) issueSession(w http.ResponseWriter, r *http.Request, user store.User, method string) {
	token, id, err := NewSessionToken()
	if err != nil {
		h.serverError(w, r, "session: token", err)
		return
	}
	expires := time.Now().Add(h.set.SessionTTL)
	if err := h.store.CreateSession(r.Context(), id, user.ID, method, expires); err != nil {
		h.serverError(w, r, "session: create", err)
		return
	}
	http.SetCookie(w, h.sessionCookie(token, expires))
	httpx.WriteJSON(w, http.StatusOK, sessionResponse{
		Token:     token,
		ExpiresAt: expires,
		User:      h.userResponse(r.Context(), user),
	})
}

func (h *Handlers) sessionCookie(token string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     h.set.CookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
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

func (h *Handlers) saveFlow(ctx context.Context, userID *string, kind string, sd *webauthn.SessionData) (string, error) {
	blob, err := json.Marshal(sd)
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

func (h *Handlers) loadFlow(ctx context.Context, id, kind string, expectUser *string) (*webauthn.SessionData, error) {
	flow, err := h.loadFlowAny(ctx, id, kind)
	if err != nil {
		return nil, err
	}
	if expectUser != nil {
		if flow.UserID == nil || *flow.UserID != *expectUser {
			return nil, errors.New("auth: flow does not belong to user")
		}
	}
	var sd webauthn.SessionData
	if err := json.Unmarshal(flow.SessionData, &sd); err != nil {
		return nil, err
	}
	return &sd, nil
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
