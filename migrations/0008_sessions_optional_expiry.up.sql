-- 0008_sessions_optional_expiry.up.sql
-- Make session expiry OPTIONAL so a session can be persistent: it lives until
-- the user explicitly logs out and never dies on its own. Before this,
-- expires_at was NOT NULL and every session expired at a fixed wall-clock time
-- (24h after login by default) regardless of activity — the behaviour we are
-- deliberately moving away from.
--
-- A NULL expires_at now means "no expiry". The live-session lookup in
-- internal/store treats (expires_at IS NULL OR expires_at > now()) as valid.
-- Existing rows keep their concrete expiry, so this migration is backward
-- compatible; only sessions minted while GOTUNNELS_SESSION_TTL=0 (the new
-- default) are written with NULL.
ALTER TABLE sessions ALTER COLUMN expires_at DROP NOT NULL;

-- Recreate the expiry index as a PARTIAL index over non-NULL expiries: rows
-- that never expire carry no entry, so the index stays small and the
-- "expires_at > now()" half of the lookup stays a tight range scan. Recreated
-- rather than left in place so the intent is explicit in the schema history.
DROP INDEX IF EXISTS sessions_expires_at_idx;
CREATE INDEX sessions_expires_at_idx ON sessions (expires_at) WHERE expires_at IS NOT NULL;
