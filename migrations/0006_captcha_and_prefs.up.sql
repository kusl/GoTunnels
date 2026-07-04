-- 0006_captcha_and_prefs.up.sql
-- Per-user CAPTCHA game statistics and a small generic user-preferences store.
--
-- captcha_stats is ONE aggregate row per user, not one row per solve. The
-- page's auto-solver ("Magic Solve") can complete hundreds of puzzles per
-- second at maximum speed, so per-solve rows (and per-solve HTTP requests)
-- would grow unboundedly for zero analytical value. The client batches deltas
-- and the API folds them in atomically; storage stays O(users).

CREATE TABLE captcha_stats (
    user_id        uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    best_streak    bigint NOT NULL DEFAULT 0 CHECK (best_streak >= 0),
    current_streak bigint NOT NULL DEFAULT 0 CHECK (current_streak >= 0),
    total_solves   bigint NOT NULL DEFAULT 0 CHECK (total_solves >= 0),
    manual_solves  bigint NOT NULL DEFAULT 0 CHECK (manual_solves >= 0),
    auto_solves    bigint NOT NULL DEFAULT 0 CHECK (auto_solves >= 0),
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);

-- Supports the leaderboard ordering (best streak, then volume).
CREATE INDEX captcha_stats_leaderboard_idx
    ON captcha_stats (best_streak DESC, total_solves DESC);

-- user_prefs holds small per-user UI preferences (for example whether the
-- CAPTCHA leaderboard is expanded). Keys are constrained to a conservative
-- charset in the application layer; values are short opaque strings. User
-- preferences live server-side so they follow the account across devices;
-- device-specific settings (like the Magic Solve speed, which depends on the
-- device's refresh rate) stay in the browser's localStorage instead.
CREATE TABLE user_prefs (
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key        text NOT NULL CHECK (char_length(key) BETWEEN 1 AND 64),
    value      text NOT NULL DEFAULT '' CHECK (octet_length(value) <= 4096),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, key)
);
