-- +goose Up
ALTER TABLE runs ADD COLUMN source TEXT;

UPDATE runs SET source = 'tfe-api';

ALTER TABLE runs ALTER COLUMN source SET NOT NULL;

-- +goose Down
ALTER TABLE runs DROP COLUMN source;
