-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS organizations (
    id serial,
    created_at timestamptz,
    updated_at timestamptz,
    external_id text,
    name text,
    email text,
    session_remember integer,
    session_timeout integer,
    PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_organizations_external_id ON organizations(external_id);

CREATE TABLE IF NOT EXISTS workspaces (
    id serial,
    created_at timestamptz,
    updated_at timestamptz,
    external_id text,
    allow_destroy_plan boolean,
    auto_apply boolean,
    can_queue_destroy_plan boolean,
    description text,
    environment text,
    execution_mode text,
    file_triggers_enabled boolean,
    global_remote_state boolean,
    locked boolean,
    migration_environment text,
    name text,
    queue_all_runs boolean,
    speculative_enabled boolean,
    source_name text,
    source_url text,
    structured_run_output_enabled boolean,
    terraform_version text,
    trigger_prefixes text,
    working_directory text,
    organization_id serial,
    PRIMARY KEY (id),
    CONSTRAINT fk_workspaces_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_workspaces_external_id ON workspaces(external_id);

CREATE TABLE IF NOT EXISTS configuration_versions (
    id serial,
    created_at timestamptz,
    updated_at timestamptz,
    external_id text,
    auto_queue_runs boolean,
    source text,
    speculative boolean,
    status text,
    status_timestamps text,
    blob_id text,
    workspace_id serial,
    PRIMARY KEY (id),
    CONSTRAINT fk_configuration_versions_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON UPDATE CASCADE ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_configuration_versions_external_id ON configuration_versions(external_id);

CREATE TABLE IF NOT EXISTS runs (
    id serial,
    created_at timestamptz,
    updated_at timestamptz,
    external_id text,
    is_destroy boolean,
    position_in_queue integer,
    refresh boolean,
    refresh_only boolean,
    status text,
    status_timestamps text,
    replace_addrs text,
    target_addrs text,
    workspace_id serial,
    configuration_version_id serial,
    PRIMARY KEY (id),
    CONSTRAINT fk_runs_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_runs_configuration_version FOREIGN KEY (configuration_version_id) REFERENCES configuration_versions(id) ON UPDATE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_runs_external_id ON runs(external_id);

CREATE TABLE IF NOT EXISTS applies (
    id serial,
    created_at timestamptz,
    updated_at timestamptz,
    external_id text,
    resource_additions integer,
    resource_changes integer,
    resource_destructions integer,
    status text,
    status_timestamps text,
    logs_blob_id text,
    run_id serial,
    PRIMARY KEY (id),
    CONSTRAINT fk_runs_apply FOREIGN KEY (run_id) REFERENCES runs(id) ON UPDATE CASCADE ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_applies_external_id ON applies(external_id);

CREATE TABLE IF NOT EXISTS plans (
    id serial,
    created_at timestamptz,
    updated_at timestamptz,
    external_id text,
    resource_additions integer,
    resource_changes integer,
    resource_destructions integer,
    status text,
    status_timestamps text,
    logs_blob_id text,
    plan_file_blob_id text,
    plan_json_blob_id text,
    run_id serial,
    PRIMARY KEY (id),
    CONSTRAINT fk_runs_plan FOREIGN KEY (run_id) REFERENCES runs(id) ON UPDATE CASCADE ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_plans_external_id ON plans(external_id);

CREATE TABLE IF NOT EXISTS state_versions (
    id serial,
    created_at timestamptz,
    updated_at timestamptz,
    external_id text,
    serial integer,
    vcs_commit_sha text,
    vcs_commit_url text,
    blob_id text,
    workspace_id serial,
    PRIMARY KEY (id),
    CONSTRAINT fk_state_versions_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON UPDATE CASCADE ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_state_versions_external_id ON state_versions(external_id);

CREATE TABLE IF NOT EXISTS state_version_outputs (
    id serial,
    created_at timestamptz,
    updated_at timestamptz,
    external_id text,
    name text,
    sensitive boolean,
    type text,
    value text,
    state_version_id serial,
    PRIMARY KEY (id),
    CONSTRAINT fk_state_versions_outputs FOREIGN KEY (state_version_id) REFERENCES state_versions(id) ON UPDATE CASCADE ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_state_version_outputs_external_id ON state_version_outputs(external_id);
-- +goose StatementEnd
