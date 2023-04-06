-- +goose Up
ALTER TABLE users ADD COLUMN site_admin BOOL DEFAULT false NOT NULL;

-- +goose Down
ALTER TABLE users DROP COLUMN site_admin;
