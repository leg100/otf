-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    user_id text,
    username text not null,
    created_at timestamptz,
    updated_at timestamptz,
    PRIMARY KEY (user_id)
);

CREATE TABLE IF NOT EXISTS sessions (
    token text,
    created_at timestamptz,
    updated_at timestamptz,
    address text not null,
    flash jsonb,
    expiry timestamptz not null,
    user_id text REFERENCES users ON UPDATE CASCADE ON DELETE CASCADE,
    PRIMARY KEY (token)
);

INSERT INTO users (
    user_id,
    username,
    created_at,
    updated_at,
    anonymous
) VALUES (
    'user-anonymous',
    'anonymous',
    now(),
    now(),
)

-- +goose Down
DELETE FROM users WHERE user_id = 'user-anonymous';
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
