-- +goose Up
CREATE TABLE IF NOT EXISTS organizations (
    organization_id  TEXT,
    created_at       TIMESTAMPTZ NOT NULL,
    updated_at       TIMESTAMPTZ NOT NULL,
    name             TEXT        NOT NULL,
    session_remember INTEGER     NOT NULL,
    session_timeout  INTEGER     NOT NULL,
                     UNIQUE  (name),
                     PRIMARY KEY (organization_id)
);

CREATE TABLE IF NOT EXISTS workspaces (
    workspace_id                    TEXT,
    created_at                      TIMESTAMPTZ NOT NULL,
    updated_at                      TIMESTAMPTZ NOT NULL,
    allow_destroy_plan              BOOLEAN     NOT NULL,
    auto_apply                      BOOLEAN     NOT NULL,
    can_queue_destroy_plan          BOOLEAN     NOT NULL,
    description                     TEXT        NOT NULL,
    environment                     TEXT        NOT NULL,
    execution_mode                  TEXT        NOT NULL,
    file_triggers_enabled           BOOLEAN     NOT NULL,
    global_remote_state             BOOLEAN     NOT NULL,
    locked                          BOOLEAN     NOT NULL,
    migration_environment           TEXT        NOT NULL,
    name                            TEXT        NOT NULL,
    queue_all_runs                  BOOLEAN     NOT NULL,
    speculative_enabled             BOOLEAN     NOT NULL,
    source_name                     TEXT        NOT NULL,
    source_url                      TEXT        NOT NULL,
    structured_run_output_enabled   BOOLEAN     NOT NULL,
    terraform_version               TEXT        NOT NULL,
    trigger_prefixes                TEXT[],
    working_directory               TEXT        NOT NULL,
    organization_id                 TEXT REFERENCES organizations ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                                    PRIMARY KEY (workspace_id),
                                    UNIQUE (name, organization_id)
);

CREATE TABLE IF NOT EXISTS users (
    user_id                 TEXT,
    username                TEXT        NOT NULL,
    created_at              TIMESTAMPTZ NOT NULL,
    updated_at              TIMESTAMPTZ NOT NULL,
                            PRIMARY KEY (user_id),
                            UNIQUE (username)
);

CREATE TABLE IF NOT EXISTS organization_memberships (
    user_id         TEXT REFERENCES users ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    organization_id TEXT REFERENCES organizations ON UPDATE CASCADE ON DELETE CASCADE NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    token           TEXT,
    created_at      TIMESTAMPTZ NOT NULL,
    address         TEXT        NOT NULL,
    expiry          TIMESTAMPTZ NOT NULL,
    user_id         TEXT REFERENCES users ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                    PRIMARY KEY (token)
);

CREATE TABLE IF NOT EXISTS tokens (
    token_id        TEXT,
    token           TEXT,
    created_at      TIMESTAMPTZ NOT NULL,
    description     TEXT        NOT NULL,
    user_id         TEXT REFERENCES users ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                    PRIMARY KEY (token_id),
                    UNIQUE (token)
);

CREATE TABLE IF NOT EXISTS configuration_versions (
    configuration_version_id     TEXT,
    created_at                   TIMESTAMPTZ NOT NULL,
    auto_queue_runs              BOOLEAN     NOT NULL,
    source                       TEXT        NOT NULL,
    speculative                  BOOLEAN     NOT NULL,
    status                       TEXT        NOT NULL,
    config                       BYTEA,
    workspace_id                 TEXT REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                                 PRIMARY KEY (configuration_version_id)
);

CREATE TABLE IF NOT EXISTS configuration_version_status_timestamps (
    configuration_version_id TEXT REFERENCES configuration_versions ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    status                   TEXT        NOT NULL,
    timestamp                TIMESTAMPTZ NOT NULL,
                             PRIMARY KEY (configuration_version_id, status)
);

CREATE TABLE IF NOT EXISTS run_statuses (
    status TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS plan_statuses (
    status TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS apply_statuses (
    status TEXT PRIMARY KEY
);

INSERT INTO run_statuses (status) VALUES
	('applied'),
	('apply_queued'),
	('applying'),
	('canceled'),
	('force_canceled'),
	('confirmed'),
	('discarded'),
	('errored'),
	('pending'),
	('plan_queued'),
	('planned'),
	('planned_and_finished'),
	('planning');

INSERT INTO plan_statuses (status) VALUES
    ('canceled'),
    ('errored'),
    ('finished'),
    ('pending'),
    ('queued'),
    ('running'),
    ('unreachable');

INSERT INTO apply_statuses (status) VALUES
    ('canceled'),
    ('errored'),
    ('finished'),
    ('pending'),
    ('queued'),
    ('running'),
    ('unreachable');

CREATE TABLE IF NOT EXISTS runs (
    run_id                          TEXT,
    created_at                      TIMESTAMPTZ     NOT NULL,
    is_destroy                      BOOLEAN         NOT NULL,
    position_in_queue               INTEGER         NOT NULL,
    refresh                         BOOLEAN         NOT NULL,
    refresh_only                    BOOLEAN         NOT NULL,
    replace_addrs                   TEXT[],
    target_addrs                    TEXT[],
    status                          TEXT REFERENCES run_statuses  NOT NULL,
    workspace_id                    TEXT REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    configuration_version_id        TEXT REFERENCES configuration_versions ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                                    PRIMARY KEY (run_id)
);

CREATE TABLE IF NOT EXISTS run_status_timestamps (
    run_id      TEXT        REFERENCES runs ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    status      TEXT        REFERENCES run_statuses NOT NULL,
    timestamp   TIMESTAMPTZ NOT NULL,
                PRIMARY KEY (run_id, status)
);

CREATE TABLE IF NOT EXISTS plans (
    plan_id                     TEXT            NOT NULL,
    run_id                      TEXT REFERENCES runs ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    plan_bin                    BYTEA,
    plan_json                   BYTEA,
    additions                   INTEGER         NOT NULL,
    changes                     INTEGER         NOT NULL,
    destructions                INTEGER         NOT NULL,
    status                      TEXT REFERENCES plan_statuses NOT NULL,
                                PRIMARY KEY (plan_id)
);

CREATE TABLE IF NOT EXISTS plan_status_timestamps (
    plan_id      TEXT       REFERENCES plans ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    status      TEXT        REFERENCES plan_statuses NOT NULL,
    timestamp   TIMESTAMPTZ NOT NULL,
                PRIMARY KEY (plan_id, status)
);

CREATE TABLE IF NOT EXISTS applies (
    apply_id          TEXT            NOT NULL,
    run_id            TEXT REFERENCES runs ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    additions         INTEGER         NOT NULL,
    changes           INTEGER         NOT NULL,
    destructions      INTEGER         NOT NULL,
    status            TEXT REFERENCES apply_statuses NOT NULL,
                      PRIMARY KEY (apply_id)
);

CREATE TABLE IF NOT EXISTS apply_status_timestamps (
    apply_id    TEXT        REFERENCES applies ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    status      TEXT        REFERENCES apply_statuses NOT NULL,
    timestamp   TIMESTAMPTZ NOT NULL,
                PRIMARY KEY (apply_id, status)
);

CREATE TABLE IF NOT EXISTS plan_logs (
    plan_id     TEXT REFERENCES plans ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    chunk_id    INT GENERATED ALWAYS AS IDENTITY,
    chunk       BYTEA   NOT NULL,
                PRIMARY KEY (plan_id, chunk_id)
);

CREATE TABLE IF NOT EXISTS apply_logs (
    apply_id    TEXT REFERENCES applies ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    chunk_id    INT GENERATED ALWAYS AS IDENTITY,
    chunk       BYTEA   NOT NULL,
                PRIMARY KEY (apply_id, chunk_id)
);

CREATE TABLE IF NOT EXISTS state_versions (
    state_version_id TEXT,
    created_at       TIMESTAMPTZ NOT NULL,
    serial           INTEGER     NOT NULL,
    state            BYTEA       NOT NULL,
    workspace_id     TEXT REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE,
                     PRIMARY KEY (state_version_id)
);

CREATE TABLE IF NOT EXISTS state_version_outputs (
    state_version_output_id TEXT,
    name                    TEXT        NOT NULL,
    sensitive               BOOLEAN     NOT NULL,
    type                    TEXT        NOT NULL,
    value                   TEXT        NOT NULL,
    state_version_id        TEXT REFERENCES state_versions ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                            PRIMARY KEY (state_version_output_id)
);

-- +goose Down
DROP TABLE IF EXISTS state_version_outputs;
DROP TABLE IF EXISTS state_versions;
DROP TABLE IF EXISTS apply_logs;
DROP TABLE IF EXISTS plan_logs;
DROP TABLE IF EXISTS apply_status_timestamps;
DROP TABLE IF EXISTS applies;
DROP TABLE IF EXISTS plan_status_timestamps;
DROP TABLE IF EXISTS plans;
DROP TABLE IF EXISTS run_status_timestamps;
DROP TABLE IF EXISTS runs;
DROP TABLE IF EXISTS apply_statuses;
DROP TABLE IF EXISTS plan_statuses;
DROP TABLE IF EXISTS run_statuses;
DROP TABLE IF EXISTS configuration_version_status_timestamps;
DROP TABLE IF EXISTS configuration_versions;
DROP TABLE IF EXISTS tokens;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS organization_memberships;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS workspaces;
DROP TABLE IF EXISTS organizations;
