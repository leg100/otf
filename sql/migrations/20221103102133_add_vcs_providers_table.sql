-- +goose Up
CREATE TABLE IF NOT EXISTS vcs_providers (
    vcs_provider_id   TEXT,
    token             TEXT,
    created_at        TIMESTAMPTZ NOT NULL,
    name              TEXT        NOT NULL,
    organization_name TEXT REFERENCES organizations (name) ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                      PRIMARY KEY (vcs_provider_id)
);

-- +goose Down
DROP TABLE IF EXISTS vcs_providers;
