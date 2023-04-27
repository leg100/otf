-- +goose Up
CREATE TABLE IF NOT EXISTS tags (
    tag_id              TEXT NOT NULL,
    name                TEXT NOT NULL,
    organization_name   TEXT REFERENCES organizations (name) ON UPDATE CASCADE ON DELETE CASCADE,
                        PRIMARY KEY (tag_id),
                        UNIQUE (organization_name, name)
);

CREATE TABLE IF NOT EXISTS workspace_tags (
    tag_id          TEXT REFERENCES tags ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    workspace_id    TEXT REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                    UNIQUE (tag_id, workspace_id)
);

-- +goose Down
DROP TABLE IF EXISTS workspace_tags;
DROP TABLE IF EXISTS tags;
