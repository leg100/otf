-- add support for tofu; terraform is not the only game in town now, so rename
-- the table of latest_terraform_version to latest_engine_version; add a
-- column to distinguish the engine type, and set any existing rows (should
-- only be one) to the default, terraform.
ALTER TABLE latest_terraform_version RENAME TO latest_engine_version;
ALTER TABLE latest_engine_version ADD COLUMN engine TEXT;
UPDATE latest_engine_version SET engine = 'terraform';
ALTER TABLE latest_engine_version ALTER COLUMN engine SET NOT NULL;

---- create above / drop below ----
ALTER TABLE latest_engine_version DROP COLUMN engine;
ALTER TABLE latest_engine_version RENAME TO latest_terraform_version;
