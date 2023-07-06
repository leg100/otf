-- +goose Up
ALTER TABLE organizations
    ADD COLUMN email TEXT,
    ADD COLUMN collaborator_auth_policy TEXT,
    ADD COLUMN allow_force_delete_workspaces BOOL NOT NULL DEFAULT FALSE,
    ALTER COLUMN session_remember DROP NOT NULL,
    ALTER COLUMN session_timeout DROP NOT NULL;

UPDATE organizations
SET session_remember = NULL, session_timeout = NULL;

-- +goose Down
ALTER TABLE organizations
    DROP COLUMN email,
    DROP COLUMN collaborator_auth_policy,
    DROP COLUMN allow_force_delete_workspaces;

UPDATE organizations
SET session_remember = 0, session_timeout = 0;

ALTER TABLE organizations
    ALTER COLUMN session_remember SET NOT NULL,
    ALTER COLUMN session_timeout SET NOT NULL;
