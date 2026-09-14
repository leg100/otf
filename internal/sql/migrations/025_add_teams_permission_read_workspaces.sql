-- Add organization-level 'read workspaces' permission to teams, granting
-- read-only access to every workspace in the organization.
--
-- Teams that can manage workspaces are implicitly able to read them, so
-- backfill those teams accordingly.

ALTER TABLE teams
    ADD COLUMN permission_read_workspaces boolean DEFAULT false NOT NULL;

UPDATE teams
SET permission_read_workspaces = true
WHERE permission_manage_workspaces;

---- create above / drop below ----

ALTER TABLE teams
    DROP COLUMN permission_read_workspaces;
