-- +goose Up
ALTER TABLE vcs_providers ALTER name DROP NOT NULL;

-- +goose Down
ALTER TABLE vcs_providers ALTER name SET NOT NULL;
