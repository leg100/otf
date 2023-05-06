-- +goose Up
CREATE TABLE IF NOT EXISTS ingress_attributes (
    branch                      TEXT NOT NULL,
    commit_sha                  TEXT NOT NULL,
    identifier                  TEXT NOT NULL,
    is_pull_request             BOOL NOT NULL,
    on_default_branch           BOOL NOT NULL,
    configuration_version_id    TEXT REFERENCES configuration_versions ON UPDATE CASCADE ON DELETE CASCADE NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS ingress_attributes;
