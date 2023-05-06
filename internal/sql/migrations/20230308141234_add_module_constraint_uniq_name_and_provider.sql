-- +goose Up
ALTER TABLE modules
	ADD CONSTRAINT modules_org_name_provider_uniq UNIQUE (organization_name, name, provider);

-- +goose Down
ALTER TABLE modules DROP CONSTRAINT modules_org_name_provider_uniq;
