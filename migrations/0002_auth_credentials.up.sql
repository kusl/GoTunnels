-- 0002_auth_credentials.up.sql
-- Credential material for the three supported authentication methods:
-- password (fallback), WebAuthn/passkeys (primary), and TOTP (optional 2FA).

-- Argon2id password hashes, stored in PHC string format.
CREATE TABLE password_credentials (
    user_id       uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    password_hash text NOT NULL,
    updated_at    timestamptz NOT NULL DEFAULT now()
);

-- Registered WebAuthn credentials (passkeys / security keys).
CREATE TABLE webauthn_credentials (
    id               bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id          uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credential_id    bytea NOT NULL,
    public_key       bytea NOT NULL,
    attestation_type text NOT NULL DEFAULT '',
    aaguid           bytea NOT NULL DEFAULT '\x',
    sign_count       bigint NOT NULL DEFAULT 0,
    transports       text[] NOT NULL DEFAULT '{}',
    clone_warning    boolean NOT NULL DEFAULT false,
    credential       jsonb NOT NULL,
    created_at       timestamptz NOT NULL DEFAULT now(),
    last_used_at     timestamptz
);

CREATE UNIQUE INDEX webauthn_credentials_credential_id_key
    ON webauthn_credentials (credential_id);
CREATE INDEX webauthn_credentials_user_id_idx
    ON webauthn_credentials (user_id);

-- Short-lived server-side state for in-progress WebAuthn ceremonies. The flow
-- id is handed to the browser and echoed back on the finish request, so the
-- ceremony never depends on a cross-site cookie.
CREATE TABLE webauthn_flows (
    id           text PRIMARY KEY,
    user_id      uuid REFERENCES users(id) ON DELETE CASCADE,
    kind         text NOT NULL,
    session_data jsonb NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    expires_at   timestamptz NOT NULL
);

CREATE INDEX webauthn_flows_expires_at_idx ON webauthn_flows (expires_at);

-- Encrypted TOTP shared secrets. The secret is encrypted with AES-256-GCM
-- before it ever touches the database (see internal/auth/totp.go).
CREATE TABLE totp_secrets (
    user_id          uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    secret_encrypted bytea NOT NULL,
    confirmed        boolean NOT NULL DEFAULT false,
    created_at       timestamptz NOT NULL DEFAULT now(),
    confirmed_at     timestamptz
);

-- One-time recovery codes for TOTP. Only the hash is stored.
CREATE TABLE totp_recovery_codes (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash  text NOT NULL,
    used_at    timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX totp_recovery_codes_user_id_idx ON totp_recovery_codes (user_id);
