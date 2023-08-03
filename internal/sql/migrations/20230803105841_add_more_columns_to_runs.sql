-- +goose Up
ALTER TABLE runs
    ADD COLUMN source TEXT,
    ADD COLUMN terraform_version TEXT,
    ADD COLUMN allow_empty_apply BOOL NOT NULL DEFAULT false;

UPDATE runs SET source = 'tfe-api';

UPDATE runs
SET terraform_version = w.terraform_version
FROM workspaces w
WHERE runs.workspace_id = w.workspace_id;

ALTER TABLE runs
    ALTER COLUMN source SET NOT NULL,
    ALTER COLUMN terraform_version SET NOT NULL;

CREATE TABLE IF NOT EXISTS run_variables (
    run_id TEXT REFERENCES runs ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS run_variables;
ALTER TABLE runs
    DROP COLUMN source,
    DROP COLUMN terraform_version,
    DROP COLUMN allow_empty_apply;
