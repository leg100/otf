-- +goose Up
CREATE TABLE IF NOT EXISTS vcs_repos (
    identifier        TEXT NOT NULL,
    branch            TEXT NOT NULL,
    vcs_provider_id   TEXT REFERENCES vcs_providers ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    workspace_id      TEXT REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                      UNIQUE (workspace_id)
);

-- +goose Down
DROP TABLE IF EXISTS vcs_repos;
