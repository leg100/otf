-- +goose Up
CREATE TABLE IF NOT EXISTS remote_state_consumers (
    workspace_id TEXT REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    consumer_id TEXT REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    UNIQUE (workspace_id, consumer_id),
    CHECK (workspace_id <> consumer_id)
);

-- +goose Down
DROP TABLE IF EXISTS remote_state_consumers;
