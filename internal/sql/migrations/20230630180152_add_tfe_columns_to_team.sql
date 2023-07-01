-- +goose Up
ALTER TABLE teams
    ADD COLUMN sso_team_id TEXT,
    ADD COLUMN visibility TEXT NOT NULL DEFAULT 'secret',
    ADD COLUMN permission_manage_policies BOOL NOT NULL DEFAULT FALSE,
    ADD COLUMN permission_manage_policy_overrides BOOL NOT NULL DEFAULT FALSE,
    ADD COLUMN permission_manage_providers BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE teams
    RENAME COLUMN permission_manage_registry TO permission_manage_modules;

-- +goose Down
ALTER TABLE teams
    DROP COLUMN sso_team_id,
    DROP COLUMN visibility,
    DROP COLUMN permission_manage_policies,
    DROP COLUMN permission_manage_policy_overrides,
    DROP COLUMN permission_manage_providers;
ALTER TABLE teams
    RENAME COLUMN permission_manage_modules TO permission_manage_registry;
