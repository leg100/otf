-- +goose Up
ALTER TABLE organizations ADD CONSTRAINT org_name_key UNIQUE (name);

-- +goose Down
ALTER TABLE organizations DROP CONSTRAINT IF EXISTS org_name_key;
