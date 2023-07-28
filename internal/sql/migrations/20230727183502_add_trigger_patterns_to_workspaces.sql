-- +goose Up
ALTER TABLE workspaces
    ADD COLUMN trigger_patterns TEXT[],
    ADD COLUMN vcs_tags_regex TEXT,
    ALTER COLUMN branch DROP NOT NULL;

-- +goose Down
ALTER TABLE workspaces
    DROP COLUMN trigger_patterns,
    DROP COLUMN vcs_tags_regex,
    ALTER COLUMN branch SET NOT NULL;
