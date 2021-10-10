CREATE TABLE IF NOT EXISTS organizations (
    id integer,
    created_at datetime,
    updated_at datetime,
    external_id text,
    name text,
    email text,
    session_remember integer,
    session_timeout integer,
    PRIMARY KEY (id));
CREATE UNIQUE INDEX IF NOT EXISTS idx_organizations_external_id ON organizations(external_id);

CREATE TABLE IF NOT EXISTS workspaces (
    id integer,
    created_at datetime,
    updated_at datetime,
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
    queue_all_runs numeric,
    speculative_enabled numeric,
    source_name text,
    source_url text,
    terraform_version text,
    trigger_prefixes text,
    working_directory text,
    organization_id integer,
    PRIMARY KEY (id),
    CONSTRAINT fk_workspaces_organization FOREIGN KEY (organization_id) REFERENCES organizations(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_workspaces_external_id ON workspaces(external_id);

CREATE TABLE IF NOT EXISTS configuration_versions (
    id integer,
    created_at datetime,
    updated_at datetime,
    external_id text,
    auto_queue_runs numeric,
    source text,
    speculative numeric,
    status text,
    status_timestamps text,
    blob_id text,
    workspace_id integer,
    PRIMARY KEY (id),
    CONSTRAINT fk_configuration_versions_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_configuration_versions_external_id ON configuration_versions(external_id);

CREATE TABLE IF NOT EXISTS runs (
    id integer,
    created_at datetime,
    updated_at datetime,
    external_id text,
    force_cancel_available_at datetime,
    is_destroy numeric,
    position_in_queue integer,
    refresh numeric,
    refresh_only numeric,
    status text,
    status_timestamps text,
    replace_addrs text,
    target_addrs text,
    workspace_id integer,
    configuration_version_id integer,
    PRIMARY KEY (id),
    CONSTRAINT fk_runs_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id),
    CONSTRAINT fk_runs_configuration_version FOREIGN KEY (configuration_version_id) REFERENCES configuration_versions(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_runs_external_id ON runs(external_id);

CREATE TABLE IF NOT EXISTS run_timestamps (
    id integer,
    status text,
    timestamp datetime,
    PRIMARY KEY (id, status),
    CONSTRAINT fk_run_timestamps_runs FOREIGN KEY (id) REFERENCES runs(id)
);

CREATE TABLE IF NOT EXISTS applies (
    id integer,
    created_at datetime,
    updated_at datetime,
    external_id text,
    resource_additions integer,
    resource_changes integer,
    resource_destructions integer,
    status text,
    status_timestamps text,
    logs_blob_id text,
    run_id integer,
    PRIMARY KEY (id),
    CONSTRAINT fk_runs_apply FOREIGN KEY (run_id) REFERENCES runs(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_applies_external_id ON applies(external_id);

CREATE TABLE IF NOT EXISTS plans (
    id integer,
    created_at datetime,
    updated_at datetime,
    external_id text,
    resource_additions integer,
    resource_changes integer,
    resource_destructions integer,
    status text,
    status_timestamps text,
    logs_blob_id text,
    plan_file_blob_id text,
    plan_json_blob_id text,
    run_id integer,
    PRIMARY KEY (id),
    CONSTRAINT fk_runs_plan FOREIGN KEY (run_id) REFERENCES runs(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_plans_external_id ON plans(external_id);

CREATE TABLE IF NOT EXISTS state_versions (
    id integer,
    created_at datetime,
    updated_at datetime,
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

CREATE TABLE IF NOT EXISTS state_version_outputs (
    id integer,
    created_at datetime,
    updated_at datetime,
    external_id text,
    name text,
    sensitive numeric,
    type text,
    value text,
    state_version_id integer,
    PRIMARY KEY (id),
    CONSTRAINT fk_state_versions_outputs FOREIGN KEY (state_version_id) REFERENCES state_versions(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_state_version_outputs_external_id ON state_version_outputs(external_id);
