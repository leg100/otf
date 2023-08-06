-- +goose Up
UPDATE runs SET source = '';

-- +goose Down
UPDATE runs SET source = 'tfe-api';
