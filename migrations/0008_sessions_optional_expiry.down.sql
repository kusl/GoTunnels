-- 0008_sessions_optional_expiry.down.sql
-- Reverse the optional-expiry change. Any persistent (NULL) sessions must be
-- given a concrete expiry before NOT NULL can be restored; one year out keeps
-- existing logins working across the rollback.
UPDATE sessions SET expires_at = now() + interval '365 days' WHERE expires_at IS NULL;
ALTER TABLE sessions ALTER COLUMN expires_at SET NOT NULL;

DROP INDEX IF EXISTS sessions_expires_at_idx;
CREATE INDEX sessions_expires_at_idx ON sessions (expires_at);
