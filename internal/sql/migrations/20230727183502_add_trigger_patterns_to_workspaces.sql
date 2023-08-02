-- +goose Up
ALTER TABLE workspaces
    ADD COLUMN trigger_patterns TEXT[],
    ADD COLUMN vcs_tags_regex TEXT,
    DROP COLUMN file_triggers_enabled;

-- +goose Down
ALTER TABLE workspaces
    DROP COLUMN trigger_patterns,
    DROP COLUMN vcs_tags_regex,
    ADD COLUMN file_triggers_enabled BOOL DEFAULT false NOT NULL;

UPDATE workspaces
SET file_triggers_enabled = true
WHERE cardinality(trigger_prefixes) IS NOT NULL
;
