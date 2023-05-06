-- +goose Up
CREATE TABLE IF NOT EXISTS module_statuses (
    status TEXT PRIMARY KEY
);
INSERT INTO module_statuses (status) VALUES
	('pending'),
	('no_version_tags'),
	('setup_failed'),
	('setup_complete');

CREATE TABLE IF NOT EXISTS modules (
    module_id       TEXT,
    created_at      TIMESTAMPTZ NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL,
    name            TEXT        NOT NULL,
    provider        TEXT        NOT NULL,
    status          TEXT REFERENCES module_statuses ON UPDATE CASCADE NOT NULL,
    organization_id TEXT REFERENCES organizations ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                    PRIMARY KEY (module_id),
                    UNIQUE (organization_id, provider, name)
);

CREATE TABLE IF NOT EXISTS module_version_statuses (
    status TEXT PRIMARY KEY
);
INSERT INTO module_version_statuses (status) VALUES
	('pending'),
	('cloning'),
	('clone_failed'),
	('reg_ingress_req_failed'),
	('reg_ingressing'),
	('reg_ingress_failed'),
	('ok');

CREATE TABLE IF NOT EXISTS module_versions (
    module_version_id TEXT,
    version           TEXT NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL,
    updated_at        TIMESTAMPTZ NOT NULL,
    status            TEXT REFERENCES module_version_statuses ON UPDATE CASCADE NOT NULL,
    status_error      TEXT,
    module_id         TEXT REFERENCES modules ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                      PRIMARY KEY (module_version_id),
                      UNIQUE (module_id, version)
);

CREATE TABLE IF NOT EXISTS module_tarballs (
    tarball           BYTEA NOT NULL,
    module_version_id TEXT REFERENCES module_versions ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                    UNIQUE (module_version_id)
);

CREATE TABLE IF NOT EXISTS module_repos (
    -- do not cascade deletes because the otfd code relies on getting an error
    -- when attempting to delete a webhook, to determine whether there are any
    -- module repos referencing it; only when no more module repos are referencing
    -- a webhook do we delete it.
    webhook_id        UUID REFERENCES webhooks ON UPDATE CASCADE NOT NULL,
    vcs_provider_id   TEXT REFERENCES vcs_providers ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    module_id         TEXT REFERENCES modules ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                      UNIQUE (module_id)
);

CREATE TABLE IF NOT EXISTS registry_sessions (
    token               TEXT,
    expiry              TIMESTAMPTZ NOT NULL,
    organization_name   TEXT REFERENCES organizations (name) ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                        PRIMARY KEY (token)
);

-- +goose Down
DROP TABLE IF EXISTS registry_sessions;
DROP TABLE IF EXISTS module_repos;
DROP TABLE IF EXISTS module_tarballs;
DROP TABLE IF EXISTS module_versions;
DROP TABLE IF EXISTS module_version_statuses;
DROP TABLE IF EXISTS modules;
DROP TABLE IF EXISTS module_statuses;
