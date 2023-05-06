-- +goose Up
ALTER TABLE runs ADD COLUMN auto_apply BOOL;
UPDATE runs SET auto_apply = false;
ALTER TABLE runs ALTER COLUMN auto_apply SET NOT NULL;

-- +goose Down
ALTER TABLE runs DROP COLUMN auto_apply;
