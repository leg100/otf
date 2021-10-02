CREATE TABLE IF NOT EXISTS organizations (
    id integer,
    created_at datetime,
    updated_at datetime,
    deleted_at datetime,
    external_id text,
    name text,
    collaborator_auth_policy text,
    cost_estimation_enabled numeric,
    email text,
    owners_team_saml_role_id text,
    permission_can_create_team numeric,
    permission_can_create_workspace numeric,
    permission_can_create_workspace_migration numeric,
    permission_can_destroy numeric,
    permission_can_traverse numeric,
    permission_can_update numeric,
    permission_can_update_api_token numeric,
    permission_can_update_o_auth numeric,
    permission_can_update_sentinel numeric,
    saml_enabled numeric,
    session_remember integer,
    session_timeout integer,
    trial_expires_at datetime,
    two_factor_conformant numeric,
    PRIMARY KEY (id));
CREATE UNIQUE INDEX IF NOT EXISTS idx_organizations_external_id ON organizations(external_id);
CREATE INDEX IF NOT EXISTS idx_organizations_deleted_at ON organizations(deleted_at);

CREATE TABLE IF NOT EXISTS workspaces (
    id integer,
    created_at datetime,
    updated_at datetime,
    deleted_at datetime,
    external_id text,
    allow_destroy_plan numeric,
    auto_apply numeric,
    can_queue_destroy_plan numeric,
    description text,
    environment text,
    execution_mode text,
    file_triggers_enabled numeric,
    global_remote_state numeric,
    locked numeric,
    migration_environment text,
    name text,
    permission_can_destroy numeric,
    permission_can_force_unlock numeric,
    permission_can_lock numeric,
    permission_can_queue_apply numeric,
    permission_can_queue_destroy numeric,
    permission_can_queue_run numeric,
    permission_can_read_settings numeric,
    permission_can_unlock numeric,
    permission_can_update numeric,
    permission_can_update_variable numeric,
    queue_all_runs numeric,
    speculative_enabled numeric,
    source_name text,
    source_url text,
    structured_run_output_enabled numeric,
    terraform_version text,
    trigger_prefixes text,
    working_directory text,
    resource_count integer,
    apply_duration_average integer,
    plan_duration_average integer,
    policy_check_failures integer,
    run_failures integer,
    runs_count integer,
    organization_id integer,
    PRIMARY KEY (id),
    CONSTRAINT fk_workspaces_organization FOREIGN KEY (organization_id) REFERENCES organizations(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_workspaces_external_id ON workspaces(external_id);
CREATE INDEX IF NOT EXISTS idx_workspaces_deleted_at ON workspaces(deleted_at);

CREATE TABLE IF NOT EXISTS configuration_versions (
    id integer,
    created_at datetime,
    updated_at datetime,
    deleted_at datetime,
    external_id text,
    auto_queue_runs numeric,
    error text,
    error_message text,
    source text,
    speculative numeric,
    status text,
    timestamp_finished_at datetime,
    timestamp_queued_at datetime,
    timestamp_started_at datetime,
    blob_id text,
    workspace_id integer,
    PRIMARY KEY (id),
    CONSTRAINT fk_configuration_versions_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_configuration_versions_external_id ON configuration_versions(external_id);
CREATE INDEX IF NOT EXISTS idx_configuration_versions_deleted_at ON configuration_versions(deleted_at);

CREATE TABLE IF NOT EXISTS runs (
    id integer,
    created_at datetime,
    updated_at datetime,
    deleted_at datetime,
    external_id text,
    force_cancel_available_at datetime,
    is_destroy numeric,
    message text,
    permission_can_apply numeric,
    permission_can_cancel numeric,
    permission_can_discard numeric,
    permission_can_force_cancel numeric,
    permission_can_force_execute numeric,
    position_in_queue integer,
    refresh numeric,
    refresh_only numeric,
    status text,
    timestamp_applied_at datetime,
    timestamp_apply_queued_at datetime,
    timestamp_applying_at datetime,
    timestamp_canceled_at datetime,
    timestamp_confirmed_at datetime,
    timestamp_cost_estimated_at datetime,
    timestamp_cost_estimating_at datetime,
    timestamp_discarded_at datetime,
    timestamp_errored_at datetime,
    timestamp_force_canceled_at datetime,
    timestamp_plan_queueable_at datetime,
    timestamp_plan_queued_at datetime,
    timestamp_planned_and_finished_at datetime,
    timestamp_planned_at datetime,
    timestamp_planning_at datetime,
    timestamp_policy_checked_at datetime,
    timestamp_policy_soft_failed_at datetime,
    replace_addrs text,
    target_addrs text,
    workspace_id integer,
    configuration_version_id integer,
    PRIMARY KEY (id),
    CONSTRAINT fk_runs_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id),
    CONSTRAINT fk_runs_configuration_version FOREIGN KEY (configuration_version_id) REFERENCES configuration_versions(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_runs_external_id ON runs(external_id);
CREATE INDEX IF NOT EXISTS idx_runs_deleted_at ON runs(deleted_at);

CREATE TABLE IF NOT EXISTS applies (
    id integer,
    created_at datetime,
    updated_at datetime,
    deleted_at datetime,
    external_id text,
    resource_additions integer,
    resource_changes integer,
    resource_destructions integer,
    status text,
    timestamp_canceled_at datetime,
    timestamp_errored_at datetime,
    timestamp_finished_at datetime,
    timestamp_force_canceled_at datetime,
    timestamp_queued_at datetime,
    timestamp_started_at datetime,
    logs_blob_id text,
    run_id integer,
    PRIMARY KEY (id),
    CONSTRAINT fk_runs_apply FOREIGN KEY (run_id) REFERENCES runs(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_applies_external_id ON applies(external_id);
CREATE INDEX IF NOT EXISTS idx_applies_deleted_at ON applies(deleted_at);

CREATE TABLE IF NOT EXISTS plans (
    id integer,
    created_at datetime,
    updated_at datetime,
    deleted_at datetime,
    external_id text,
    resource_additions integer,
    resource_changes integer,
    resource_destructions integer,
    status text,
    timestamp_canceled_at datetime,
    timestamp_errored_at datetime,
    timestamp_finished_at datetime,
    timestamp_force_canceled_at datetime,
    timestamp_queued_at datetime,
    timestamp_started_at datetime,
    logs_blob_id text,
    plan_file_blob_id text,
    plan_json_blob_id text,
    run_id integer,
    PRIMARY KEY (id),
    CONSTRAINT fk_runs_plan FOREIGN KEY (run_id) REFERENCES runs(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_plans_external_id ON plans(external_id);
CREATE INDEX IF NOT EXISTS idx_plans_deleted_at ON plans(deleted_at);

CREATE TABLE IF NOT EXISTS state_versions (
    id integer,
    created_at datetime,
    updated_at datetime,
    deleted_at datetime,
    external_id text,
    serial integer,
    vcs_commit_sha text,
    vcs_commit_url text,
    blob_id text,
    workspace_id integer,
    PRIMARY KEY (id),
    CONSTRAINT fk_state_versions_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_state_versions_external_id ON state_versions(external_id);
CREATE INDEX IF NOT EXISTS idx_state_versions_deleted_at ON state_versions(deleted_at);

CREATE TABLE IF NOT EXISTS state_version_outputs (
    id integer,
    created_at datetime,
    updated_at datetime,
    deleted_at datetime,
    external_id text,
    name text,
    sensitive numeric,
    type text,
    value text,
    state_version_id integer,
    PRIMARY KEY (id),
    CONSTRAINT fk_state_versions_outputs FOREIGN KEY (state_version_id) REFERENCES state_versions(id)
);
CREATE INDEX IF NOT EXISTS idx_state_version_outputs_deleted_at ON state_version_outputs(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_state_version_outputs_external_id ON state_version_outputs(external_id);
