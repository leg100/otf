-- +goose Up
CREATE TABLE IF NOT EXISTS organizations (
    organization_id  TEXT,
    created_at       TIMESTAMPTZ NOT NULL,
    updated_at       TIMESTAMPTZ NOT NULL,
    name             TEXT        NOT NULL,
    session_remember INTEGER,
    session_timeout  INTEGER,
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
    trigger_prefixes                TEXT        NOT NULL,
    working_directory               TEXT        NOT NULL,
    organization_id                 TEXT REFERENCES organizations ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                                    PRIMARY KEY (workspace_id)
);

-- postgres does not support declaring a non-unique index in CREATE TABLE
CREATE INDEX ON workspaces (name);

CREATE TABLE IF NOT EXISTS users (
    user_id                 TEXT,
    username                TEXT        NOT NULL,
    created_at              TIMESTAMPTZ NOT NULL,
    updated_at              TIMESTAMPTZ NOT NULL,
    current_organization    TEXT,
                            UNIQUE (username),
                            PRIMARY KEY (user_id)
);

CREATE TABLE IF NOT EXISTS organization_memberships (
    user_id         TEXT REFERENCES users ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    organization_id TEXT REFERENCES organizations ON UPDATE CASCADE ON DELETE CASCADE NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    token           TEXT,
    created_at      TIMESTAMPTZ NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL,
    address         TEXT        NOT NULL,
    flash           JSONB,
    expiry          TIMESTAMPTZ NOT NULL,
    user_id         TEXT REFERENCES users ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                    PRIMARY KEY (token)
);

CREATE TABLE IF NOT EXISTS tokens (
    token_id        TEXT,
    token           TEXT,
    created_at      TIMESTAMPTZ NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL,
    description     TEXT        NOT NULL,
    user_id         TEXT REFERENCES users ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                    UNIQUE (token),
                    PRIMARY KEY (token_id)
);

CREATE TABLE IF NOT EXISTS configuration_versions (
    configuration_version_id     TEXT,
    created_at                   TIMESTAMPTZ NOT NULL,
    updated_at                   TIMESTAMPTZ NOT NULL,
    auto_queue_runs              BOOLEAN     NOT NULL,
    source                       TEXT        NOT NULL,
    speculative                  BOOLEAN     NOT NULL,
    status                       TEXT        NOT NULL,
    status_timestamps            TEXT        NOT NULL,
    config                       BYTEA,
    workspace_id                 TEXT REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                                 PRIMARY KEY (configuration_version_id)
);

CREATE TABLE IF NOT EXISTS runs (
    run_id                   TEXT,
    created_at               TIMESTAMPTZ NOT NULL,
    updated_at               TIMESTAMPTZ NOT NULL,
    is_destroy               BOOLEAN     NOT NULL,
    position_in_queue        INTEGER     NOT NULL,
    refresh                  BOOLEAN     NOT NULL,
    refresh_only             BOOLEAN     NOT NULL,
    status                   TEXT        NOT NULL,
    status_timestamps        TEXT        NOT NULL,
    replace_addrs            TEXT        NOT NULL,
    target_addrs             TEXT        NOT NULL,
    workspace_id             TEXT REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    configuration_version_id TEXT REFERENCES configuration_versions ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                             PRIMARY KEY (run_id)
);

CREATE TABLE IF NOT EXISTS applies (
    apply_id              TEXT,
    created_at            TIMESTAMPTZ NOT NULL,
    updated_at            TIMESTAMPTZ NOT NULL,
    resource_additions    INTEGER     NOT NULL,
    resource_changes      INTEGER     NOT NULL,
    resource_destructions INTEGER     NOT NULL,
    status                TEXT        NOT NULL,
    status_timestamps     TEXT        NOT NULL,
    run_id                TEXT REFERENCES runs ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                          PRIMARY KEY (apply_id)
);

CREATE TABLE IF NOT EXISTS plans (
    plan_id               TEXT,
    created_at            TIMESTAMPTZ NOT NULL,
    updated_at            TIMESTAMPTZ NOT NULL,
    resource_additions    INTEGER     NOT NULL,
    resource_changes      INTEGER     NOT NULL,
    resource_destructions INTEGER     NOT NULL,
    status                TEXT        NOT NULL,
    status_timestamps     TEXT        NOT NULL,
    plan_bin              BYTEA,
    plan_json             BYTEA,
    run_id                TEXT REFERENCES runs ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                          PRIMARY KEY (plan_id)
);

CREATE TABLE IF NOT EXISTS state_versions (
    state_version_id TEXT,
    created_at       TIMESTAMPTZ NOT NULL,
    updated_at       TIMESTAMPTZ NOT NULL,
    serial           INTEGER     NOT NULL,
    vcs_commit_sha   TEXT,
    vcs_commit_url   TEXT,
    state            BYTEA       NOT NULL,
    workspace_id     TEXT REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                     PRIMARY KEY (state_version_id)
);

CREATE TABLE IF NOT EXISTS state_version_outputs (
    state_version_output_id TEXT,
    created_at              TIMESTAMPTZ NOT NULL,
    updated_at              TIMESTAMPTZ NOT NULL,
    name                    TEXT        NOT NULL,
    sensitive               BOOLEAN     NOT NULL,
    type                    TEXT        NOT NULL,
    value                   TEXT        NOT NULL,
    state_version_id        TEXT REFERENCES state_versions ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                            PRIMARY KEY (state_version_output_id)
);

CREATE TABLE IF NOT EXISTS plan_logs (
    plan_id  TEXT REFERENCES plans ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    chunk_id SERIAL  NOT NULL,
    chunk    BYTEA   NOT NULL,
    size     INTEGER NOT NULL,
    start    BOOLEAN NOT NULL,
    _end     BOOLEAN NOT NULL,
             PRIMARY KEY (plan_id, chunk_id)
);

CREATE TABLE IF NOT EXISTS apply_logs (
    apply_id TEXT REFERENCES applies ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    chunk_id SERIAL  NOT NULL,
    chunk    BYTEA   NOT NULL,
    size     INTEGER NOT NULL,
    start    BOOLEAN NOT NULL,
    _end     BOOLEAN NOT NULL,
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
DROP TABLE IF EXISTS tokens;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS organization_memberships;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS workspaces;
DROP TABLE IF EXISTS organizations;
