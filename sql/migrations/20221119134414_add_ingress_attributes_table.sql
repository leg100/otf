-- +goose Up
CREATE TABLE IF NOT EXISTS ingress_attributes (
    ingress_attribute_id    TEXT,
    branch                  TEXT NOT NULL,
    commit_sha              TEXT NOT NULL,
    identifier              TEXT NOT NULL,
    is_pull_request         BOOL NOT NULL,
    on_default_branch       BOOL NOT NULL,
    organization_id   TEXT REFERENCES organizations ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                      PRIMARY KEY (ingress_attribute_id)
);
ALTER TABLE configuration_versions ADD COLUMN ingress_attribute_id TEXT;
ALTER TABLE configuration_versions ADD CONSTRAINT ingress_attribute_fk FOREIGN_KEY REFERENCES ingress_attributes;

-- +goose Down
ALTER TABLE configuration_versions DROP CONSTRAINT ingress_attribute_fk;
ALTER TABLE configuration_versions DROP COLUMN ingress_attribute_id;
DROP TABLE IF EXISTS ingress_attributes;
