-- add support for tofu; terraform is not the only game in town now, so
-- introduce a table of engines: terraform and tofu
CREATE TABLE engines (
    name text PRIMARY KEY
);
INSERT INTO engines (name) VALUES ('terraform'), ('tofu');

-- rename the table of latest_terraform_version to latest_engine_version; add a
-- column to distinguish the engine type, and set any existing rows (should
-- only be one) to the default, terraform.
ALTER TABLE latest_terraform_version RENAME TO latest_engine_version;
ALTER TABLE latest_engine_version ADD COLUMN engine TEXT REFERENCES engines(name);
UPDATE latest_engine_version SET engine = 'terraform';
ALTER TABLE latest_engine_version ALTER COLUMN engine SET NOT NULL;
ALTER TABLE latest_engine_version ADD CONSTRAINT latest_engine_version_engine_key UNIQUE (engine);

-- add engine column to workspaces and set existing workspaces to use terraform.
ALTER TABLE workspaces ADD COLUMN engine TEXT REFERENCES engines(name);
UPDATE workspaces SET engine = 'terraform';
ALTER TABLE workspaces ALTER COLUMN engine SET NOT NULL;

-- add engine column to runs and set existing runs to use terraform.
ALTER TABLE runs ADD COLUMN engine TEXT REFERENCES engines(name);
UPDATE runs SET engine = 'terraform';
ALTER TABLE runs ALTER COLUMN engine SET NOT NULL;

-- rename terraform_version column on workspaces table to engine_version
ALTER TABLE workspaces RENAME COLUMN terraform_version to engine_version;

-- rename terraform_version column on runs table to engine_version
ALTER TABLE runs RENAME COLUMN terraform_version to engine_version;

---- create above / drop below ----
ALTER TABLE runs RENAME COLUMN engine_version to terraform_version;
ALTER TABLE workspaces RENAME COLUMN engine_version to terraform_version;
ALTER TABLE latest_engine_version DROP COLUMN engine;
ALTER TABLE latest_engine_version RENAME TO latest_terraform_version;
-- delete all rows in latest_terraform_version table, and let the latest
-- version checker re-populate it; if we don't drop it then if the user
-- re-migrates the schema then the unique engine constraint above will produce
-- an error
DELETE FROM latest_terraform_version;
ALTER TABLE workspaces DROP COLUMN engine;
ALTER TABLE runs DROP COLUMN engine;
DROP TABLE engines;
