-- 0009_totp_replay_guard.up.sql
-- RFC 6238 section 5.2: a TOTP code must be accepted at most once. The highest
-- accepted time step is recorded per user, and a code is only accepted when
-- its step is strictly greater than the last one used.
ALTER TABLE totp_secrets ADD COLUMN last_used_step bigint;
