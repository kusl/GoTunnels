-- 0007_notes.up.sql
-- Notes: a deliberately plain public microblog. Plain text only (rendering is
-- done with textContent on the client, so nothing here is ever interpreted as
-- HTML), no attachments, no edits. Deletion is a hard DELETE: consistent with
-- this project's privacy posture, "deleted" means the row is gone, not
-- flagged. Rows cascade away if the author's account is ever removed
-- (contrast activity_log, which keeps audit rows with user_id set NULL).
CREATE TABLE notes (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body       text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    -- char_length counts characters (code points), matching the API's rune
    -- count. The API validates first; this is defence in depth.
    CONSTRAINT notes_body_len CHECK (char_length(body) BETWEEN 1 AND 500)
);

-- The feed reads newest-first by id (the PK index already serves that);
-- this index serves per-user lookups and the ownership-checked delete.
CREATE INDEX notes_user_id_idx ON notes (user_id);
