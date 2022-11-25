-- +goose Up
CREATE TABLE IF NOT EXISTS webhooks (
    webhook_id      UUID,
    vcs_id          TEXT,
    endpoint        TEXT NOT NULL,
    secret          TEXT NOT NULL,
    identifier      TEXT NOT NULL,
    http_url        TEXT NOT NULL,
                    PRIMARY KEY (webhook_id),
                    UNIQUE (http_url)
);

CREATE TABLE IF NOT EXISTS workspace_repos (
    branch            TEXT NOT NULL,
    -- do not cascade deletes because the otfd code relies on getting an error
    -- when attempting to delete a webhook, to determine whether there are any
    -- workspace repos referencing it; only when no more workspace repos are referencing
    -- a webhook do we delete it.
    webhook_id        UUID REFERENCES webhooks ON UPDATE CASCADE NOT NULL,
    vcs_provider_id   TEXT REFERENCES vcs_providers ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    workspace_id      TEXT REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                      UNIQUE (workspace_id)
);

-- +goose Down
DROP TABLE IF EXISTS workspace_repos;
DROP TABLE IF EXISTS webhooks;
