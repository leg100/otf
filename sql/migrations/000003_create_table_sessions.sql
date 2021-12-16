-- +goose Up
CREATE TABLE IF NOT EXISTS sessions (
    token text,
    data bytea not null,
    expiry timestamptz not null,
    PRIMARY KEY (token)
);

-- +goose Down
DROP TABLE IF EXISTS sessions;
