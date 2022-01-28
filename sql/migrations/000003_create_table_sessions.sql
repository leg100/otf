-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    user_id TEXT,
    username TEXT NOT NULL,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    PRIMARY KEY (user_id)
);

CREATE TABLE IF NOT EXISTS sessions (
    token TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    address TEXT NOT NULL,
    flash JSONB,
    organization TEXT,
    expiry TIMESTAMPTZ NOT NULL,
    user_id TEXT REFERENCES users ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    PRIMARY KEY (token)
);

INSERT INTO users (
    user_id,
    username,
    created_at,
    updated_at
) VALUES (
    'user-anonymous',
    'anonymous',
    now(),
    now()
);

-- +goose Down
DELETE FROM users WHERE user_id = 'user-anonymous';
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
