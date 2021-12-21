-- +goose Up
CREATE TABLE IF NOT EXISTS sessions (
    token text,
    data bytea not null,
    expiry timestamptz not null,
    user_id text REFERENCES users ON UPDATE CASCADE ON DELETE CASCADE,
    PRIMARY KEY (token)
);

CREATE TABLE IF NOT EXISTS users (
    user_id text,
    username text,
    created_at timestamptz,
    updated_at timestamptz,
    PRIMARY KEY (user_id)
);

-- +goose Down
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
