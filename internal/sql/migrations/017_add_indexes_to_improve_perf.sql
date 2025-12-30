CREATE INDEX IF NOT EXISTS idx_runs_created_at ON runs(created_at);
CREATE INDEX IF NOT EXISTS idx_runs_configuration_version_id ON runs(configuration_version_id);
CREATE INDEX IF NOT EXISTS idx_runs_workspace_id ON runs(workspace_id);
CREATE INDEX IF NOT EXISTS idx_runs_status ON runs(status);
CREATE INDEX IF NOT EXISTS idx_workspace_organization_name ON workspaces(organization_name);
CREATE INDEX IF NOT EXISTS idx_configuration_version_workspace_id ON configuration_versions(workspace_id);
CREATE INDEX IF NOT EXISTS idx_ingress_attributes_configuration_version_id ON ingress_attributes(configuration_version_id);
CREATE INDEX IF NOT EXISTS idx_state_versions_workspace_id ON state_versions(workspace_id);
CREATE INDEX IF NOT EXISTS idx_state_versions_created_at ON state_versions(created_at);
---- create above / drop below ----
DROP INDEX idx_runs_created_at;
DROP INDEX idx_runs_configuration_version_id;
DROP INDEX idx_runs_workspace_id;
DROP INDEX idx_runs_status;
DROP INDEX idx_workspace_organization_name;
DROP INDEX idx_configuration_version_workspace_id;
DROP INDEX idx_ingress_attributes_configuration_version_id;
DROP INDEX idx_state_versions_workspace_id;
DROP INDEX idx_state_versions_created_at;
