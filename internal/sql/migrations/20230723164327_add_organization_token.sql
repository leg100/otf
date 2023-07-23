-- +goose Up
CREATE TABLE IF NOT EXISTS organization_tokens (
    organization_token_id   TEXT NOT NULL,
    created_at              TIMESTAMPTZ NOT NULL,
    organization_name       TEXT REFERENCES organizations (name) ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                            PRIMARY KEY (organization_token_id),
                            UNIQUE (organization_name)
);

-- +goose Down
DROP TABLE IF EXISTS organization_tokens;
