-- +goose Up
CREATE TABLE IF NOT EXISTS latest_terraform_version (
    version TEXT NOT NULL,
    checkpoint TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS latest_terraform_version;
