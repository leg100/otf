--
-- PostgreSQL database dump
--

-- Dumped from database version 16.3
-- Dumped by pg_dump version 16.3

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA public;


--
-- Name: report; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.report AS (
	additions integer,
	changes integer,
	destructions integer
);


--
-- Name: agent_pools_notify_event(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE OR REPLACE FUNCTION public.agent_pools_notify_event() RETURNS trigger
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


--
-- Name: agents_notify_event(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE OR REPLACE FUNCTION public.agents_notify_event() RETURNS trigger
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


--
-- Name: delete_tags(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE OR REPLACE FUNCTION public.delete_tags() RETURNS trigger
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


--
-- Name: jobs_notify_event(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE OR REPLACE FUNCTION public.jobs_notify_event() RETURNS trigger
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


--
-- Name: logs_notify_event(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE OR REPLACE FUNCTION public.logs_notify_event() RETURNS trigger
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


--
-- Name: notification_configurations_notify_event(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE OR REPLACE FUNCTION public.notification_configurations_notify_event() RETURNS trigger
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


--
-- Name: organizations_notify_event(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE OR REPLACE FUNCTION public.organizations_notify_event() RETURNS trigger
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


--
-- Name: runs_notify_event(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE OR REPLACE FUNCTION public.runs_notify_event() RETURNS trigger
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


--
-- Name: workspaces_notify_event(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE OR REPLACE FUNCTION public.workspaces_notify_event() RETURNS trigger
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


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: agent_pool_allowed_workspaces; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.agent_pool_allowed_workspaces (
    agent_pool_id text NOT NULL,
    workspace_id text NOT NULL
);


--
-- Name: agent_pools; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.agent_pools (
    agent_pool_id text NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    organization_name text NOT NULL,
    organization_scoped boolean NOT NULL
);


--
-- Name: agent_statuses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.agent_statuses (
    status text NOT NULL
);


--
-- Name: agent_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.agent_tokens (
    agent_token_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    description text NOT NULL,
    agent_pool_id text NOT NULL
);


--
-- Name: agents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.agents (
    agent_id text NOT NULL,
    name text NOT NULL,
    version text NOT NULL,
    max_jobs integer NOT NULL,
    ip_address inet NOT NULL,
    last_ping_at timestamp with time zone NOT NULL,
    last_status_at timestamp with time zone NOT NULL,
    status text NOT NULL,
    agent_pool_id text
);


--
-- Name: applies; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.applies (
    run_id text NOT NULL,
    status text NOT NULL,
    resource_report public.report
);


--
-- Name: configuration_version_status_timestamps; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.configuration_version_status_timestamps (
    configuration_version_id text NOT NULL,
    status text NOT NULL,
    "timestamp" timestamp with time zone NOT NULL
);


--
-- Name: configuration_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.configuration_versions (
    configuration_version_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    auto_queue_runs boolean NOT NULL,
    source text NOT NULL,
    speculative boolean NOT NULL,
    status text NOT NULL,
    config bytea,
    workspace_id text NOT NULL
);


--
-- Name: destination_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.destination_types (
    name text NOT NULL
);


--
-- Name: github_app_installs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.github_app_installs (
    github_app_id bigint NOT NULL,
    install_id bigint NOT NULL,
    username text,
    organization text,
    vcs_provider_id text NOT NULL,
    CONSTRAINT github_app_installs_check CHECK ((((username IS NOT NULL) AND (organization IS NULL)) OR ((username IS NULL) AND (organization IS NOT NULL))))
);


--
-- Name: github_apps; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.github_apps (
    github_app_id bigint NOT NULL,
    webhook_secret text NOT NULL,
    private_key text NOT NULL,
    slug text NOT NULL,
    organization text
);


--
-- Name: goose_db_version; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.goose_db_version (
    id integer NOT NULL,
    version_id bigint NOT NULL,
    is_applied boolean NOT NULL,
    tstamp timestamp without time zone DEFAULT now()
);


--
-- Name: goose_db_version_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.goose_db_version_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: goose_db_version_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.goose_db_version_id_seq OWNED BY public.goose_db_version.id;


--
-- Name: ingress_attributes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.ingress_attributes (
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
    sender_html_url text NOT NULL
);


--
-- Name: job_phases; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.job_phases (
    phase text NOT NULL
);


--
-- Name: job_statuses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.job_statuses (
    status text NOT NULL
);


--
-- Name: jobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.jobs (
    run_id text NOT NULL,
    phase text NOT NULL,
    status text NOT NULL,
    agent_id text,
    signaled boolean
);


--
-- Name: latest_terraform_version; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.latest_terraform_version (
    version text NOT NULL,
    checkpoint timestamp with time zone NOT NULL
);


--
-- Name: logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.logs (
    run_id text NOT NULL,
    phase text NOT NULL,
    chunk_id integer NOT NULL,
    chunk bytea NOT NULL,
    _offset integer DEFAULT 0 NOT NULL
);


--
-- Name: logs_chunk_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.logs ALTER COLUMN chunk_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.logs_chunk_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: module_statuses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.module_statuses (
    status text NOT NULL
);


--
-- Name: module_tarballs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.module_tarballs (
    tarball bytea NOT NULL,
    module_version_id text NOT NULL
);


--
-- Name: module_version_statuses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.module_version_statuses (
    status text NOT NULL
);


--
-- Name: module_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.module_versions (
    module_version_id text NOT NULL,
    version text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    status text NOT NULL,
    status_error text,
    module_id text NOT NULL
);


--
-- Name: modules; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.modules (
    module_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    name text NOT NULL,
    provider text NOT NULL,
    status text NOT NULL,
    organization_name text NOT NULL
);


--
-- Name: notification_configurations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.notification_configurations (
    notification_configuration_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    name text NOT NULL,
    url text,
    triggers text[],
    destination_type text NOT NULL,
    workspace_id text NOT NULL,
    enabled boolean NOT NULL
);


--
-- Name: organization_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.organization_tokens (
    organization_token_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    organization_name text NOT NULL,
    expiry timestamp with time zone
);


--
-- Name: organizations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.organizations (
    organization_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    name text NOT NULL,
    session_remember integer,
    session_timeout integer,
    email text,
    collaborator_auth_policy text,
    allow_force_delete_workspaces boolean DEFAULT false NOT NULL,
    cost_estimation_enabled boolean DEFAULT false NOT NULL
);


--
-- Name: phase_status_timestamps; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.phase_status_timestamps (
    run_id text NOT NULL,
    phase text NOT NULL,
    status text NOT NULL,
    "timestamp" timestamp with time zone NOT NULL
);


--
-- Name: phase_statuses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.phase_statuses (
    status text NOT NULL
);


--
-- Name: phases; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.phases (
    phase text NOT NULL
);


--
-- Name: plans; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.plans (
    run_id text NOT NULL,
    status text NOT NULL,
    plan_bin bytea,
    plan_json bytea,
    resource_report public.report,
    output_report public.report
);


--
-- Name: registry_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.registry_sessions (
    token text NOT NULL,
    expiry timestamp with time zone NOT NULL,
    organization_name text NOT NULL
);


--
-- Name: repo_connections; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.repo_connections (
    module_id text,
    workspace_id text,
    repo_path text NOT NULL,
    vcs_provider_id text NOT NULL,
    CONSTRAINT repo_connections_check CHECK (((module_id IS NULL) <> (workspace_id IS NULL)))
);


--
-- Name: repohooks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.repohooks (
    repohook_id uuid NOT NULL,
    vcs_id text,
    secret text NOT NULL,
    repo_path text NOT NULL,
    vcs_provider_id text NOT NULL
);


--
-- Name: run_status_timestamps; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.run_status_timestamps (
    run_id text NOT NULL,
    status text NOT NULL,
    "timestamp" timestamp with time zone NOT NULL
);


--
-- Name: run_statuses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.run_statuses (
    status text NOT NULL
);


--
-- Name: run_variables; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.run_variables (
    run_id text NOT NULL,
    key text NOT NULL,
    value text NOT NULL
);


--
-- Name: runs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.runs (
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
    allow_empty_apply boolean DEFAULT false NOT NULL
);


--
-- Name: sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.sessions (
    token text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    address text NOT NULL,
    expiry timestamp with time zone NOT NULL,
    username text NOT NULL
);


--
-- Name: state_version_outputs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.state_version_outputs (
    state_version_output_id text NOT NULL,
    name text NOT NULL,
    sensitive boolean NOT NULL,
    type text NOT NULL,
    value bytea NOT NULL,
    state_version_id text NOT NULL
);


--
-- Name: state_version_statuses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.state_version_statuses (
    status text NOT NULL
);


--
-- Name: state_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.state_versions (
    state_version_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    serial integer NOT NULL,
    state bytea,
    workspace_id text,
    status text NOT NULL
);


--
-- Name: tags; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.tags (
    tag_id text NOT NULL,
    name text NOT NULL,
    organization_name text NOT NULL
);


--
-- Name: team_memberships; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.team_memberships (
    team_id text NOT NULL,
    username text NOT NULL
);


--
-- Name: team_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.team_tokens (
    team_token_id text NOT NULL,
    description text,
    created_at timestamp with time zone NOT NULL,
    team_id text NOT NULL,
    expiry timestamp with time zone
);


--
-- Name: teams; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.teams (
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
    permission_manage_providers boolean DEFAULT false NOT NULL
);


--
-- Name: tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.tokens (
    token_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    description text NOT NULL,
    username text NOT NULL
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.users (
    user_id text NOT NULL,
    username text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    site_admin boolean DEFAULT false NOT NULL
);


--
-- Name: variable_categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.variable_categories (
    category text NOT NULL
);


--
-- Name: variable_set_variables; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.variable_set_variables (
    variable_set_id text NOT NULL,
    variable_id text NOT NULL
);


--
-- Name: variable_set_workspaces; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.variable_set_workspaces (
    variable_set_id text NOT NULL,
    workspace_id text NOT NULL
);


--
-- Name: variable_sets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.variable_sets (
    variable_set_id text NOT NULL,
    global boolean NOT NULL,
    name text NOT NULL,
    description text NOT NULL,
    organization_name text NOT NULL
);


--
-- Name: variables; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.variables (
    variable_id text NOT NULL,
    key text NOT NULL,
    value text NOT NULL,
    description text NOT NULL,
    category text NOT NULL,
    sensitive boolean NOT NULL,
    hcl boolean NOT NULL,
    version_id text NOT NULL
);


--
-- Name: vcs_kinds; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.vcs_kinds (
    name text NOT NULL
);


--
-- Name: vcs_providers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.vcs_providers (
    vcs_provider_id text NOT NULL,
    token text,
    created_at timestamp with time zone NOT NULL,
    name text NOT NULL,
    vcs_kind text NOT NULL,
    organization_name text NOT NULL,
    github_app_id bigint
);


--
-- Name: workspace_permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.workspace_permissions (
    workspace_id text NOT NULL,
    team_id text NOT NULL,
    role text NOT NULL
);


--
-- Name: workspace_roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.workspace_roles (
    role text NOT NULL
);


--
-- Name: workspace_tags; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.workspace_tags (
    tag_id text NOT NULL,
    workspace_id text NOT NULL
);


--
-- Name: workspace_variables; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.workspace_variables (
    workspace_id text NOT NULL,
    variable_id text NOT NULL
);


--
-- Name: workspaces; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.workspaces (
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
    CONSTRAINT agent_pool_chk CHECK (((execution_mode <> 'agent'::text) OR (agent_pool_id IS NOT NULL)))
);


--
-- Name: goose_db_version id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.goose_db_version ALTER COLUMN id SET DEFAULT nextval('public.goose_db_version_id_seq'::regclass);


--
-- Data for Name: agent_pool_allowed_workspaces; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: agent_pools; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: agent_statuses; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.agent_statuses VALUES ('busy') ON CONFLICT DO NOTHING;
INSERT INTO public.agent_statuses VALUES ('idle') ON CONFLICT DO NOTHING;
INSERT INTO public.agent_statuses VALUES ('exited') ON CONFLICT DO NOTHING;
INSERT INTO public.agent_statuses VALUES ('errored') ON CONFLICT DO NOTHING;
INSERT INTO public.agent_statuses VALUES ('unknown') ON CONFLICT DO NOTHING;


--
-- Data for Name: agent_tokens; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: agents; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: applies; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: configuration_version_status_timestamps; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: configuration_versions; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: destination_types; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.destination_types VALUES ('generic') ON CONFLICT DO NOTHING;
INSERT INTO public.destination_types VALUES ('gcppubsub') ON CONFLICT DO NOTHING;
INSERT INTO public.destination_types VALUES ('email') ON CONFLICT DO NOTHING;
INSERT INTO public.destination_types VALUES ('slack') ON CONFLICT DO NOTHING;


--
-- Data for Name: github_app_installs; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: github_apps; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: goose_db_version; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.goose_db_version VALUES (1, 0, true, '2024-09-29 16:23:09.400517') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (2, 1, true, '2024-09-29 16:23:24.895609') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (3, 2, true, '2024-09-29 16:23:25.055724') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (4, 3, true, '2024-09-29 16:23:25.058042') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (5, 4, true, '2024-09-29 16:23:25.061121') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (6, 5, true, '2024-09-29 16:23:25.062753') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (7, 6, true, '2024-09-29 16:23:25.073736') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (8, 7, true, '2024-09-29 16:23:25.076471') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (9, 8, true, '2024-09-29 16:23:25.090908') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (10, 20221012182042, true, '2024-09-29 16:23:25.092197') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (11, 20221017170815, true, '2024-09-29 16:23:25.107376') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (12, 20221103102133, true, '2024-09-29 16:23:25.111554') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (13, 20221104124638, true, '2024-09-29 16:23:25.12636') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (14, 20221119134414, true, '2024-09-29 16:23:25.144093') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (15, 20221207194038, true, '2024-09-29 16:23:25.148809') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (16, 20221208185308, true, '2024-09-29 16:23:25.150432') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (17, 20230105092014, true, '2024-09-29 16:23:25.206606') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (18, 20230110113551, true, '2024-09-29 16:23:25.207894') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (19, 20230111170925, true, '2024-09-29 16:23:25.21288') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (20, 20230114084040, true, '2024-09-29 16:23:25.231977') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (21, 20230114150642, true, '2024-09-29 16:23:25.234905') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (22, 20230114153752, true, '2024-09-29 16:23:25.236267') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (23, 20230114183712, true, '2024-09-29 16:23:25.238935') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (24, 20230121080309, true, '2024-09-29 16:23:25.241273') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (25, 20230121091625, true, '2024-09-29 16:23:25.243731') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (26, 20230201113207, true, '2024-09-29 16:23:25.244882') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (27, 20230223181343, true, '2024-09-29 16:23:25.24772') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (28, 20230224094530, true, '2024-09-29 16:23:25.252335') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (29, 20230304130926, true, '2024-09-29 16:23:25.256751') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (30, 20230306081037, true, '2024-09-29 16:23:25.272064') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (31, 20230308141234, true, '2024-09-29 16:23:25.273601') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (32, 20230328194515, true, '2024-09-29 16:23:25.277996') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (33, 20230330105923, true, '2024-09-29 16:23:25.286187') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (34, 20230405133920, true, '2024-09-29 16:23:25.28847') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (35, 20230406182826, true, '2024-09-29 16:23:25.290322') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (36, 20230413160941, true, '2024-09-29 16:23:25.291607') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (37, 20230417181955, true, '2024-09-29 16:23:25.293765') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (38, 20230427144312, true, '2024-09-29 16:23:25.299195') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (39, 20230509164012, true, '2024-09-29 16:23:25.317286') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (40, 20230528095219, true, '2024-09-29 16:23:25.332476') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (41, 20230613192456, true, '2024-09-29 16:23:25.334986') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (42, 20230621161651, true, '2024-09-29 16:23:25.339769') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (43, 20230630180152, true, '2024-09-29 16:23:25.357454') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (44, 20230701125813, true, '2024-09-29 16:23:25.359253') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (45, 20230705172351, true, '2024-09-29 16:23:25.360738') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (46, 20230706145931, true, '2024-09-29 16:23:25.362048') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (47, 20230709110236, true, '2024-09-29 16:23:25.363656') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (48, 20230723164327, true, '2024-09-29 16:23:25.372346') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (49, 20230727183502, true, '2024-09-29 16:23:25.38348') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (50, 20230802093004, true, '2024-09-29 16:23:25.3849') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (51, 20230802183439, true, '2024-09-29 16:23:25.387081') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (52, 20230803084105, true, '2024-09-29 16:23:25.388302') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (53, 20230803084106, true, '2024-09-29 16:23:25.389628') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (54, 20230803105841, true, '2024-09-29 16:23:25.391318') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (55, 20230806154420, true, '2024-09-29 16:23:25.397236') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (56, 20230806181842, true, '2024-09-29 16:23:25.398413') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (57, 20230815143723, true, '2024-09-29 16:23:25.399878') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (58, 20231010191539, true, '2024-09-29 16:23:25.432007') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (59, 20231010191540, true, '2024-09-29 16:23:25.436675') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (60, 20231010191541, true, '2024-09-29 16:23:25.451231') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (61, 20231024213208, true, '2024-09-29 16:23:25.462249') ON CONFLICT DO NOTHING;
INSERT INTO public.goose_db_version VALUES (62, 20231031184346, true, '2024-09-29 16:23:25.471006') ON CONFLICT DO NOTHING;


--
-- Data for Name: ingress_attributes; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: job_phases; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.job_phases VALUES ('plan') ON CONFLICT DO NOTHING;
INSERT INTO public.job_phases VALUES ('apply') ON CONFLICT DO NOTHING;


--
-- Data for Name: job_statuses; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.job_statuses VALUES ('unallocated') ON CONFLICT DO NOTHING;
INSERT INTO public.job_statuses VALUES ('allocated') ON CONFLICT DO NOTHING;
INSERT INTO public.job_statuses VALUES ('running') ON CONFLICT DO NOTHING;
INSERT INTO public.job_statuses VALUES ('finished') ON CONFLICT DO NOTHING;
INSERT INTO public.job_statuses VALUES ('errored') ON CONFLICT DO NOTHING;
INSERT INTO public.job_statuses VALUES ('canceled') ON CONFLICT DO NOTHING;


--
-- Data for Name: jobs; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: latest_terraform_version; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: logs; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: module_statuses; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.module_statuses VALUES ('pending') ON CONFLICT DO NOTHING;
INSERT INTO public.module_statuses VALUES ('no_version_tags') ON CONFLICT DO NOTHING;
INSERT INTO public.module_statuses VALUES ('setup_failed') ON CONFLICT DO NOTHING;
INSERT INTO public.module_statuses VALUES ('setup_complete') ON CONFLICT DO NOTHING;


--
-- Data for Name: module_tarballs; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: module_version_statuses; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.module_version_statuses VALUES ('pending') ON CONFLICT DO NOTHING;
INSERT INTO public.module_version_statuses VALUES ('cloning') ON CONFLICT DO NOTHING;
INSERT INTO public.module_version_statuses VALUES ('clone_failed') ON CONFLICT DO NOTHING;
INSERT INTO public.module_version_statuses VALUES ('reg_ingress_req_failed') ON CONFLICT DO NOTHING;
INSERT INTO public.module_version_statuses VALUES ('reg_ingressing') ON CONFLICT DO NOTHING;
INSERT INTO public.module_version_statuses VALUES ('reg_ingress_failed') ON CONFLICT DO NOTHING;
INSERT INTO public.module_version_statuses VALUES ('ok') ON CONFLICT DO NOTHING;


--
-- Data for Name: module_versions; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: modules; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: notification_configurations; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: organization_tokens; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: organizations; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: phase_status_timestamps; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: phase_statuses; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.phase_statuses VALUES ('canceled') ON CONFLICT DO NOTHING;
INSERT INTO public.phase_statuses VALUES ('errored') ON CONFLICT DO NOTHING;
INSERT INTO public.phase_statuses VALUES ('finished') ON CONFLICT DO NOTHING;
INSERT INTO public.phase_statuses VALUES ('pending') ON CONFLICT DO NOTHING;
INSERT INTO public.phase_statuses VALUES ('queued') ON CONFLICT DO NOTHING;
INSERT INTO public.phase_statuses VALUES ('running') ON CONFLICT DO NOTHING;
INSERT INTO public.phase_statuses VALUES ('unreachable') ON CONFLICT DO NOTHING;


--
-- Data for Name: phases; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.phases VALUES ('plan') ON CONFLICT DO NOTHING;
INSERT INTO public.phases VALUES ('apply') ON CONFLICT DO NOTHING;


--
-- Data for Name: plans; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: registry_sessions; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: repo_connections; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: repohooks; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: run_status_timestamps; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: run_statuses; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.run_statuses VALUES ('applied') ON CONFLICT DO NOTHING;
INSERT INTO public.run_statuses VALUES ('apply_queued') ON CONFLICT DO NOTHING;
INSERT INTO public.run_statuses VALUES ('applying') ON CONFLICT DO NOTHING;
INSERT INTO public.run_statuses VALUES ('canceled') ON CONFLICT DO NOTHING;
INSERT INTO public.run_statuses VALUES ('force_canceled') ON CONFLICT DO NOTHING;
INSERT INTO public.run_statuses VALUES ('confirmed') ON CONFLICT DO NOTHING;
INSERT INTO public.run_statuses VALUES ('discarded') ON CONFLICT DO NOTHING;
INSERT INTO public.run_statuses VALUES ('errored') ON CONFLICT DO NOTHING;
INSERT INTO public.run_statuses VALUES ('pending') ON CONFLICT DO NOTHING;
INSERT INTO public.run_statuses VALUES ('plan_queued') ON CONFLICT DO NOTHING;
INSERT INTO public.run_statuses VALUES ('planned') ON CONFLICT DO NOTHING;
INSERT INTO public.run_statuses VALUES ('planned_and_finished') ON CONFLICT DO NOTHING;
INSERT INTO public.run_statuses VALUES ('planning') ON CONFLICT DO NOTHING;
INSERT INTO public.run_statuses VALUES ('cost_estimated') ON CONFLICT DO NOTHING;


--
-- Data for Name: run_variables; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: runs; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: sessions; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: state_version_outputs; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: state_version_statuses; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.state_version_statuses VALUES ('pending') ON CONFLICT DO NOTHING;
INSERT INTO public.state_version_statuses VALUES ('finalized') ON CONFLICT DO NOTHING;
INSERT INTO public.state_version_statuses VALUES ('discarded') ON CONFLICT DO NOTHING;


--
-- Data for Name: state_versions; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: tags; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: team_memberships; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: team_tokens; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: teams; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: tokens; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.users VALUES ('user-site-admin', 'site-admin', '2024-09-29 16:23:25.055724+01', '2024-09-29 16:23:25.055724+01', false) ON CONFLICT DO NOTHING;


--
-- Data for Name: variable_categories; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.variable_categories VALUES ('terraform') ON CONFLICT DO NOTHING;
INSERT INTO public.variable_categories VALUES ('env') ON CONFLICT DO NOTHING;


--
-- Data for Name: variable_set_variables; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: variable_set_workspaces; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: variable_sets; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: variables; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: vcs_kinds; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.vcs_kinds VALUES ('github') ON CONFLICT DO NOTHING;
INSERT INTO public.vcs_kinds VALUES ('gitlab') ON CONFLICT DO NOTHING;


--
-- Data for Name: vcs_providers; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: workspace_permissions; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: workspace_roles; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.workspace_roles VALUES ('read') ON CONFLICT DO NOTHING;
INSERT INTO public.workspace_roles VALUES ('plan') ON CONFLICT DO NOTHING;
INSERT INTO public.workspace_roles VALUES ('write') ON CONFLICT DO NOTHING;
INSERT INTO public.workspace_roles VALUES ('admin') ON CONFLICT DO NOTHING;


--
-- Data for Name: workspace_tags; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: workspace_variables; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: workspaces; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Name: goose_db_version_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.goose_db_version_id_seq', 62, true);


--
-- Name: logs_chunk_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.logs_chunk_id_seq', 1, false);


--
-- Name: agent_pool_allowed_workspaces agent_pool_allowed_workspaces_agent_pool_id_workspace_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_pool_allowed_workspaces
    ADD CONSTRAINT agent_pool_allowed_workspaces_agent_pool_id_workspace_id_key UNIQUE (agent_pool_id, workspace_id);


--
-- Name: agent_pools agent_pools_organization_name_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_pools
    ADD CONSTRAINT agent_pools_organization_name_name_key UNIQUE (organization_name, name);


--
-- Name: agent_pools agent_pools_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_pools
    ADD CONSTRAINT agent_pools_pkey PRIMARY KEY (agent_pool_id);


--
-- Name: agent_statuses agent_statuses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_statuses
    ADD CONSTRAINT agent_statuses_pkey PRIMARY KEY (status);


--
-- Name: agent_tokens agent_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_tokens
    ADD CONSTRAINT agent_tokens_pkey PRIMARY KEY (agent_token_id);


--
-- Name: agents agents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agents
    ADD CONSTRAINT agents_pkey PRIMARY KEY (agent_id);


--
-- Name: applies applies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.applies
    ADD CONSTRAINT applies_pkey PRIMARY KEY (run_id);


--
-- Name: vcs_kinds clouds_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vcs_kinds
    ADD CONSTRAINT clouds_pkey PRIMARY KEY (name);


--
-- Name: configuration_version_status_timestamps configuration_version_status_timestamps_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.configuration_version_status_timestamps
    ADD CONSTRAINT configuration_version_status_timestamps_pkey PRIMARY KEY (configuration_version_id, status);


--
-- Name: configuration_versions configuration_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.configuration_versions
    ADD CONSTRAINT configuration_versions_pkey PRIMARY KEY (configuration_version_id);


--
-- Name: destination_types destination_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.destination_types
    ADD CONSTRAINT destination_types_pkey PRIMARY KEY (name);


--
-- Name: github_apps github_apps_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.github_apps
    ADD CONSTRAINT github_apps_pkey PRIMARY KEY (github_app_id);


--
-- Name: goose_db_version goose_db_version_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.goose_db_version
    ADD CONSTRAINT goose_db_version_pkey PRIMARY KEY (id);


--
-- Name: job_phases job_phases_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_phases
    ADD CONSTRAINT job_phases_pkey PRIMARY KEY (phase);


--
-- Name: job_statuses job_statuses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_statuses
    ADD CONSTRAINT job_statuses_pkey PRIMARY KEY (status);


--
-- Name: logs logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.logs
    ADD CONSTRAINT logs_pkey PRIMARY KEY (run_id, phase, chunk_id);


--
-- Name: module_statuses module_statuses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_statuses
    ADD CONSTRAINT module_statuses_pkey PRIMARY KEY (status);


--
-- Name: module_tarballs module_tarballs_module_version_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_tarballs
    ADD CONSTRAINT module_tarballs_module_version_id_key UNIQUE (module_version_id);


--
-- Name: module_version_statuses module_version_statuses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_version_statuses
    ADD CONSTRAINT module_version_statuses_pkey PRIMARY KEY (status);


--
-- Name: module_versions module_versions_module_id_version_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_versions
    ADD CONSTRAINT module_versions_module_id_version_key UNIQUE (module_id, version);


--
-- Name: module_versions module_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_versions
    ADD CONSTRAINT module_versions_pkey PRIMARY KEY (module_version_id);


--
-- Name: modules modules_org_name_provider_uniq; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.modules
    ADD CONSTRAINT modules_org_name_provider_uniq UNIQUE (organization_name, name, provider);


--
-- Name: modules modules_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.modules
    ADD CONSTRAINT modules_pkey PRIMARY KEY (module_id);


--
-- Name: notification_configurations notification_configurations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_configurations
    ADD CONSTRAINT notification_configurations_pkey PRIMARY KEY (notification_configuration_id);


--
-- Name: organization_tokens organization_tokens_organization_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_tokens
    ADD CONSTRAINT organization_tokens_organization_name_key UNIQUE (organization_name);


--
-- Name: organization_tokens organization_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_tokens
    ADD CONSTRAINT organization_tokens_pkey PRIMARY KEY (organization_token_id);


--
-- Name: organizations organizations_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_name_key UNIQUE (name);


--
-- Name: organizations organizations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_pkey PRIMARY KEY (organization_id);


--
-- Name: phase_status_timestamps phase_status_timestamps_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.phase_status_timestamps
    ADD CONSTRAINT phase_status_timestamps_pkey PRIMARY KEY (run_id, phase, status);


--
-- Name: phase_statuses phase_statuses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.phase_statuses
    ADD CONSTRAINT phase_statuses_pkey PRIMARY KEY (status);


--
-- Name: phases phases_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.phases
    ADD CONSTRAINT phases_pkey PRIMARY KEY (phase);


--
-- Name: plans plans_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.plans
    ADD CONSTRAINT plans_pkey PRIMARY KEY (run_id);


--
-- Name: registry_sessions registry_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.registry_sessions
    ADD CONSTRAINT registry_sessions_pkey PRIMARY KEY (token);


--
-- Name: repo_connections repo_connections_module_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repo_connections
    ADD CONSTRAINT repo_connections_module_id_key UNIQUE (module_id);


--
-- Name: repo_connections repo_connections_workspace_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repo_connections
    ADD CONSTRAINT repo_connections_workspace_id_key UNIQUE (workspace_id);


--
-- Name: run_status_timestamps run_status_timestamps_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.run_status_timestamps
    ADD CONSTRAINT run_status_timestamps_pkey PRIMARY KEY (run_id, status);


--
-- Name: run_statuses run_statuses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.run_statuses
    ADD CONSTRAINT run_statuses_pkey PRIMARY KEY (status);


--
-- Name: runs runs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.runs
    ADD CONSTRAINT runs_pkey PRIMARY KEY (run_id);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (token);


--
-- Name: state_version_outputs state_version_outputs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.state_version_outputs
    ADD CONSTRAINT state_version_outputs_pkey PRIMARY KEY (state_version_output_id);


--
-- Name: state_version_statuses state_version_statuses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.state_version_statuses
    ADD CONSTRAINT state_version_statuses_pkey PRIMARY KEY (status);


--
-- Name: state_versions state_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.state_versions
    ADD CONSTRAINT state_versions_pkey PRIMARY KEY (state_version_id);


--
-- Name: tags tags_organization_name_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_organization_name_name_key UNIQUE (organization_name, name);


--
-- Name: tags tags_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_pkey PRIMARY KEY (tag_id);


--
-- Name: teams team_name_uniq; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT team_name_uniq UNIQUE (organization_name, name);


--
-- Name: team_tokens team_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_tokens
    ADD CONSTRAINT team_tokens_pkey PRIMARY KEY (team_token_id);


--
-- Name: team_tokens team_tokens_team_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_tokens
    ADD CONSTRAINT team_tokens_team_id_key UNIQUE (team_id);


--
-- Name: teams teams_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT teams_pkey PRIMARY KEY (team_id);


--
-- Name: tokens tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tokens
    ADD CONSTRAINT tokens_pkey PRIMARY KEY (token_id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (user_id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: variable_categories variable_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variable_categories
    ADD CONSTRAINT variable_categories_pkey PRIMARY KEY (category);


--
-- Name: variable_set_variables variable_set_variables_variable_set_id_variable_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variable_set_variables
    ADD CONSTRAINT variable_set_variables_variable_set_id_variable_id_key UNIQUE (variable_set_id, variable_id);


--
-- Name: variable_set_workspaces variable_set_workspaces_variable_set_id_workspace_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variable_set_workspaces
    ADD CONSTRAINT variable_set_workspaces_variable_set_id_workspace_id_key UNIQUE (variable_set_id, workspace_id);


--
-- Name: variable_sets variable_sets_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variable_sets
    ADD CONSTRAINT variable_sets_name_key UNIQUE (name);


--
-- Name: variable_sets variable_sets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variable_sets
    ADD CONSTRAINT variable_sets_pkey PRIMARY KEY (variable_set_id);


--
-- Name: variables variables_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variables
    ADD CONSTRAINT variables_pkey PRIMARY KEY (variable_id);


--
-- Name: vcs_providers vcs_providers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vcs_providers
    ADD CONSTRAINT vcs_providers_pkey PRIMARY KEY (vcs_provider_id);


--
-- Name: repohooks webhooks_cloud_id_uniq; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repohooks
    ADD CONSTRAINT webhooks_cloud_id_uniq UNIQUE (repo_path, vcs_provider_id);


--
-- Name: repohooks webhooks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repohooks
    ADD CONSTRAINT webhooks_pkey PRIMARY KEY (repohook_id);


--
-- Name: workspaces workspace_name_uniq; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspaces
    ADD CONSTRAINT workspace_name_uniq UNIQUE (organization_name, name);


--
-- Name: workspace_permissions workspace_permissions_workspace_id_team_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspace_permissions
    ADD CONSTRAINT workspace_permissions_workspace_id_team_id_key UNIQUE (workspace_id, team_id);


--
-- Name: workspace_roles workspace_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspace_roles
    ADD CONSTRAINT workspace_roles_pkey PRIMARY KEY (role);


--
-- Name: workspace_tags workspace_tags_tag_id_workspace_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspace_tags
    ADD CONSTRAINT workspace_tags_tag_id_workspace_id_key UNIQUE (tag_id, workspace_id);


--
-- Name: workspace_variables workspace_variables_workspace_id_variable_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspace_variables
    ADD CONSTRAINT workspace_variables_workspace_id_variable_id_key UNIQUE (workspace_id, variable_id);


--
-- Name: workspaces workspaces_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspaces
    ADD CONSTRAINT workspaces_pkey PRIMARY KEY (workspace_id);


--
-- Name: workspace_tags delete_tags; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER delete_tags AFTER DELETE ON public.workspace_tags FOR EACH STATEMENT EXECUTE FUNCTION public.delete_tags();


--
-- Name: agent_pools notify_event; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON public.agent_pools FOR EACH ROW EXECUTE FUNCTION public.agent_pools_notify_event();


--
-- Name: agents notify_event; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON public.agents FOR EACH ROW EXECUTE FUNCTION public.agents_notify_event();


--
-- Name: jobs notify_event; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON public.jobs FOR EACH ROW EXECUTE FUNCTION public.jobs_notify_event();


--
-- Name: logs notify_event; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON public.logs FOR EACH ROW EXECUTE FUNCTION public.logs_notify_event();


--
-- Name: notification_configurations notify_event; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON public.notification_configurations FOR EACH ROW EXECUTE FUNCTION public.notification_configurations_notify_event();


--
-- Name: organizations notify_event; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON public.organizations FOR EACH ROW EXECUTE FUNCTION public.organizations_notify_event();


--
-- Name: runs notify_event; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON public.runs FOR EACH ROW EXECUTE FUNCTION public.runs_notify_event();


--
-- Name: workspaces notify_event; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON public.workspaces FOR EACH ROW EXECUTE FUNCTION public.workspaces_notify_event();


--
-- Name: agent_pool_allowed_workspaces agent_pool_allowed_workspaces_agent_pool_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_pool_allowed_workspaces
    ADD CONSTRAINT agent_pool_allowed_workspaces_agent_pool_id_fkey FOREIGN KEY (agent_pool_id) REFERENCES public.agent_pools(agent_pool_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: agent_pool_allowed_workspaces agent_pool_allowed_workspaces_workspace_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_pool_allowed_workspaces
    ADD CONSTRAINT agent_pool_allowed_workspaces_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: workspaces agent_pool_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspaces
    ADD CONSTRAINT agent_pool_fk FOREIGN KEY (agent_pool_id) REFERENCES public.agent_pools(agent_pool_id) ON UPDATE CASCADE;


--
-- Name: agent_tokens agent_pool_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_tokens
    ADD CONSTRAINT agent_pool_id_fk FOREIGN KEY (agent_pool_id) REFERENCES public.agent_pools(agent_pool_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: agent_pools agent_pools_organization_name_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_pools
    ADD CONSTRAINT agent_pools_organization_name_fkey FOREIGN KEY (organization_name) REFERENCES public.organizations(name) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: agents agents_agent_pool_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agents
    ADD CONSTRAINT agents_agent_pool_id_fkey FOREIGN KEY (agent_pool_id) REFERENCES public.agent_pools(agent_pool_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: agents agents_status_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agents
    ADD CONSTRAINT agents_status_fkey FOREIGN KEY (status) REFERENCES public.agent_statuses(status) ON UPDATE CASCADE;


--
-- Name: applies applies_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.applies
    ADD CONSTRAINT applies_run_id_fkey FOREIGN KEY (run_id) REFERENCES public.runs(run_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: applies applies_status_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.applies
    ADD CONSTRAINT applies_status_fkey FOREIGN KEY (status) REFERENCES public.phase_statuses(status);


--
-- Name: configuration_version_status_timestamps configuration_version_status_time_configuration_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.configuration_version_status_timestamps
    ADD CONSTRAINT configuration_version_status_time_configuration_version_id_fkey FOREIGN KEY (configuration_version_id) REFERENCES public.configuration_versions(configuration_version_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: configuration_versions configuration_versions_workspace_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.configuration_versions
    ADD CONSTRAINT configuration_versions_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: workspaces current_state_version_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspaces
    ADD CONSTRAINT current_state_version_id_fk FOREIGN KEY (current_state_version_id) REFERENCES public.state_versions(state_version_id) ON UPDATE CASCADE;


--
-- Name: vcs_providers github_app_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vcs_providers
    ADD CONSTRAINT github_app_id_fk FOREIGN KEY (github_app_id) REFERENCES public.github_apps(github_app_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: github_app_installs github_app_installs_github_app_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.github_app_installs
    ADD CONSTRAINT github_app_installs_github_app_id_fkey FOREIGN KEY (github_app_id) REFERENCES public.github_apps(github_app_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: github_app_installs github_app_installs_vcs_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.github_app_installs
    ADD CONSTRAINT github_app_installs_vcs_provider_id_fkey FOREIGN KEY (vcs_provider_id) REFERENCES public.vcs_providers(vcs_provider_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: ingress_attributes ingress_attributes_configuration_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ingress_attributes
    ADD CONSTRAINT ingress_attributes_configuration_version_id_fkey FOREIGN KEY (configuration_version_id) REFERENCES public.configuration_versions(configuration_version_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: jobs jobs_agent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jobs
    ADD CONSTRAINT jobs_agent_id_fkey FOREIGN KEY (agent_id) REFERENCES public.agents(agent_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: jobs jobs_phase_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jobs
    ADD CONSTRAINT jobs_phase_fkey FOREIGN KEY (phase) REFERENCES public.job_phases(phase) ON UPDATE CASCADE;


--
-- Name: jobs jobs_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jobs
    ADD CONSTRAINT jobs_run_id_fkey FOREIGN KEY (run_id) REFERENCES public.runs(run_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: jobs jobs_status_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jobs
    ADD CONSTRAINT jobs_status_fkey FOREIGN KEY (status) REFERENCES public.job_statuses(status) ON UPDATE CASCADE;


--
-- Name: workspaces latest_run_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspaces
    ADD CONSTRAINT latest_run_id_fk FOREIGN KEY (latest_run_id) REFERENCES public.runs(run_id) ON UPDATE CASCADE;


--
-- Name: workspaces lock_run_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspaces
    ADD CONSTRAINT lock_run_id_fk FOREIGN KEY (lock_run_id) REFERENCES public.runs(run_id) ON UPDATE CASCADE;


--
-- Name: logs logs_phase_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.logs
    ADD CONSTRAINT logs_phase_fkey FOREIGN KEY (phase) REFERENCES public.phases(phase) ON UPDATE CASCADE;


--
-- Name: logs logs_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.logs
    ADD CONSTRAINT logs_run_id_fkey FOREIGN KEY (run_id) REFERENCES public.runs(run_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: module_tarballs module_tarballs_module_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_tarballs
    ADD CONSTRAINT module_tarballs_module_version_id_fkey FOREIGN KEY (module_version_id) REFERENCES public.module_versions(module_version_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: module_versions module_versions_module_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_versions
    ADD CONSTRAINT module_versions_module_id_fkey FOREIGN KEY (module_id) REFERENCES public.modules(module_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: module_versions module_versions_status_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_versions
    ADD CONSTRAINT module_versions_status_fkey FOREIGN KEY (status) REFERENCES public.module_version_statuses(status) ON UPDATE CASCADE;


--
-- Name: modules modules_organization_name_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.modules
    ADD CONSTRAINT modules_organization_name_fkey FOREIGN KEY (organization_name) REFERENCES public.organizations(name) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: modules modules_status_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.modules
    ADD CONSTRAINT modules_status_fkey FOREIGN KEY (status) REFERENCES public.module_statuses(status) ON UPDATE CASCADE;


--
-- Name: notification_configurations notification_configurations_destination_type_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_configurations
    ADD CONSTRAINT notification_configurations_destination_type_fkey FOREIGN KEY (destination_type) REFERENCES public.destination_types(name) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: notification_configurations notification_configurations_workspace_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_configurations
    ADD CONSTRAINT notification_configurations_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: organization_tokens organization_tokens_organization_name_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_tokens
    ADD CONSTRAINT organization_tokens_organization_name_fkey FOREIGN KEY (organization_name) REFERENCES public.organizations(name) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: phase_status_timestamps phase_status_timestamps_phase_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.phase_status_timestamps
    ADD CONSTRAINT phase_status_timestamps_phase_fkey FOREIGN KEY (phase) REFERENCES public.phases(phase) ON UPDATE CASCADE;


--
-- Name: phase_status_timestamps phase_status_timestamps_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.phase_status_timestamps
    ADD CONSTRAINT phase_status_timestamps_run_id_fkey FOREIGN KEY (run_id) REFERENCES public.plans(run_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: phase_status_timestamps phase_status_timestamps_status_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.phase_status_timestamps
    ADD CONSTRAINT phase_status_timestamps_status_fkey FOREIGN KEY (status) REFERENCES public.phase_statuses(status);


--
-- Name: plans plans_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.plans
    ADD CONSTRAINT plans_run_id_fkey FOREIGN KEY (run_id) REFERENCES public.runs(run_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: plans plans_status_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.plans
    ADD CONSTRAINT plans_status_fkey FOREIGN KEY (status) REFERENCES public.phase_statuses(status);


--
-- Name: registry_sessions registry_sessions_organization_name_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.registry_sessions
    ADD CONSTRAINT registry_sessions_organization_name_fkey FOREIGN KEY (organization_name) REFERENCES public.organizations(name) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: repo_connections repo_connections_module_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repo_connections
    ADD CONSTRAINT repo_connections_module_id_fkey FOREIGN KEY (module_id) REFERENCES public.modules(module_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: repo_connections repo_connections_workspace_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repo_connections
    ADD CONSTRAINT repo_connections_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: run_status_timestamps run_status_timestamps_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.run_status_timestamps
    ADD CONSTRAINT run_status_timestamps_run_id_fkey FOREIGN KEY (run_id) REFERENCES public.runs(run_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: run_status_timestamps run_status_timestamps_status_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.run_status_timestamps
    ADD CONSTRAINT run_status_timestamps_status_fkey FOREIGN KEY (status) REFERENCES public.run_statuses(status);


--
-- Name: run_variables run_variables_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.run_variables
    ADD CONSTRAINT run_variables_run_id_fkey FOREIGN KEY (run_id) REFERENCES public.runs(run_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: runs runs_configuration_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.runs
    ADD CONSTRAINT runs_configuration_version_id_fkey FOREIGN KEY (configuration_version_id) REFERENCES public.configuration_versions(configuration_version_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: runs runs_status_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.runs
    ADD CONSTRAINT runs_status_fkey FOREIGN KEY (status) REFERENCES public.run_statuses(status);


--
-- Name: runs runs_workspace_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.runs
    ADD CONSTRAINT runs_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: sessions session_username_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT session_username_fk FOREIGN KEY (username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: state_version_outputs state_version_outputs_state_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.state_version_outputs
    ADD CONSTRAINT state_version_outputs_state_version_id_fkey FOREIGN KEY (state_version_id) REFERENCES public.state_versions(state_version_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: state_versions state_versions_workspace_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.state_versions
    ADD CONSTRAINT state_versions_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: state_versions status_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.state_versions
    ADD CONSTRAINT status_fk FOREIGN KEY (status) REFERENCES public.state_version_statuses(status) ON UPDATE CASCADE;


--
-- Name: tags tags_organization_name_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_organization_name_fkey FOREIGN KEY (organization_name) REFERENCES public.organizations(name) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: team_memberships team_member_username_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_memberships
    ADD CONSTRAINT team_member_username_fk FOREIGN KEY (username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: team_memberships team_memberships_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_memberships
    ADD CONSTRAINT team_memberships_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: team_tokens team_tokens_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_tokens
    ADD CONSTRAINT team_tokens_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: teams teams_organization_name_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT teams_organization_name_fkey FOREIGN KEY (organization_name) REFERENCES public.organizations(name) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: tokens token_username_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tokens
    ADD CONSTRAINT token_username_fk FOREIGN KEY (username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: variable_set_variables variable_set_variables_variable_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variable_set_variables
    ADD CONSTRAINT variable_set_variables_variable_id_fkey FOREIGN KEY (variable_id) REFERENCES public.variables(variable_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: variable_set_variables variable_set_variables_variable_set_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variable_set_variables
    ADD CONSTRAINT variable_set_variables_variable_set_id_fkey FOREIGN KEY (variable_set_id) REFERENCES public.variable_sets(variable_set_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: variable_set_workspaces variable_set_workspaces_variable_set_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variable_set_workspaces
    ADD CONSTRAINT variable_set_workspaces_variable_set_id_fkey FOREIGN KEY (variable_set_id) REFERENCES public.variable_sets(variable_set_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: variable_set_workspaces variable_set_workspaces_workspace_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variable_set_workspaces
    ADD CONSTRAINT variable_set_workspaces_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: variable_sets variable_sets_organization_name_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variable_sets
    ADD CONSTRAINT variable_sets_organization_name_fkey FOREIGN KEY (organization_name) REFERENCES public.organizations(name) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: variables variables_category_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.variables
    ADD CONSTRAINT variables_category_fkey FOREIGN KEY (category) REFERENCES public.variable_categories(category) ON UPDATE CASCADE;


--
-- Name: repohooks vcs_provider_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repohooks
    ADD CONSTRAINT vcs_provider_id_fk FOREIGN KEY (vcs_provider_id) REFERENCES public.vcs_providers(vcs_provider_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: repo_connections vcs_provider_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repo_connections
    ADD CONSTRAINT vcs_provider_id_fk FOREIGN KEY (vcs_provider_id) REFERENCES public.vcs_providers(vcs_provider_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: vcs_providers vcs_providers_cloud_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vcs_providers
    ADD CONSTRAINT vcs_providers_cloud_fkey FOREIGN KEY (vcs_kind) REFERENCES public.vcs_kinds(name) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: vcs_providers vcs_providers_organization_name_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vcs_providers
    ADD CONSTRAINT vcs_providers_organization_name_fkey FOREIGN KEY (organization_name) REFERENCES public.organizations(name) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: workspaces workspace_lock_username_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspaces
    ADD CONSTRAINT workspace_lock_username_fk FOREIGN KEY (lock_username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: workspace_permissions workspace_permissions_role_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspace_permissions
    ADD CONSTRAINT workspace_permissions_role_fkey FOREIGN KEY (role) REFERENCES public.workspace_roles(role) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: workspace_permissions workspace_permissions_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspace_permissions
    ADD CONSTRAINT workspace_permissions_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: workspace_permissions workspace_permissions_workspace_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspace_permissions
    ADD CONSTRAINT workspace_permissions_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: workspace_tags workspace_tags_tag_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspace_tags
    ADD CONSTRAINT workspace_tags_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES public.tags(tag_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: workspace_tags workspace_tags_workspace_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspace_tags
    ADD CONSTRAINT workspace_tags_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: workspace_variables workspace_variables_variable_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspace_variables
    ADD CONSTRAINT workspace_variables_variable_id_fkey FOREIGN KEY (variable_id) REFERENCES public.variables(variable_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: workspace_variables workspace_variables_workspace_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspace_variables
    ADD CONSTRAINT workspace_variables_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: workspaces workspaces_organization_name_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspaces
    ADD CONSTRAINT workspaces_organization_name_fkey FOREIGN KEY (organization_name) REFERENCES public.organizations(name) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--