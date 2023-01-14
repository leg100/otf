--
-- migrate teams organization fk from id to name
--
-- +goose Up
ALTER TABLE teams
    ADD COLUMN organization_name TEXT,
    ADD CONSTRAINT teams_organization_name_fkey
        FOREIGN KEY (organization_name) REFERENCES organizations (name) ON UPDATE CASCADE ON DELETE CASCADE
;
UPDATE teams t
SET organization_name = o.name
FROM organizations o
WHERE t.organization_id = o.organization_id
;
ALTER TABLE teams
    ALTER COLUMN organization_name SET NOT NULL,
    DROP COLUMN organization_id
;

-- +goose Down
ALTER TABLE teams
    ADD COLUMN organization_id TEXT,
    ADD CONSTRAINT teams_organization_id_fkey
        FOREIGN KEY (organization_id) REFERENCES organizations ON UPDATE CASCADE ON DELETE CASCADE
;
UPDATE teams t
SET organization_id = o.organization_id
FROM organizations o
WHERE t.organization_name = o.name
;
ALTER TABLE teams
    ALTER COLUMN organization_id SET NOT NULL,
    DROP COLUMN organization_name
;
