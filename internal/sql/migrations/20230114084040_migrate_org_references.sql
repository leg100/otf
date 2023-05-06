--
-- migrate workspaces organization fk from id to name
--
-- +goose Up
ALTER TABLE workspaces
    ADD COLUMN organization_name TEXT,
    ADD CONSTRAINT workspaces_organization_name_fkey
        FOREIGN KEY (organization_name) REFERENCES organizations (name) ON UPDATE CASCADE ON DELETE CASCADE
;
UPDATE workspaces w
SET organization_name = o.name
FROM organizations o
WHERE w.organization_id = o.organization_id
;
ALTER TABLE workspaces
    ALTER COLUMN organization_name SET NOT NULL,
    DROP COLUMN organization_id
;

-- +goose Down
ALTER TABLE workspaces
    ADD COLUMN organization_id TEXT,
    ADD CONSTRAINT workspaces_organization_id_fkey
        FOREIGN KEY (organization_id) REFERENCES organizations ON UPDATE CASCADE ON DELETE CASCADE
;
UPDATE workspaces w
SET organization_id = o.organization_id
FROM organizations o
WHERE w.organization_name = o.name
;
ALTER TABLE workspaces
    ALTER COLUMN organization_id SET NOT NULL,
    DROP COLUMN organization_name
;
