-- +goose Up
ALTER TABLE runs ADD COLUMN source TEXT DEFAULT 'tfe-api' NOT NULL;


-- +goose Down
ALTER TABLE runs DROP COLUMN source;
