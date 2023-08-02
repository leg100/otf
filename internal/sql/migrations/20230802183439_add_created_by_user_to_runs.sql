-- +goose Up
ALTER TABLE runs ADD COLUMN created_by TEXT;

-- +goose Down
ALTER TABLE runs DROP COLUMN created_by;
