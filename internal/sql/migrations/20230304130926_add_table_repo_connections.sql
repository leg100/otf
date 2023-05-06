-- +goose Up
CREATE TABLE IF NOT EXISTS repo_connections (
    webhook_id      UUID REFERENCES webhooks ON UPDATE CASCADE NOT NULL,
    vcs_provider_id TEXT REFERENCES vcs_providers ON UPDATE CASCADE NOT NULL,
    module_id       TEXT REFERENCES modules ON UPDATE CASCADE,
    workspace_id    TEXT REFERENCES workspaces ON UPDATE CASCADE,
                    UNIQUE (module_id),
                    UNIQUE (workspace_id),
                    CHECK ((module_id IS NULL) != (workspace_id IS NULL))
);

INSERT INTO repo_connections (webhook_id, vcs_provider_id, workspace_id)
SELECT webhook_id, vcs_provider_id, workspace_id
FROM workspace_repos;

INSERT INTO repo_connections (webhook_id, vcs_provider_id, module_id)
SELECT webhook_id, vcs_provider_id, module_id
FROM module_repos;

DROP TABLE workspace_repos;

DROP TABLE module_repos;

ALTER TABLE webhooks DROP COLUMN connected;

-- +goose Down
ALTER TABLE webhooks ADD COLUMN connected INTEGER DEFAULT 0;

CREATE TABLE IF NOT EXISTS workspace_repos (
    webhook_id        UUID REFERENCES webhooks ON UPDATE CASCADE NOT NULL,
    vcs_provider_id   TEXT REFERENCES vcs_providers ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    workspace_id      TEXT REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                      UNIQUE (workspace_id)
);

CREATE TABLE IF NOT EXISTS module_repos (
    webhook_id        UUID REFERENCES webhooks ON UPDATE CASCADE NOT NULL,
    vcs_provider_id   TEXT REFERENCES vcs_providers ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    module_id         TEXT REFERENCES modules ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                      UNIQUE (module_id)
);

INSERT INTO workspace_repos (webhook_id, vcs_provider_id, workspace_id)
SELECT webhook_id, vcs_provider_id, workspace_id
FROM repo_connections;

INSERT INTO module_repos (webhook_id, vcs_provider_id, module_id)
SELECT webhook_id, vcs_provider_id, module_id
FROM repo_connections;

DROP TABLE IF EXISTS repo_connections;
