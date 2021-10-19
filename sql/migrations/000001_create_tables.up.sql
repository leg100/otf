-- +goose Up
CREATE TABLE IF NOT EXISTS organizations (
    organization_id text,
    created_at timestamptz,
    updated_at timestamptz,
    name text,
    email text,
    session_remember integer,
    session_timeout integer,
    PRIMARY KEY (organization_id)
);

CREATE TABLE IF NOT EXISTS workspaces (
    workspace_id text,
    created_at timestamptz,
    updated_at timestamptz,
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
    organization_id text REFERENCES organizations ON UPDATE CASCADE ON DELETE CASCADE,
    PRIMARY KEY (workspace_id)
);

CREATE TABLE IF NOT EXISTS configuration_versions (
    configuration_version_id text,
    created_at timestamptz,
    updated_at timestamptz,
    auto_queue_runs boolean,
    source text,
    speculative boolean,
    status text,
    status_timestamps text,
    blob_id text,
    workspace_id text REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE,
    PRIMARY KEY (configuration_version_id)
);

CREATE TABLE IF NOT EXISTS runs (
    run_id text,
    created_at timestamptz,
    updated_at timestamptz,
    is_destroy boolean,
    position_in_queue integer,
    refresh boolean,
    refresh_only boolean,
    status text,
    status_timestamps text,
    replace_addrs text,
    target_addrs text,
    workspace_id text REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE,
    configuration_version_id text REFERENCES configuration_versions ON UPDATE CASCADE ON DELETE CASCADE,
    PRIMARY KEY (run_id)
);

CREATE TABLE IF NOT EXISTS applies (
    apply_id text,
    created_at timestamptz,
    updated_at timestamptz,
    resource_additions integer,
    resource_changes integer,
    resource_destructions integer,
    status text,
    status_timestamps text,
    logs_blob_id text,
    run_id text REFERENCES runs ON UPDATE CASCADE ON DELETE CASCADE,
    PRIMARY KEY (apply_id)
);

CREATE TABLE IF NOT EXISTS plans (
    plan_id text,
    created_at timestamptz,
    updated_at timestamptz,
    resource_additions integer,
    resource_changes integer,
    resource_destructions integer,
    status text,
    status_timestamps text,
    logs_blob_id text,
    plan_file_blob_id text,
    plan_json_blob_id text,
    run_id text REFERENCES runs ON UPDATE CASCADE ON DELETE CASCADE,
    PRIMARY KEY (plan_id)
);

CREATE TABLE IF NOT EXISTS state_versions (
    state_version_id text,
    created_at timestamptz,
    updated_at timestamptz,
    serial integer,
    vcs_commit_sha text,
    vcs_commit_url text,
    blob_id text,
    workspace_id text REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE,
    PRIMARY KEY (state_version_id)
);

CREATE TABLE IF NOT EXISTS state_version_outputs (
    state_version_output_id text,
    created_at timestamptz,
    updated_at timestamptz,
    name text,
    sensitive boolean,
    type text,
    value text,
    state_version_id text REFERENCES state_versions ON UPDATE CASCADE ON DELETE CASCADE,
    PRIMARY KEY (state_version_output_id)
);
-- +goose Down
DROP TABLE IF EXISTS applies;
DROP TABLE IF EXISTS plans;
DROP TABLE IF EXISTS runs;
DROP TABLE IF EXISTS state_version_outputs;
DROP TABLE IF EXISTS state_versions;
DROP TABLE IF EXISTS configuration_versions;
DROP TABLE IF EXISTS workspaces;
DROP TABLE IF EXISTS organizations;
