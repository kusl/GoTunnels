// Package store is the single data-access layer. Every SQL query in the
// application lives here, so the rest of the code deals in Go types and never
// in SQL strings. It wraps a pgx connection pool.
//
// UUID handling note: to avoid pulling in a separate UUID dependency, user ids
// are exchanged as plain strings. Queries always select `id::text` and always
// bind uuid parameters with an explicit `$N::uuid` cast, which sidesteps the
// lack of an implicit text->uuid cast in PostgreSQL.
package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when a lookup finds no row.
var ErrNotFound = errors.New("store: not found")

// Store is the data-access facade over a pgx pool.
type Store struct {
	pool *pgxpool.Pool
}

// New wraps a pool.
func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

// Pool exposes the underlying pool for health checks.
func (s *Store) Pool() *pgxpool.Pool { return s.pool }

// Ping verifies database connectivity.
func (s *Store) Ping(ctx context.Context) error { return s.pool.Ping(ctx) }

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// User is an application user.
type User struct {
	ID          string
	Username    string
	DisplayName string
	CreatedAt   time.Time
}

// Session is a server-side session record.
type Session struct {
	ID         string
	UserID     string
	AuthMethod string
	CreatedAt  time.Time
	LastSeenAt time.Time
	ExpiresAt  time.Time
	RevokedAt  *time.Time
}

// Flow is an in-progress WebAuthn ceremony's stored state.
type Flow struct {
	ID          string
	UserID      *string
	Kind        string
	SessionData []byte
	ExpiresAt   time.Time
}

// Activity is one audit-log row.
type Activity struct {
	ID         int64           `json:"id"`
	UserID     *string         `json:"user_id,omitempty"`
	Username   string          `json:"username"`
	EventType  string          `json:"event_type"`
	AuthMethod string          `json:"auth_method"`
	Outcome    string          `json:"outcome"`
	IPHash     string          `json:"ip_hash"`
	UserAgent  string          `json:"user_agent,omitempty"`
	Detail     json.RawMessage `json:"detail,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

// ActivityInput is the payload for recording an activity event.
type ActivityInput struct {
	UserID     *string
	Username   string
	EventType  string
	AuthMethod string
	Outcome    string
	IPHash     string
	UserAgent  string
	Detail     map[string]any
}

// CSPReportInput is a normalised CSP violation ready to persist.
type CSPReportInput struct {
	DocumentURI        string
	Referrer           string
	BlockedURI         string
	ViolatedDirective  string
	EffectiveDirective string
	OriginalPolicy     string
	Disposition        string
	SourceFile         string
	LineNumber         int
	ColumnNumber       int
	StatusCode         int
	ScriptSample       string
	IPHash             string
	UserAgent          string
	Raw                json.RawMessage
}

// ---------------------------------------------------------------------------
// Users & roles
// ---------------------------------------------------------------------------

// CreateUser inserts a user and grants the default "user" role atomically.
func (s *Store) CreateUser(ctx context.Context, username, displayName string) (User, error) {
	lower := normalizeUsername(username)
	if displayName == "" {
		displayName = username
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return User{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var u User
	err = tx.QueryRow(ctx, `
		INSERT INTO users (username, username_lower, display_name)
		VALUES ($1, $2, $3)
		RETURNING id::text, username, display_name, created_at`,
		username, lower, displayName,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.CreatedAt)
	if err != nil {
		return User{}, err
	}
	if _, err = tx.Exec(ctx,
		`INSERT INTO user_roles (user_id, role) VALUES ($1::uuid, 'user')`, u.ID); err != nil {
		return User{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return User{}, err
	}
	return u, nil
}

// GetUserByUsername looks up a user case-insensitively.
func (s *Store) GetUserByUsername(ctx context.Context, username string) (User, error) {
	lower := normalizeUsername(username)
	var u User
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, username, display_name, created_at
		FROM users WHERE username_lower = $1`, lower,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.CreatedAt)
	return u, mapErr(err)
}

// GetUserByID looks up a user by id.
func (s *Store) GetUserByID(ctx context.Context, id string) (User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, username, display_name, created_at
		FROM users WHERE id = $1::uuid`, id,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.CreatedAt)
	return u, mapErr(err)
}

// UsernameExists reports whether a username is already taken.
func (s *Store) UsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE username_lower = $1)`,
		normalizeUsername(username),
	).Scan(&exists)
	return exists, err
}

// UserRoles returns the role names granted to a user.
func (s *Store) UserRoles(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT role FROM user_roles WHERE user_id = $1::uuid ORDER BY role`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []string
	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}
	return roles, rows.Err()
}

// ---------------------------------------------------------------------------
// Password credentials
// ---------------------------------------------------------------------------

// SetPassword upserts the password hash for a user.
func (s *Store) SetPassword(ctx context.Context, userID, hash string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO password_credentials (user_id, password_hash, updated_at)
		VALUES ($1::uuid, $2, now())
		ON CONFLICT (user_id)
		DO UPDATE SET password_hash = EXCLUDED.password_hash, updated_at = now()`,
		userID, hash)
	return err
}

// GetPasswordHash returns the stored PHC hash, or ErrNotFound.
func (s *Store) GetPasswordHash(ctx context.Context, userID string) (string, error) {
	var hash string
	err := s.pool.QueryRow(ctx,
		`SELECT password_hash FROM password_credentials WHERE user_id = $1::uuid`,
		userID).Scan(&hash)
	return hash, mapErr(err)
}

// ---------------------------------------------------------------------------
// WebAuthn credentials
// ---------------------------------------------------------------------------

// AddWebAuthnCredential stores a freshly registered credential. The full
// webauthn.Credential is persisted as JSON (the source of truth for later
// reconstruction) alongside broken-out columns used for indexing/uniqueness.
func (s *Store) AddWebAuthnCredential(ctx context.Context, userID string, cred *webauthn.Credential) error {
	blob, err := json.Marshal(cred)
	if err != nil {
		return err
	}
	transports := make([]string, 0, len(cred.Transport))
	for _, t := range cred.Transport {
		transports = append(transports, string(t))
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO webauthn_credentials
			(user_id, credential_id, public_key, attestation_type, aaguid,
			 sign_count, transports, clone_warning, credential)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9)`,
		userID,
		cred.ID,
		cred.PublicKey,
		cred.AttestationType,
		cred.Authenticator.AAGUID,
		int64(cred.Authenticator.SignCount),
		transports,
		cred.Authenticator.CloneWarning,
		blob,
	)
	return err
}

// GetWebAuthnCredentials reconstructs a user's credentials from stored JSON.
func (s *Store) GetWebAuthnCredentials(ctx context.Context, userID string) ([]webauthn.Credential, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT credential FROM webauthn_credentials WHERE user_id = $1::uuid ORDER BY id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var creds []webauthn.Credential
	for rows.Next() {
		var blob []byte
		if err := rows.Scan(&blob); err != nil {
			return nil, err
		}
		var c webauthn.Credential
		if err := json.Unmarshal(blob, &c); err != nil {
			return nil, err
		}
		creds = append(creds, c)
	}
	return creds, rows.Err()
}

// CountWebAuthnCredentials returns how many passkeys a user has.
func (s *Store) CountWebAuthnCredentials(ctx context.Context, userID string) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx,
		`SELECT count(*) FROM webauthn_credentials WHERE user_id = $1::uuid`, userID).Scan(&n)
	return n, err
}

// UpdateWebAuthnCredential persists post-login changes (sign count, flags).
func (s *Store) UpdateWebAuthnCredential(ctx context.Context, userID string, cred *webauthn.Credential) error {
	blob, err := json.Marshal(cred)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		UPDATE webauthn_credentials
		SET sign_count = $1, clone_warning = $2, credential = $3, last_used_at = now()
		WHERE user_id = $4::uuid AND credential_id = $5`,
		int64(cred.Authenticator.SignCount),
		cred.Authenticator.CloneWarning,
		blob,
		userID,
		cred.ID,
	)
	return err
}

// ---------------------------------------------------------------------------
// WebAuthn flows (ceremony state)
// ---------------------------------------------------------------------------

// SaveFlow stores ceremony state keyed by a random flow id.
func (s *Store) SaveFlow(ctx context.Context, f Flow) error {
	var uid any
	if f.UserID != nil {
		uid = *f.UserID
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO webauthn_flows (id, user_id, kind, session_data, expires_at)
		VALUES ($1, $2::uuid, $3, $4, $5)`,
		f.ID, uid, f.Kind, f.SessionData, f.ExpiresAt)
	return err
}

// GetFlow fetches ceremony state, or ErrNotFound if missing/expired.
func (s *Store) GetFlow(ctx context.Context, id string) (Flow, error) {
	var f Flow
	var uid *string
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id::text, kind, session_data, expires_at
		FROM webauthn_flows WHERE id = $1 AND expires_at > now()`, id,
	).Scan(&f.ID, &uid, &f.Kind, &f.SessionData, &f.ExpiresAt)
	if err != nil {
		return Flow{}, mapErr(err)
	}
	f.UserID = uid
	return f, nil
}

// DeleteFlow removes ceremony state (called once consumed).
func (s *Store) DeleteFlow(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM webauthn_flows WHERE id = $1`, id)
	return err
}

// ---------------------------------------------------------------------------
// TOTP
// ---------------------------------------------------------------------------

// UpsertTOTPSecret stores an unconfirmed encrypted TOTP secret.
func (s *Store) UpsertTOTPSecret(ctx context.Context, userID string, encrypted []byte) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO totp_secrets (user_id, secret_encrypted, confirmed, created_at)
		VALUES ($1::uuid, $2, false, now())
		ON CONFLICT (user_id)
		DO UPDATE SET secret_encrypted = EXCLUDED.secret_encrypted,
		              confirmed = false, created_at = now(), confirmed_at = NULL`,
		userID, encrypted)
	return err
}

// ConfirmTOTP marks a user's TOTP secret confirmed.
func (s *Store) ConfirmTOTP(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE totp_secrets SET confirmed = true, confirmed_at = now() WHERE user_id = $1::uuid`,
		userID)
	return err
}

// GetTOTPSecret returns the encrypted secret and confirmation state.
func (s *Store) GetTOTPSecret(ctx context.Context, userID string) (encrypted []byte, confirmed bool, err error) {
	err = s.pool.QueryRow(ctx,
		`SELECT secret_encrypted, confirmed FROM totp_secrets WHERE user_id = $1::uuid`,
		userID).Scan(&encrypted, &confirmed)
	return encrypted, confirmed, mapErr(err)
}

// DeleteTOTP disables TOTP for a user (secret + recovery codes).
func (s *Store) DeleteTOTP(ctx context.Context, userID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `DELETE FROM totp_recovery_codes WHERE user_id = $1::uuid`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM totp_secrets WHERE user_id = $1::uuid`, userID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// AddRecoveryCodes stores hashed one-time recovery codes.
func (s *Store) AddRecoveryCodes(ctx context.Context, userID string, hashes []string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `DELETE FROM totp_recovery_codes WHERE user_id = $1::uuid`, userID); err != nil {
		return err
	}
	for _, h := range hashes {
		if _, err := tx.Exec(ctx,
			`INSERT INTO totp_recovery_codes (user_id, code_hash) VALUES ($1::uuid, $2)`,
			userID, h); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// UseRecoveryCode marks a matching unused recovery code as used. It reports
// whether a code was consumed.
func (s *Store) UseRecoveryCode(ctx context.Context, userID, codeHash string) (bool, error) {
	ct, err := s.pool.Exec(ctx, `
		UPDATE totp_recovery_codes SET used_at = now()
		WHERE user_id = $1::uuid AND code_hash = $2 AND used_at IS NULL`,
		userID, codeHash)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() > 0, nil
}

// ---------------------------------------------------------------------------
// Sessions
// ---------------------------------------------------------------------------

// CreateSession inserts a new session row.
func (s *Store) CreateSession(ctx context.Context, id, userID, authMethod string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sessions (id, user_id, auth_method, expires_at)
		VALUES ($1, $2::uuid, $3, $4)`,
		id, userID, authMethod, expiresAt)
	return err
}

// GetSession fetches a live (non-revoked, non-expired) session.
func (s *Store) GetSession(ctx context.Context, id string) (Session, error) {
	var sess Session
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id::text, auth_method, created_at, last_seen_at, expires_at, revoked_at
		FROM sessions
		WHERE id = $1 AND revoked_at IS NULL AND expires_at > now()`, id,
	).Scan(&sess.ID, &sess.UserID, &sess.AuthMethod, &sess.CreatedAt,
		&sess.LastSeenAt, &sess.ExpiresAt, &sess.RevokedAt)
	return sess, mapErr(err)
}

// TouchSession updates last_seen_at.
func (s *Store) TouchSession(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE sessions SET last_seen_at = now() WHERE id = $1`, id)
	return err
}

// RevokeSession marks a session revoked (logout).
func (s *Store) RevokeSession(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE sessions SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL`, id)
	return err
}

// ---------------------------------------------------------------------------
// Activity log
// ---------------------------------------------------------------------------

// InsertActivity records an audit event.
func (s *Store) InsertActivity(ctx context.Context, in ActivityInput) error {
	detail := in.Detail
	if detail == nil {
		detail = map[string]any{}
	}
	blob, err := json.Marshal(detail)
	if err != nil {
		return err
	}
	outcome := in.Outcome
	if outcome == "" {
		outcome = "success"
	}
	var uid any
	if in.UserID != nil {
		uid = *in.UserID
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO activity_log
			(user_id, username, event_type, auth_method, outcome, ip_hash, user_agent, detail)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8)`,
		uid, in.Username, in.EventType, in.AuthMethod, outcome, in.IPHash, in.UserAgent, blob)
	return err
}

// ListActivityForUser returns a user's most recent audit events.
func (s *Store) ListActivityForUser(ctx context.Context, userID string, limit int) ([]Activity, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id::text, username, event_type, auth_method, outcome,
		       ip_hash, user_agent, detail, created_at
		FROM activity_log
		WHERE user_id = $1::uuid
		ORDER BY created_at DESC
		LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Activity
	for rows.Next() {
		var a Activity
		if err := rows.Scan(&a.ID, &a.UserID, &a.Username, &a.EventType, &a.AuthMethod,
			&a.Outcome, &a.IPHash, &a.UserAgent, &a.Detail, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------------------
// Health check log
// ---------------------------------------------------------------------------

// InsertHealthCheck records the outcome of a readiness probe.
func (s *Store) InsertHealthCheck(ctx context.Context, checkName, status string, latencyMs float64, detail string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO health_check_log (check_name, status, latency_ms, detail)
		VALUES ($1, $2, $3, $4)`,
		checkName, status, latencyMs, detail)
	return err
}

// ---------------------------------------------------------------------------
// CSP reports
// ---------------------------------------------------------------------------

// InsertCSPReport persists a normalised CSP violation report.
func (s *Store) InsertCSPReport(ctx context.Context, in CSPReportInput) error {
	raw := in.Raw
	if len(raw) == 0 {
		raw = json.RawMessage("{}")
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO csp_reports
			(document_uri, referrer, blocked_uri, violated_directive, effective_directive,
			 original_policy, disposition, source_file, line_number, column_number,
			 status_code, script_sample, ip_hash, user_agent, raw)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		in.DocumentURI, in.Referrer, in.BlockedURI, in.ViolatedDirective, in.EffectiveDirective,
		in.OriginalPolicy, in.Disposition, in.SourceFile, in.LineNumber, in.ColumnNumber,
		in.StatusCode, in.ScriptSample, in.IPHash, in.UserAgent, []byte(raw))
	return err
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// normalizeUsername lowercases and trims a username for case-insensitive
// comparison. Kept here so store lookups and inserts agree on the rule.
func normalizeUsername(u string) string {
	return toLowerTrim(u)
}

func toLowerTrim(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		out = append(out, r)
	}
	// trim spaces
	start, end := 0, len(out)
	for start < end && isSpace(out[start]) {
		start++
	}
	for end > start && isSpace(out[end-1]) {
		end--
	}
	return string(out[start:end])
}

func isSpace(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' }
