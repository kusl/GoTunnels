-- 0004_activity_and_health.up.sql
-- Audit trail and health observations. IP addresses are never stored in the
-- clear: ip_hash holds sha256(pepper || ip) as lowercase hex.

CREATE TABLE activity_log (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id     uuid REFERENCES users(id) ON DELETE SET NULL,
    username    text NOT NULL DEFAULT '',
    event_type  text NOT NULL,
    auth_method text NOT NULL DEFAULT '',
    outcome     text NOT NULL DEFAULT 'success',
    ip_hash     text NOT NULL DEFAULT '',
    user_agent  text NOT NULL DEFAULT '',
    detail      jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX activity_log_user_id_created_at_idx
    ON activity_log (user_id, created_at DESC);
CREATE INDEX activity_log_created_at_idx
    ON activity_log (created_at DESC);

CREATE TABLE health_check_log (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    check_name text NOT NULL,
    status     text NOT NULL,
    latency_ms double precision NOT NULL DEFAULT 0,
    detail     text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX health_check_log_created_at_idx
    ON health_check_log (created_at DESC);
