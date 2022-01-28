-- +goose Up
CREATE TABLE IF NOT EXISTS organizations (
    organization_id TEXT,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    name TEXT,
    session_remember INTEGER,
    session_timeout INTEGER,
    UNIQUE (name),
    PRIMARY KEY (organization_id)
);

CREATE TABLE IF NOT EXISTS workspaces (
    workspace_id TEXT,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    allow_destroy_plan BOOLEAN,
    auto_apply BOOLEAN,
    can_queue_destroy_plan BOOLEAN,
    description TEXT,
    environment TEXT,
    execution_mode TEXT,
    file_triggers_enabled BOOLEAN,
    global_remote_state BOOLEAN,
    locked BOOLEAN,
    migration_environment TEXT,
    name TEXT,
    queue_all_runs BOOLEAN,
    speculative_enabled BOOLEAN,
    source_name TEXT,
    source_url TEXT,
    structured_run_output_enabled BOOLEAN,
    terraform_version TEXT,
    trigger_prefixes TEXT,
    working_directory TEXT,
    organization_id TEXT REFERENCES organizations ON UPDATE CASCADE ON DELETE CASCADE,
    UNIQUE (organization_id, name),
    PRIMARY KEY (workspace_id)
);

CREATE TABLE IF NOT EXISTS users (
    user_id TEXT,
    username TEXT NOT NULL,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    PRIMARY KEY (user_id)
);

CREATE TABLE IF NOT EXISTS organization_memberships (
    user_id text REFERENCES users ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    organization_id text REFERENCES organizations ON UPDATE CASCADE ON DELETE CASCADE NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    token TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    address TEXT NOT NULL,
    flash JSONB,
    organization TEXT,
    expiry TIMESTAMPTZ NOT NULL,
    user_id TEXT REFERENCES users ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    PRIMARY KEY (token)
);

CREATE TABLE IF NOT EXISTS configuration_versions (
    configuration_version_id TEXT,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    auto_queue_runs BOOLEAN,
    source TEXT,
    speculative BOOLEAN,
    status TEXT,
    status_timestamps TEXT,
    config BYTEA,
    workspace_id TEXT REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE,
    PRIMARY KEY (configuration_version_id)
);

CREATE TABLE IF NOT EXISTS runs (
    run_id TEXT,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    is_destroy BOOLEAN,
    position_in_queue INTEGER,
    refresh BOOLEAN,
    refresh_only BOOLEAN,
    status TEXT,
    status_timestamps TEXT,
    replace_addrs TEXT,
    target_addrs TEXT,
    workspace_id TEXT REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE,
    configuration_version_id TEXT REFERENCES configuration_versions ON UPDATE CASCADE ON DELETE CASCADE,
    PRIMARY KEY (run_id)
);

CREATE TABLE IF NOT EXISTS applies (
    apply_id TEXT,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    resource_additions INTEGER,
    resource_changes INTEGER,
    resource_destructions INTEGER,
    status TEXT,
    status_timestamps TEXT,
    run_id TEXT REFERENCES runs ON UPDATE CASCADE ON DELETE CASCADE,
    PRIMARY KEY (apply_id)
);

CREATE TABLE IF NOT EXISTS plans (
    plan_id TEXT,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    resource_additions INTEGER,
    resource_changes INTEGER,
    resource_destructions INTEGER,
    status TEXT,
    status_timestamps TEXT,
    plan_file BYTEA,
    plan_json BYTEA,
    run_id TEXT REFERENCES runs ON UPDATE CASCADE ON DELETE CASCADE,
    PRIMARY KEY (plan_id)
);

CREATE TABLE IF NOT EXISTS state_versions (
    state_version_id TEXT,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    serial INTEGER,
    vcs_commit_sha TEXT,
    vcs_commit_url TEXT,
    state BYTEA,
    workspace_id TEXT REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE,
    PRIMARY KEY (state_version_id)
);

CREATE TABLE IF NOT EXISTS state_version_outputs (
    state_version_output_id TEXT,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    name TEXT,
    sensitive BOOLEAN,
    type TEXT,
    value TEXT,
    state_version_id TEXT REFERENCES state_versions ON UPDATE CASCADE ON DELETE CASCADE,
    PRIMARY KEY (state_version_output_id)
);

CREATE TABLE IF NOT EXISTS plan_logs (
    plan_id TEXT REFERENCES plans ON UPDATE CASCADE ON DELETE CASCADE,
    chunk_id SERIAL,
    chunk BYTEA,
    size INTEGER,
    start BOOLEAN,
    _end BOOLEAN,
    PRIMARY KEY (plan_id, chunk_id)
);

CREATE TABLE IF NOT EXISTS apply_logs (
    apply_id TEXT REFERENCES applies ON UPDATE CASCADE ON DELETE CASCADE,
    chunk_id SERIAL,
    chunk BYTEA,
    size INTEGER,
    start BOOLEAN,
    _end BOOLEAN,
    PRIMARY KEY (apply_id, chunk_id)
);

-- +goose Down
DROP TABLE IF EXISTS apply_logs;
DROP TABLE IF EXISTS plan_logs;
DROP TABLE IF EXISTS state_version_outputs;
DROP TABLE IF EXISTS state_versions;
DROP TABLE IF EXISTS plans;
DROP TABLE IF EXISTS applies;
DROP TABLE IF EXISTS runs;
DROP TABLE IF EXISTS configuration_versions;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS organization_memberships;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS workspaces;
DROP TABLE IF EXISTS organizations;
