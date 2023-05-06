-- +goose Up
CREATE TABLE IF NOT EXISTS clouds (
    name TEXT PRIMARY KEY
);
INSERT INTO clouds (name) VALUES
	('github'),
	('gitlab');

CREATE TABLE IF NOT EXISTS vcs_providers (
    vcs_provider_id         TEXT,
    token                   TEXT        NOT NULL,
    created_at              TIMESTAMPTZ NOT NULL,
    name                    TEXT        NOT NULL,
    cloud                   TEXT REFERENCES clouds (name) ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    organization_name       TEXT REFERENCES organizations (name) ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                            PRIMARY KEY (vcs_provider_id)
);

-- +goose Down
DROP TABLE IF EXISTS vcs_providers;
DROP TABLE IF EXISTS clouds;
