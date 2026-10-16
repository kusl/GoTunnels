-- 0009_totp_replay_guard.down.sql
ALTER TABLE totp_secrets DROP COLUMN IF EXISTS last_used_step;
