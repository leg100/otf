-- +goose Up
CREATE TABLE IF NOT EXISTS agent_tokens (
    token_id          TEXT,
    token             TEXT,
    created_at        TIMESTAMPTZ NOT NULL,
    description       TEXT        NOT NULL,
    organization_name TEXT REFERENCES organizations (name) ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                      PRIMARY KEY (token_id),
                      UNIQUE (token)
);

-- +goose Down
DROP TABLE IF EXISTS agent_tokens;
