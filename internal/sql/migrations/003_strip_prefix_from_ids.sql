UPDATE organizations SET organization_id = trim(leading 'organization-' from organization_id);
UPDATE workspaces SET workspace_id = trim(leading 'ws-' from workspace_id);
UPDATE runs SET run_id = trim(leading 'run-' from run_id);
UPDATE state_versions SET state_version_id = trim(leading 'sv-' from state_version_id);
UPDATE state_version_outputs SET state_version_output_id = trim(leading 'svo-' from state_version_output_id);
UPDATE configuration_versions SET configuration_version_id = trim(leading 'cv-' from configuration_version_id);
UPDATE runners SET runner_id = trim(leading 'runner-' from runner_id);

UPDATE runs SET run_id = trim(leading 'run-' from run_id);
UPDATE runs SET run_id = trim(leading 'run-' from run_id);
UPDATE runs SET run_id = trim(leading 'run-' from run_id);
UPDATE runs SET run_id = trim(leading 'run-' from run_id);

---- create above / drop below ----

UPDATE organizations SET organization_id = 'organization-' || organization_id;
UPDATE workspaces SET workspace_id = 'ws-' || workspace_id;
UPDATE runs SET run_id = 'run-' || run_id;
UPDATE state_versions SET state_version_id = 'sv-' || state_version_id;
UPDATE state_version_outputs SET state_version_output_id = 'svo-' || state_version_output_id;
UPDATE configuration_version_id SET configuration_version_id = 'cv-' || configuration_version_id;
UPDATE runners SET runner_id = 'runner-' || runner_id;
