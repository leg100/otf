-- Replaces unique constraint on variable_sets table, adding the organization
-- name to the constraint.
ALTER TABLE variable_sets DROP CONSTRAINT variable_sets_name_key;
ALTER TABLE variable_sets ADD CONSTRAINT variable_sets_name_key UNIQUE (organization_name,name);
---- create above / drop below ----
ALTER TABLE variable_sets DROP CONSTRAINT variable_sets_name_key;
ALTER TABLE variable_sets ADD CONSTRAINT variable_sets_name_key UNIQUE (name);
