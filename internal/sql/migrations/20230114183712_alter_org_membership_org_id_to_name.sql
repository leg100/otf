--
-- migrate organization membership's organization fk from id to name
--
-- +goose Up
ALTER TABLE organization_memberships
    ADD COLUMN organization_name TEXT,
    ADD CONSTRAINT organization_memberships_org_name_fkey
        FOREIGN KEY (organization_name) REFERENCES organizations (name) ON UPDATE CASCADE ON DELETE CASCADE
;
UPDATE organization_memberships om
SET organization_name = o.name
FROM organizations o
WHERE om.organization_id = o.organization_id
;
ALTER TABLE organization_memberships
    ALTER COLUMN organization_name SET NOT NULL,
    DROP COLUMN organization_id
;

-- +goose Down
ALTER TABLE organization_memberships
    ADD COLUMN organization_id TEXT,
    ADD CONSTRAINT organization_memberships_organization_id_fkey
        FOREIGN KEY (organization_id) REFERENCES organizations ON UPDATE CASCADE ON DELETE CASCADE
;
UPDATE organization_memberships om
SET organization_id = o.organization_id
FROM organizations o
WHERE om.organization_name = o.name
;
ALTER TABLE organization_memberships
    ALTER COLUMN organization_id SET NOT NULL,
    DROP COLUMN organization_name
;
