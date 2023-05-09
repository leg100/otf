-- +goose Up
CREATE TABLE IF NOT EXISTS destination_types (
    name TEXT PRIMARY KEY
);
INSERT INTO destination_types (name) VALUES
    ('generic'),
    ('gcppubsub')
;
CREATE TABLE IF NOT EXISTS notification_configurations (
    notification_configuration_id  TEXT,
    created_at       TIMESTAMPTZ NOT NULL,
    updated_at       TIMESTAMPTZ NOT NULL,
    name             TEXT        NOT NULL,
    url              TEXT,
    triggers         TEXT[],
    destination_type TEXT REFERENCES destination_types (name) ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    workspace_id     TEXT REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    enabled          BOOLEAN     NOT NULL,
                     PRIMARY KEY (notification_configuration_id)
);

-- +goose Down
DROP TABLE IF EXISTS notification_configurations;
DROP TABLE IF EXISTS destination_types;
