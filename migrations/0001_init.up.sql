-- 0001_init.up.sql
-- Core identity tables. gen_random_uuid() is part of PostgreSQL core since
-- version 13, so no extension is required.

CREATE TABLE users (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    username       text NOT NULL,
    username_lower text NOT NULL,
    display_name   text NOT NULL DEFAULT '',
    created_at     timestamptz NOT NULL DEFAULT now()
);

-- Case-insensitive uniqueness without requiring the citext extension: the
-- application lowercases usernames into username_lower for lookups.
CREATE UNIQUE INDEX users_username_lower_key ON users (username_lower);

CREATE TABLE roles (
    name        text PRIMARY KEY,
    description text NOT NULL DEFAULT ''
);

INSERT INTO roles (name, description) VALUES
    ('user',  'Standard user'),
    ('admin', 'Administrator')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE user_roles (
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role       text NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
    granted_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, role)
);
