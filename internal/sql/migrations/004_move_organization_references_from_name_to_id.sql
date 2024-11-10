-- move organization fk's from organization_name to organization_id

ALTER TABLE agent_pools ADD COLUMN organization_id TEXT;
ALTER TABLE agent_pools ADD FOREIGN KEY (organization_id) REFERENCES organizations(organization_id);
UPDATE agent_pools SET organization_id = o.organization_id FROM organizations o WHERE agent_pools.organization_name = o.name;
ALTER TABLE agent_pools DROP COLUMN organization_name;
ALTER TABLE agent_pools ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE modules ADD COLUMN organization_id TEXT;
ALTER TABLE modules ADD FOREIGN KEY (organization_id) REFERENCES organizations(organization_id);
UPDATE modules SET organization_id = o.organization_id FROM organizations o WHERE modules.organization_name = o.name;
ALTER TABLE modules DROP COLUMN organization_name;
ALTER TABLE modules ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE organization_tokens ADD COLUMN organization_id TEXT;
ALTER TABLE organization_tokens ADD FOREIGN KEY (organization_id) REFERENCES organizations(organization_id);
UPDATE organization_tokens SET organization_id = o.organization_id FROM organizations o WHERE organization_tokens.organization_name = o.name;
ALTER TABLE organization_tokens DROP COLUMN organization_name;
ALTER TABLE organization_tokens ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE registry_sessions ADD COLUMN organization_id TEXT;
ALTER TABLE registry_sessions ADD FOREIGN KEY (organization_id) REFERENCES organizations(organization_id);
UPDATE registry_sessions SET organization_id = o.organization_id FROM organizations o WHERE registry_sessions.organization_name = o.name;
ALTER TABLE registry_sessions DROP COLUMN organization_name;
ALTER TABLE registry_sessions ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE tags ADD COLUMN organization_id TEXT;
ALTER TABLE tags ADD FOREIGN KEY (organization_id) REFERENCES organizations(organization_id);
UPDATE tags SET organization_id = o.organization_id FROM organizations o WHERE tags.organization_name = o.name;
ALTER TABLE tags DROP COLUMN organization_name;
ALTER TABLE tags ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE teams ADD COLUMN organization_id TEXT;
ALTER TABLE teams ADD FOREIGN KEY (organization_id) REFERENCES organizations(organization_id);
UPDATE teams SET organization_id = o.organization_id FROM organizations o WHERE teams.organization_name = o.name;
ALTER TABLE teams DROP COLUMN organization_name;
ALTER TABLE teams ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE variable_sets ADD COLUMN organization_id TEXT;
ALTER TABLE variable_sets ADD FOREIGN KEY (organization_id) REFERENCES organizations(organization_id);
UPDATE variable_sets SET organization_id = o.organization_id FROM organizations o WHERE variable_sets.organization_name = o.name;
ALTER TABLE variable_sets DROP COLUMN organization_name;
ALTER TABLE variable_sets ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE vcs_providers ADD COLUMN organization_id TEXT;
ALTER TABLE vcs_providers ADD FOREIGN KEY (organization_id) REFERENCES organizations(organization_id);
UPDATE vcs_providers SET organization_id = o.organization_id FROM organizations o WHERE vcs_providers.organization_name = o.name;
ALTER TABLE vcs_providers DROP COLUMN organization_name;
ALTER TABLE vcs_providers ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE workspaces ADD COLUMN organization_id TEXT;
ALTER TABLE workspaces ADD FOREIGN KEY (organization_id) REFERENCES organizations(organization_id);
UPDATE workspaces SET organization_id = o.organization_id FROM organizations o WHERE workspaces.organization_name = o.name;
ALTER TABLE workspaces DROP COLUMN organization_name;
ALTER TABLE workspaces ALTER COLUMN organization_id SET NOT NULL;

---- create above / drop below ----

ALTER TABLE agent_pools ADD COLUMN organization_name TEXT;
ALTER TABLE agent_pools ADD FOREIGN KEY (organization_name) REFERENCES organizations(name);
UPDATE agent_pools SET organization_name = o.name FROM organizations o WHERE agent_pools.organization_id = o.organization_id;
ALTER TABLE agent_pools DROP COLUMN organization_id;
ALTER TABLE agent_pools ALTER COLUMN organization_name SET NOT NULL;

ALTER TABLE modules ADD COLUMN organization_name TEXT;
ALTER TABLE modules ADD FOREIGN KEY (organization_name) REFERENCES organizations(name);
UPDATE modules SET organization_name = o.name FROM organizations o WHERE modules.organization_id = o.organization_id;
ALTER TABLE modules DROP COLUMN organization_id;
ALTER TABLE modules ALTER COLUMN organization_name SET NOT NULL;

ALTER TABLE organization_tokens ADD COLUMN organization_name TEXT;
ALTER TABLE organization_tokens ADD FOREIGN KEY (organization_name) REFERENCES organizations(name);
UPDATE organization_tokens SET organization_name = o.name FROM organizations o WHERE organization_tokens.organization_id = o.organization_id;
ALTER TABLE organization_tokens DROP COLUMN organization_id;
ALTER TABLE organization_tokens ALTER COLUMN organization_name SET NOT NULL;

ALTER TABLE registry_sessions ADD COLUMN organization_name TEXT;
ALTER TABLE registry_sessions ADD FOREIGN KEY (organization_name) REFERENCES organizations(name);
UPDATE registry_sessions SET organization_name = o.name FROM organizations o WHERE registry_sessions.organization_id = o.organization_id;
ALTER TABLE registry_sessions DROP COLUMN organization_id;
ALTER TABLE registry_sessions ALTER COLUMN organization_name SET NOT NULL;

ALTER TABLE tags ADD COLUMN organization_name TEXT;
ALTER TABLE tags ADD FOREIGN KEY (organization_name) REFERENCES organizations(name);
UPDATE tags SET organization_name = o.name FROM organizations o WHERE tags.organization_id = o.organization_id;
ALTER TABLE tags DROP COLUMN organization_id;
ALTER TABLE tags ALTER COLUMN organization_name SET NOT NULL;

ALTER TABLE teams ADD COLUMN organization_name TEXT;
ALTER TABLE teams ADD FOREIGN KEY (organization_name) REFERENCES organizations(name);
UPDATE teams SET organization_name = o.name FROM organizations o WHERE teams.organization_id = o.organization_id;
ALTER TABLE teams DROP COLUMN organization_id;
ALTER TABLE teams ALTER COLUMN organization_name SET NOT NULL;

ALTER TABLE variable_sets ADD COLUMN organization_name TEXT;
ALTER TABLE variable_sets ADD FOREIGN KEY (organization_name) REFERENCES organizations(name);
UPDATE variable_sets SET organization_name = o.name FROM organizations o WHERE variable_sets.organization_id = o.organization_id;
ALTER TABLE variable_sets DROP COLUMN organization_id;
ALTER TABLE variable_sets ALTER COLUMN organization_name SET NOT NULL;

ALTER TABLE vcs_providers ADD COLUMN organization_name TEXT;
ALTER TABLE vcs_providers ADD FOREIGN KEY (organization_name) REFERENCES organizations(name);
UPDATE vcs_providers SET organization_name = o.name FROM organizations o WHERE vcs_providers.organization_id = o.organization_id;
ALTER TABLE vcs_providers DROP COLUMN organization_id;
ALTER TABLE vcs_providers ALTER COLUMN organization_name SET NOT NULL;

ALTER TABLE workspaces ADD COLUMN organization_name TEXT;
ALTER TABLE workspaces ADD FOREIGN KEY (organization_name) REFERENCES organizations(name);
UPDATE workspaces SET organization_name = o.name FROM organizations o WHERE workspaces.organization_id = o.organization_id;
ALTER TABLE workspaces DROP COLUMN organization_id;
ALTER TABLE workspaces ALTER COLUMN organization_name SET NOT NULL;
