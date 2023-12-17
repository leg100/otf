-- migrate away from goose for db migrations; now using tern
DROP TABLE IF EXISTS goose_db_version;
DROP SEQUENCE IF EXISTS goose_db_version_id_seq;

-- belately dropping registry_sessions table, which should have been dropped
-- when job tokens were introduced
DROP TABLE IF EXISTS registry_sessions;

CREATE TABLE IF NOT EXISTS organizations (
    organization_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    name text NOT NULL,
    session_remember integer,
    session_timeout integer,
    email text,
    collaborator_auth_policy text,
    allow_force_delete_workspaces boolean DEFAULT false NOT NULL,
    cost_estimation_enabled boolean DEFAULT false NOT NULL,
    PRIMARY KEY (organization_id),
    UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS organization_tokens (
    organization_token_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    organization_name text NOT NULL,
    expiry timestamp with time zone,
    PRIMARY KEY (organization_token_id),
    UNIQUE (organization_name),
    CONSTRAINT organization_tokens_organization_name_fkey FOREIGN KEY (organization_name) REFERENCES organizations(name) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS tags (
    tag_id text NOT NULL,
    name text NOT NULL,
    organization_name text NOT NULL,
    CONSTRAINT tags_pkey PRIMARY KEY (tag_id),
    CONSTRAINT tags_organization_name_name_key UNIQUE (organization_name, name),
    CONSTRAINT tags_organization_name_fkey FOREIGN KEY (organization_name) REFERENCES organizations(name) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS users (
    user_id text NOT NULL,
    username text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    site_admin boolean DEFAULT false NOT NULL,
    CONSTRAINT users_pkey PRIMARY KEY (user_id),
    CONSTRAINT users_username_key UNIQUE (username)
);

INSERT INTO users (
    user_id,
    username,
    created_at,
    updated_at
) VALUES (
    'user-site-admin',
    'site-admin',
    now(),
    now()
)
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS tokens (
    token_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    description text NOT NULL,
    username text NOT NULL,
    CONSTRAINT tokens_pkey PRIMARY KEY (token_id),
    CONSTRAINT token_username_fk FOREIGN KEY (username) REFERENCES users(username) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS sessions (
    token text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    address text NOT NULL,
    expiry timestamp with time zone NOT NULL,
    username text NOT NULL,
    CONSTRAINT sessions_pkey PRIMARY KEY (token),
    CONSTRAINT session_username_fk FOREIGN KEY (username) REFERENCES users(username) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS teams (
    team_id text NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    permission_manage_workspaces boolean DEFAULT false NOT NULL,
    permission_manage_vcs boolean DEFAULT false NOT NULL,
    permission_manage_modules boolean DEFAULT false NOT NULL,
    organization_name text NOT NULL,
    sso_team_id text,
    visibility text DEFAULT 'secret'::text NOT NULL,
    permission_manage_policies boolean DEFAULT false NOT NULL,
    permission_manage_policy_overrides boolean DEFAULT false NOT NULL,
    permission_manage_providers boolean DEFAULT false NOT NULL,
    CONSTRAINT teams_pkey PRIMARY KEY (team_id),
    CONSTRAINT team_name_uniq UNIQUE (organization_name, name),
    CONSTRAINT teams_organization_name_fkey FOREIGN KEY (organization_name) REFERENCES organizations(name) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS team_memberships (
    team_id text NOT NULL,
    username text NOT NULL,
    CONSTRAINT team_member_username_fk FOREIGN KEY (username) REFERENCES users(username) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT team_memberships_team_id_fkey FOREIGN KEY (team_id) REFERENCES teams(team_id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS team_tokens (
    team_token_id text NOT NULL,
    description text,
    created_at timestamp with time zone NOT NULL,
    team_id text NOT NULL,
    expiry timestamp with time zone,
    CONSTRAINT team_tokens_pkey PRIMARY KEY (team_token_id),
    CONSTRAINT team_tokens_team_id_key UNIQUE (team_id),
    CONSTRAINT team_tokens_team_id_fkey FOREIGN KEY (team_id) REFERENCES teams(team_id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS agent_pools (
    agent_pool_id text NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    organization_name text NOT NULL,
    organization_scoped boolean NOT NULL,
    PRIMARY KEY (agent_pool_id),
    UNIQUE (organization_name, name),
    CONSTRAINT agent_pools_organization_name_fkey FOREIGN KEY (organization_name) REFERENCES organizations(name) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS agent_statuses (
    status text NOT NULL,
    PRIMARY KEY (status)
);

INSERT INTO agent_statuses (status) VALUES
    ('busy'),
    ('idle'),
    ('exited'),
    ('errored'),
    ('unknown')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS agent_tokens (
    agent_token_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    description text NOT NULL,
    agent_pool_id text NOT NULL,
    PRIMARY KEY (agent_token_id),
    CONSTRAINT agent_pool_id_fk FOREIGN KEY (agent_pool_id) REFERENCES agent_pools(agent_pool_id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS agents (
    agent_id text NOT NULL,
    name text NOT NULL,
    version text NOT NULL,
    max_jobs integer NOT NULL,
    ip_address inet NOT NULL,
    last_ping_at timestamp with time zone NOT NULL,
    last_status_at timestamp with time zone NOT NULL,
    status text NOT NULL,
    agent_pool_id text,
    PRIMARY KEY (agent_id),
    CONSTRAINT agents_agent_pool_id_fkey FOREIGN KEY (agent_pool_id) REFERENCES agent_pools(agent_pool_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT agents_status_fkey FOREIGN KEY (status) REFERENCES agent_statuses(status) ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS latest_terraform_version (
    version text NOT NULL,
    checkpoint timestamp with time zone NOT NULL
);

CREATE TABLE IF NOT EXISTS workspaces (
    workspace_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    allow_destroy_plan boolean NOT NULL,
    auto_apply boolean NOT NULL,
    can_queue_destroy_plan boolean NOT NULL,
    description text NOT NULL,
    environment text NOT NULL,
    execution_mode text NOT NULL,
    global_remote_state boolean NOT NULL,
    migration_environment text NOT NULL,
    name text NOT NULL,
    queue_all_runs boolean NOT NULL,
    speculative_enabled boolean NOT NULL,
    source_name text NOT NULL,
    source_url text NOT NULL,
    structured_run_output_enabled boolean NOT NULL,
    terraform_version text NOT NULL,
    trigger_prefixes text[],
    working_directory text NOT NULL,
    lock_run_id text,
    latest_run_id text,
    organization_name text NOT NULL,
    branch text NOT NULL,
    lock_username text,
    current_state_version_id text,
    trigger_patterns text[],
    vcs_tags_regex text,
    allow_cli_apply boolean DEFAULT false NOT NULL,
    agent_pool_id text,
    CONSTRAINT workspaces_pkey PRIMARY KEY (workspace_id),
    CONSTRAINT workspace_name_uniq UNIQUE (organization_name, name),
    CONSTRAINT agent_pool_chk CHECK (((execution_mode <> 'agent'::text) OR (agent_pool_id IS NOT NULL))),
    CONSTRAINT agent_pool_fk FOREIGN KEY (agent_pool_id) REFERENCES agent_pools(agent_pool_id) ON UPDATE CASCADE,
    CONSTRAINT workspace_lock_username_fk FOREIGN KEY (lock_username) REFERENCES users(username) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT workspaces_organization_name_fkey FOREIGN KEY (organization_name) REFERENCES organizations(name) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS workspace_roles (
    role text NOT NULL,
    CONSTRAINT workspace_roles_pkey PRIMARY KEY (role)
);

INSERT INTO workspace_roles (role) VALUES
   ('read'),
   ('plan'),
   ('write'),
   ('admin')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS workspace_permissions (
    workspace_id text NOT NULL,
    team_id text NOT NULL,
    role text NOT NULL,
    CONSTRAINT workspace_permissions_workspace_id_team_id_key UNIQUE (workspace_id, team_id),
    CONSTRAINT workspace_permissions_role_fkey FOREIGN KEY (role) REFERENCES workspace_roles(role) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT workspace_permissions_team_id_fkey FOREIGN KEY (team_id) REFERENCES teams(team_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT workspace_permissions_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS workspace_tags (
    tag_id text NOT NULL,
    workspace_id text NOT NULL,
    CONSTRAINT workspace_tags_tag_id_workspace_id_key UNIQUE (tag_id, workspace_id),
    CONSTRAINT workspace_tags_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES tags(tag_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT workspace_tags_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS state_version_statuses (
    status text NOT NULL,
    CONSTRAINT state_version_statuses_pkey PRIMARY KEY (status)
);

INSERT INTO state_version_statuses (status) VALUES
   ('pending'),
   ('finalized'),
   ('discarded')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS state_versions (
    state_version_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    serial integer NOT NULL,
    state bytea,
    workspace_id text,
    status text NOT NULL,
    CONSTRAINT state_versions_pkey PRIMARY KEY (state_version_id),
    CONSTRAINT state_versions_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT status_fk FOREIGN KEY (status) REFERENCES state_version_statuses(status) ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS state_version_outputs (
    state_version_output_id text NOT NULL,
    name text NOT NULL,
    sensitive boolean NOT NULL,
    type text NOT NULL,
    value bytea NOT NULL,
    state_version_id text NOT NULL,
    CONSTRAINT state_version_outputs_pkey PRIMARY KEY (state_version_output_id),
    CONSTRAINT state_version_outputs_state_version_id_fkey FOREIGN KEY (state_version_id) REFERENCES state_versions(state_version_id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS configuration_versions (
    configuration_version_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    auto_queue_runs boolean NOT NULL,
    source text NOT NULL,
    speculative boolean NOT NULL,
    status text NOT NULL,
    config bytea,
    workspace_id text NOT NULL,
    PRIMARY KEY (configuration_version_id),
    CONSTRAINT configuration_versions_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS configuration_version_status_timestamps (
    configuration_version_id text NOT NULL,
    status text NOT NULL,
    "timestamp" timestamp with time zone NOT NULL,
    PRIMARY KEY (configuration_version_id, status),
    CONSTRAINT configuration_version_status_time_configuration_version_id_fkey FOREIGN KEY (configuration_version_id) REFERENCES configuration_versions(configuration_version_id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS ingress_attributes (
    branch text NOT NULL,
    commit_sha text NOT NULL,
    identifier text NOT NULL,
    is_pull_request boolean NOT NULL,
    on_default_branch boolean NOT NULL,
    configuration_version_id text NOT NULL,
    commit_url text NOT NULL,
    pull_request_number integer NOT NULL,
    pull_request_url text NOT NULL,
    pull_request_title text NOT NULL,
    tag text NOT NULL,
    sender_username text NOT NULL,
    sender_avatar_url text NOT NULL,
    sender_html_url text NOT NULL,
    CONSTRAINT ingress_attributes_configuration_version_id_fkey FOREIGN KEY (configuration_version_id) REFERENCES configuration_versions(configuration_version_id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS destination_types (
    name text NOT NULL,
    PRIMARY KEY (name)
);

INSERT INTO destination_types (name) VALUES
    ('generic'),
    ('gcppubsub'),
    ('email'),
    ('slack')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS notification_configurations (
    notification_configuration_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    name text NOT NULL,
    url text,
    triggers text[],
    destination_type text NOT NULL,
    workspace_id text NOT NULL,
    enabled boolean NOT NULL,
    PRIMARY KEY (notification_configuration_id),
    CONSTRAINT notification_configurations_destination_type_fkey FOREIGN KEY (destination_type) REFERENCES destination_types(name) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT notification_configurations_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS phases (
    phase text NOT NULL,
    PRIMARY KEY (phase)
);

INSERT INTO phases (phase) VALUES
    ('plan'),
    ('apply')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS run_statuses (
    status text NOT NULL,
    CONSTRAINT run_statuses_pkey PRIMARY KEY (status)
);

INSERT INTO run_statuses (status) VALUES
   ('applied'),
   ('apply_queued'),
   ('applying'),
   ('canceled'),
   ('force_canceled'),
   ('confirmed'),
   ('cost_estimated'),
   ('discarded'),
   ('errored'),
   ('pending'),
   ('plan_queued'),
   ('planned'),
   ('planned_and_finished'),
   ('planning')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS runs (
    run_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    cancel_signaled_at timestamp with time zone,
    is_destroy boolean NOT NULL,
    position_in_queue integer NOT NULL,
    refresh boolean NOT NULL,
    refresh_only boolean NOT NULL,
    replace_addrs text[],
    target_addrs text[],
    lock_file bytea,
    status text NOT NULL,
    workspace_id text NOT NULL,
    configuration_version_id text NOT NULL,
    auto_apply boolean NOT NULL,
    plan_only boolean NOT NULL,
    created_by text,
    source text NOT NULL,
    terraform_version text NOT NULL,
    allow_empty_apply boolean DEFAULT false NOT NULL,
    CONSTRAINT runs_pkey PRIMARY KEY (run_id),
    CONSTRAINT runs_configuration_version_id_fkey FOREIGN KEY (configuration_version_id) REFERENCES configuration_versions(configuration_version_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT runs_status_fkey FOREIGN KEY (status) REFERENCES run_statuses(status),
    CONSTRAINT runs_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE
);

ALTER TABLE workspaces
    DROP CONSTRAINT IF EXISTS current_state_version_id_fk,
    DROP CONSTRAINT IF EXISTS latest_run_id_fk,
    DROP CONSTRAINT IF EXISTS lock_run_id_fk,
    ADD CONSTRAINT current_state_version_id_fk FOREIGN KEY (current_state_version_id) REFERENCES state_versions(state_version_id) ON UPDATE CASCADE,
    ADD CONSTRAINT latest_run_id_fk FOREIGN KEY (latest_run_id) REFERENCES runs(run_id) ON UPDATE CASCADE,
    ADD CONSTRAINT lock_run_id_fk FOREIGN KEY (lock_run_id) REFERENCES runs(run_id) ON UPDATE CASCADE;

CREATE TABLE IF NOT EXISTS run_status_timestamps (
    run_id text NOT NULL,
    status text NOT NULL,
    "timestamp" timestamp with time zone NOT NULL,
    CONSTRAINT run_status_timestamps_pkey PRIMARY KEY (run_id, status),
    CONSTRAINT run_status_timestamps_run_id_fkey FOREIGN KEY (run_id) REFERENCES runs(run_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT run_status_timestamps_status_fkey FOREIGN KEY (status) REFERENCES run_statuses(status)
);

CREATE TABLE IF NOT EXISTS run_variables (
    run_id text NOT NULL,
    key text NOT NULL,
    value text NOT NULL,
    CONSTRAINT run_variables_run_id_fkey FOREIGN KEY (run_id) REFERENCES runs(run_id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS phase_statuses (
    status text NOT NULL,
    PRIMARY KEY (status)
);

INSERT INTO phase_statuses (status) VALUES
    ('canceled'),
    ('errored'),
    ('finished'),
    ('pending'),
    ('queued'),
    ('running'),
    ('unreachable')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS phase_status_timestamps (
    run_id text NOT NULL,
    phase text NOT NULL,
    status text NOT NULL,
    "timestamp" timestamp with time zone NOT NULL,
    PRIMARY KEY (run_id, phase, status),
    CONSTRAINT phase_status_timestamps_phase_fkey FOREIGN KEY (phase) REFERENCES phases(phase) ON UPDATE CASCADE,
    CONSTRAINT phase_status_timestamps_run_id_fkey FOREIGN KEY (run_id) REFERENCES runs(run_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT phase_status_timestamps_status_fkey FOREIGN KEY (status) REFERENCES phase_statuses(status)
);

DO
$$
BEGIN
  IF NOT EXISTS (SELECT *
    FROM pg_type typ
         INNER JOIN pg_namespace nsp
                    ON nsp.oid = typ.typnamespace
    WHERE nsp.nspname = current_schema()
          AND typ.typname = 'report') THEN
    CREATE TYPE report AS (
        additions integer,
        changes integer,
        destructions integer
    );
  END IF;
END;
$$
LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS plans (
    run_id text NOT NULL,
    status text NOT NULL,
    plan_bin bytea,
    plan_json bytea,
    resource_report report,
    output_report report,
    PRIMARY KEY (run_id),
    CONSTRAINT plans_run_id_fkey FOREIGN KEY (run_id) REFERENCES runs(run_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT plans_status_fkey FOREIGN KEY (status) REFERENCES phase_statuses(status)
);

CREATE TABLE IF NOT EXISTS applies (
    run_id text NOT NULL,
    status text NOT NULL,
    resource_report report,
    PRIMARY KEY (run_id),
    CONSTRAINT applies_run_id_fkey FOREIGN KEY (run_id) REFERENCES runs(run_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT applies_status_fkey FOREIGN KEY (status) REFERENCES phase_statuses(status)
);

CREATE TABLE IF NOT EXISTS job_phases (
    phase text NOT NULL,
    PRIMARY KEY (phase)
);

INSERT INTO job_phases (phase) VALUES
    ('plan'),
    ('apply')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS job_statuses (
    status text NOT NULL,
    PRIMARY KEY (status)
);

INSERT INTO job_statuses (status) VALUES
    ('unallocated'),
    ('allocated'),
    ('running'),
    ('finished'),
    ('errored'),
    ('canceled')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS jobs (
    run_id text NOT NULL,
    phase text NOT NULL,
    status text NOT NULL,
    agent_id text,
    signaled boolean,
    CONSTRAINT jobs_agent_id_fkey FOREIGN KEY (agent_id) REFERENCES agents(agent_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT jobs_phase_fkey FOREIGN KEY (phase) REFERENCES job_phases(phase) ON UPDATE CASCADE,
    CONSTRAINT jobs_run_id_fkey FOREIGN KEY (run_id) REFERENCES runs(run_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT jobs_status_fkey FOREIGN KEY (status) REFERENCES job_statuses(status) ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS logs (
    run_id text NOT NULL,
    phase text NOT NULL,
    chunk_id integer GENERATED ALWAYS AS IDENTITY NOT NULL,
    chunk bytea NOT NULL,
    _offset integer DEFAULT 0 NOT NULL,
    PRIMARY KEY (run_id, phase, chunk_id),
    CONSTRAINT logs_phase_fkey FOREIGN KEY (phase) REFERENCES phases(phase) ON UPDATE CASCADE,
    CONSTRAINT logs_run_id_fkey FOREIGN KEY (run_id) REFERENCES runs(run_id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS agent_pool_allowed_workspaces (
    agent_pool_id text NOT NULL,
    workspace_id text NOT NULL,
    UNIQUE (agent_pool_id, workspace_id),
    CONSTRAINT agent_pool_allowed_workspaces_agent_pool_id_fkey FOREIGN KEY (agent_pool_id) REFERENCES agent_pools(agent_pool_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT agent_pool_allowed_workspaces_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS module_statuses (
    status text NOT NULL,
    PRIMARY KEY (status)
);

INSERT INTO module_statuses (status) VALUES
   ('pending'),
   ('no_version_tags'),
   ('setup_failed'),
   ('setup_complete')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS module_version_statuses (
    status text NOT NULL,
    PRIMARY KEY (status)
);

INSERT INTO module_version_statuses (status) VALUES
   ('pending'),
   ('cloning'),
   ('clone_failed'),
   ('reg_ingress_req_failed'),
   ('reg_ingressing'),
   ('reg_ingress_failed'),
   ('ok')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS modules (
    module_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    name text NOT NULL,
    provider text NOT NULL,
    status text NOT NULL,
    organization_name text NOT NULL,
    PRIMARY KEY (module_id),
    UNIQUE (organization_name, name, provider),
    CONSTRAINT modules_organization_name_fkey FOREIGN KEY (organization_name) REFERENCES organizations(name) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT modules_status_fkey FOREIGN KEY (status) REFERENCES module_statuses(status) ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS module_versions (
    module_version_id text NOT NULL,
    version text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    status text NOT NULL,
    status_error text,
    module_id text NOT NULL,
    PRIMARY KEY (module_version_id),
    UNIQUE (module_id, version),
    CONSTRAINT module_versions_module_id_fkey FOREIGN KEY (module_id) REFERENCES modules(module_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT module_versions_status_fkey FOREIGN KEY (status) REFERENCES module_version_statuses(status) ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS module_tarballs (
    tarball bytea NOT NULL,
    module_version_id text NOT NULL,
    UNIQUE (module_version_id),
    CONSTRAINT module_tarballs_module_version_id_fkey FOREIGN KEY (module_version_id) REFERENCES module_versions(module_version_id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS github_apps (
    github_app_id bigint NOT NULL,
    webhook_secret text NOT NULL,
    private_key text NOT NULL,
    slug text NOT NULL,
    organization text,
    PRIMARY KEY (github_app_id)
);

CREATE TABLE IF NOT EXISTS vcs_kinds (
    name text NOT NULL,
    PRIMARY KEY (name)
);

INSERT INTO vcs_kinds (name) VALUES
   ('github'),
   ('gitlab')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS vcs_providers (
    vcs_provider_id text NOT NULL,
    token text,
    created_at timestamp with time zone NOT NULL,
    name text NOT NULL,
    vcs_kind text NOT NULL,
    organization_name text NOT NULL,
    github_app_id bigint,
    CONSTRAINT vcs_providers_pkey PRIMARY KEY (vcs_provider_id),
    CONSTRAINT github_app_id_fk FOREIGN KEY (github_app_id) REFERENCES github_apps(github_app_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT vcs_providers_cloud_fkey FOREIGN KEY (vcs_kind) REFERENCES vcs_kinds(name) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT vcs_providers_organization_name_fkey FOREIGN KEY (organization_name) REFERENCES organizations(name) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS github_app_installs (
    github_app_id bigint NOT NULL,
    install_id bigint NOT NULL,
    username text,
    organization text,
    vcs_provider_id text NOT NULL,
    CONSTRAINT github_app_installs_check CHECK ((((username IS NOT NULL) AND (organization IS NULL)) OR ((username IS NULL) AND (organization IS NOT NULL)))),
    CONSTRAINT github_app_installs_github_app_id_fkey FOREIGN KEY (github_app_id) REFERENCES github_apps(github_app_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT github_app_installs_vcs_provider_id_fkey FOREIGN KEY (vcs_provider_id) REFERENCES vcs_providers(vcs_provider_id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS repo_connections (
    module_id text,
    workspace_id text,
    repo_path text NOT NULL,
    vcs_provider_id text NOT NULL,
    CONSTRAINT repo_connections_check CHECK (((module_id IS NULL) <> (workspace_id IS NULL))),
    UNIQUE (module_id),
    UNIQUE (workspace_id),
    CONSTRAINT vcs_provider_id_fk FOREIGN KEY (vcs_provider_id) REFERENCES vcs_providers(vcs_provider_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT repo_connections_module_id_key UNIQUE (module_id),
    CONSTRAINT repo_connections_workspace_id_key UNIQUE (workspace_id)
);

CREATE TABLE IF NOT EXISTS repohooks (
    repohook_id uuid NOT NULL,
    vcs_id text,
    secret text NOT NULL,
    repo_path text NOT NULL,
    vcs_provider_id text NOT NULL,
    CONSTRAINT webhooks_pkey PRIMARY KEY (repohook_id),
    CONSTRAINT webhooks_cloud_id_uniq UNIQUE (repo_path, vcs_provider_id),
    CONSTRAINT vcs_provider_id_fk FOREIGN KEY (vcs_provider_id) REFERENCES vcs_providers(vcs_provider_id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS variable_categories (
    category text NOT NULL,
    CONSTRAINT variable_categories_pkey PRIMARY KEY (category)
);

INSERT INTO variable_categories (category) VALUES
    ('terraform'),
    ('env')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS variables (
    variable_id text NOT NULL,
    key text NOT NULL,
    value text NOT NULL,
    description text NOT NULL,
    category text NOT NULL,
    sensitive boolean NOT NULL,
    hcl boolean NOT NULL,
    version_id text NOT NULL,
    CONSTRAINT variables_pkey PRIMARY KEY (variable_id),
    CONSTRAINT variables_category_fkey FOREIGN KEY (category) REFERENCES variable_categories(category) ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS workspace_variables (
    workspace_id text NOT NULL,
    variable_id text NOT NULL,
    CONSTRAINT workspace_variables_workspace_id_variable_id_key UNIQUE (workspace_id, variable_id),
    CONSTRAINT workspace_variables_variable_id_fkey FOREIGN KEY (variable_id) REFERENCES variables(variable_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT workspace_variables_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS variable_sets (
    variable_set_id text NOT NULL,
    global boolean NOT NULL,
    name text NOT NULL,
    description text NOT NULL,
    organization_name text NOT NULL,
    CONSTRAINT variable_sets_pkey PRIMARY KEY (variable_set_id),
    CONSTRAINT variable_sets_name_key UNIQUE (name),
    CONSTRAINT variable_sets_organization_name_fkey FOREIGN KEY (organization_name) REFERENCES organizations(name) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS variable_set_variables (
    variable_set_id text NOT NULL,
    variable_id text NOT NULL,
    CONSTRAINT variable_set_variables_variable_set_id_variable_id_key UNIQUE (variable_set_id, variable_id),
    CONSTRAINT variable_set_variables_variable_id_fkey FOREIGN KEY (variable_id) REFERENCES variables(variable_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT variable_set_variables_variable_set_id_fkey FOREIGN KEY (variable_set_id) REFERENCES variable_sets(variable_set_id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS variable_set_workspaces (
    variable_set_id text NOT NULL,
    workspace_id text NOT NULL,
    CONSTRAINT variable_set_workspaces_variable_set_id_workspace_id_key UNIQUE (variable_set_id, workspace_id),
    CONSTRAINT variable_set_workspaces_variable_set_id_fkey FOREIGN KEY (variable_set_id) REFERENCES variable_sets(variable_set_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT variable_set_workspaces_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE OR REPLACE FUNCTION agent_pools_notify_event() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    record RECORD;
    notification JSON;
BEGIN
    IF (TG_OP = 'DELETE') THEN
        record = OLD;
    ELSE
        record = NEW;
    END IF;
    notification = json_build_object(
                      'table',TG_TABLE_NAME,
                      'action', TG_OP,
                      'id', record.agent_pool_id);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$;

CREATE OR REPLACE FUNCTION agents_notify_event() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    record RECORD;
    notification JSON;
BEGIN
    IF (TG_OP = 'DELETE') THEN
        record = OLD;
    ELSE
        record = NEW;
    END IF;
    notification = json_build_object(
                      'table',TG_TABLE_NAME,
                      'action', TG_OP,
                      'id', record.agent_id);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$;

CREATE OR REPLACE FUNCTION delete_tags() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    DELETE
    FROM tags
    WHERE NOT EXISTS (
        SELECT FROM workspace_tags wt
        WHERE wt.tag_id = tags.tag_id
    );
    RETURN NULL;
END;
$$;

CREATE OR REPLACE FUNCTION jobs_notify_event() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    record RECORD;
    notification JSON;
BEGIN
    IF (TG_OP = 'DELETE') THEN
        record = OLD;
    ELSE
        record = NEW;
    END IF;
    notification = json_build_object(
                      'table',TG_TABLE_NAME,
                      'action', TG_OP,
                      'id', record.run_id || '/' || record.phase);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$;

CREATE OR REPLACE FUNCTION logs_notify_event() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    record RECORD;
    notification JSON;
BEGIN
    IF (TG_OP = 'DELETE') THEN
        record = OLD;
    ELSE
        record = NEW;
    END IF;
    notification = json_build_object(
                      'table',TG_TABLE_NAME,
                      'action', TG_OP,
                      'id', record.chunk_id::text);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$;

CREATE OR REPLACE FUNCTION notification_configurations_notify_event() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    record RECORD;
    notification JSON;
BEGIN
    IF (TG_OP = 'DELETE') THEN
        record = OLD;
    ELSE
        record = NEW;
    END IF;
    notification = json_build_object(
                      'table',TG_TABLE_NAME,
                      'action', TG_OP,
                      'id', record.notification_configuration_id);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$;

CREATE OR REPLACE FUNCTION organizations_notify_event() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    record RECORD;
    notification JSON;
BEGIN
    IF (TG_OP = 'DELETE') THEN
        record = OLD;
    ELSE
        record = NEW;
    END IF;
    notification = json_build_object(
                      'table',TG_TABLE_NAME,
                      'action', TG_OP,
                      'id', record.organization_id);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$;

CREATE OR REPLACE FUNCTION runs_notify_event() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    record RECORD;
    notification JSON;
BEGIN
    IF (TG_OP = 'DELETE') THEN
        record = OLD;
    ELSE
        record = NEW;
    END IF;
    notification = json_build_object(
                      'table',TG_TABLE_NAME,
                      'action', TG_OP,
                      'id', record.run_id);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$;

CREATE OR REPLACE FUNCTION workspaces_notify_event() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    record RECORD;
    notification JSON;
BEGIN
    IF (TG_OP = 'DELETE') THEN
        record = OLD;
    ELSE
        record = NEW;
    END IF;
    notification = json_build_object(
                      'table',TG_TABLE_NAME,
                      'action', TG_OP,
                      'id', record.workspace_id);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$;

CREATE OR REPLACE TRIGGER delete_tags AFTER DELETE ON workspace_tags FOR EACH STATEMENT EXECUTE FUNCTION delete_tags();

CREATE OR REPLACE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON agent_pools FOR EACH ROW EXECUTE FUNCTION agent_pools_notify_event();

CREATE OR REPLACE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON agents FOR EACH ROW EXECUTE FUNCTION agents_notify_event();

CREATE OR REPLACE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON jobs FOR EACH ROW EXECUTE FUNCTION jobs_notify_event();

CREATE OR REPLACE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON logs FOR EACH ROW EXECUTE FUNCTION logs_notify_event();

CREATE OR REPLACE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON notification_configurations FOR EACH ROW EXECUTE FUNCTION notification_configurations_notify_event();

CREATE OR REPLACE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON organizations FOR EACH ROW EXECUTE FUNCTION organizations_notify_event();

CREATE OR REPLACE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON runs FOR EACH ROW EXECUTE FUNCTION runs_notify_event();

CREATE OR REPLACE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON workspaces FOR EACH ROW EXECUTE FUNCTION workspaces_notify_event();

---- create above / drop below ----

DROP TABLE IF EXISTS variable_set_workspaces;
DROP TABLE IF EXISTS variable_set_variables;
DROP TABLE IF EXISTS variable_sets;
DROP TABLE IF EXISTS workspace_variables;
DROP TABLE IF EXISTS variables;
DROP TABLE IF EXISTS variable_categories;
DROP TABLE IF EXISTS repohooks;
DROP TABLE IF EXISTS repo_connections;
DROP TABLE IF EXISTS github_app_installs;
DROP TABLE IF EXISTS vcs_providers;
DROP TABLE IF EXISTS vcs_kinds;
DROP TABLE IF EXISTS github_apps;
DROP TABLE IF EXISTS module_tarballs;
DROP TABLE IF EXISTS module_versions;
DROP TABLE IF EXISTS modules;
DROP TABLE IF EXISTS module_version_statuses;
DROP TABLE IF EXISTS module_statuses;
DROP TABLE IF EXISTS agent_pool_allowed_workspaces;
DROP TABLE IF EXISTS logs;
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS job_statuses;
DROP TABLE IF EXISTS job_phases;
DROP TABLE IF EXISTS applies;
DROP TABLE IF EXISTS plans;
DROP TYPE IF EXISTS report;
DROP TABLE IF EXISTS phase_status_timestamps;
DROP TABLE IF EXISTS phase_statuses;
DROP TABLE IF EXISTS run_variables;
DROP TABLE IF EXISTS run_status_timestamps;
DROP TABLE IF EXISTS runs;
DROP TABLE IF EXISTS run_statuses;
DROP TABLE IF EXISTS phases;
DROP TABLE IF EXISTS notification_configurations;
DROP TABLE IF EXISTS destination_types;
DROP TABLE IF EXISTS ingress_attributes;
DROP TABLE IF EXISTS configuration_version_status_timestamps;
DROP TABLE IF EXISTS configuration_versions;
DROP TABLE IF EXISTS state_version_outputs;
DROP TABLE IF EXISTS state_versions;
DROP TABLE IF EXISTS state_version_statuses;
DROP TABLE IF EXISTS workspace_tags;
DROP TABLE IF EXISTS workspace_permissions;
DROP TABLE IF EXISTS workspace_roles;
DROP TABLE IF EXISTS workspaces;
DROP TABLE IF EXISTS latest_terraform_version;
DROP TABLE IF EXISTS agents;
DROP TABLE IF EXISTS agent_tokens;
DROP TABLE IF EXISTS agent_statuses;
DROP TABLE IF EXISTS agent_pools;
DROP TABLE IF EXISTS team_tokens;
DROP TABLE IF EXISTS team_memberships;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS tokens;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS organization_tokens;
DROP TABLE IF EXISTS organizations;
DROP FUNCTION IF EXISTS workspaces_notify_event();
DROP FUNCTION IF EXISTS runs_notify_event();
DROP FUNCTION IF EXISTS organizations_notify_event();
DROP FUNCTION IF EXISTS notification_configurations_notify_event();
DROP FUNCTION IF EXISTS logs_notify_event();
DROP FUNCTION IF EXISTS jobs_notify_event();
DROP FUNCTION IF EXISTS delete_tags();
DROP FUNCTION IF EXISTS agents_notify_event();
DROP FUNCTION IF EXISTS agent_pools_notify_event();
