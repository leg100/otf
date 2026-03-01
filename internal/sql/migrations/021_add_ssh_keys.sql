CREATE TABLE ssh_keys (
    ssh_key_id        TEXT PRIMARY KEY,
    created_at        TIMESTAMPTZ NOT NULL,
    updated_at        TIMESTAMPTZ NOT NULL,
    name              TEXT NOT NULL,
    private_key       TEXT NOT NULL,
    organization_name TEXT NOT NULL REFERENCES organizations(name) ON UPDATE CASCADE ON DELETE CASCADE
);

ALTER TABLE workspaces
    ADD COLUMN ssh_key_id TEXT REFERENCES ssh_keys(ssh_key_id) ON DELETE SET NULL;

---- create above / drop below ----

ALTER TABLE workspaces DROP COLUMN ssh_key_id;
DROP TABLE ssh_keys;
