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

-- add engine column to workspaces to allow explicitly setting an engine on a
-- per workspace basis.
ALTER TABLE workspaces ADD COLUMN engine TEXT REFERENCES engines(name);

---- create above / drop below ----
ALTER TABLE latest_engine_version DROP COLUMN engine;
ALTER TABLE latest_engine_version RENAME TO latest_terraform_version;
ALTER TABLE workspaces DROP COLUMN engine;
DROP TABLE engines;
