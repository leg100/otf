-- +goose Up
ALTER TABLE variables ADD COLUMN version_id TEXT;

UPDATE variables
SET version_id = md5(random()::text) || md5(random()::text);

ALTER TABLE variables ALTER COLUMN version_id SET NOT NULL;

-- +goose Down
ALTER TABLE variables DROP COLUMN version_id;
