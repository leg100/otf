-- +goose Up
ALTER TABLE state_versions ADD CONSTRAINT state_version_serial_uniq UNIQUE (workspace_id, serial);

-- +goose Down
ALTER TABLE state_versions DROP CONSTRAINT state_version_serial_uniq;
