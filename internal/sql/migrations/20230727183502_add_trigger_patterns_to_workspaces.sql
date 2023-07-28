-- +goose Up
ALTER TABLE workspaces
    ADD COLUMN trigger_patterns TEXT[],
    ADD COLUMN vcs_tags_regex TEXT;

-- +goose Down
ALTER TABLE workspaces
    DROP COLUMN trigger_patterns,
    DROP COLUMN vcs_tags_regex;
