CREATE TABLE policy_sets (
    policy_set_id text PRIMARY KEY,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    organization_name text NOT NULL REFERENCES organizations(name) ON UPDATE CASCADE ON DELETE CASCADE,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    UNIQUE (organization_name, name)
);

CREATE TABLE policies (
    policy_id text PRIMARY KEY,
    policy_set_id text NOT NULL REFERENCES policy_sets(policy_set_id) ON UPDATE CASCADE ON DELETE CASCADE,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    enforcement_level text NOT NULL,
    source text NOT NULL,
    UNIQUE (policy_set_id, name)
);

CREATE TABLE policy_set_workspaces (
    policy_set_id text NOT NULL REFERENCES policy_sets(policy_set_id) ON UPDATE CASCADE ON DELETE CASCADE,
    workspace_id text NOT NULL REFERENCES workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE,
    PRIMARY KEY (policy_set_id, workspace_id)
);

CREATE TABLE policy_checks (
    policy_check_id text PRIMARY KEY,
    run_id text NOT NULL REFERENCES runs(run_id) ON UPDATE CASCADE ON DELETE CASCADE,
    workspace_id text NOT NULL REFERENCES workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE,
    policy_set_id text NOT NULL REFERENCES policy_sets(policy_set_id) ON UPDATE CASCADE ON DELETE CASCADE,
    policy_id text NOT NULL REFERENCES policies(policy_id) ON UPDATE CASCADE ON DELETE CASCADE,
    organization_name text NOT NULL REFERENCES organizations(name) ON UPDATE CASCADE ON DELETE CASCADE,
    policy_name text NOT NULL,
    policy_set_name text NOT NULL,
    enforcement_level text NOT NULL,
    passed boolean NOT NULL,
    output text NOT NULL DEFAULT '',
    created_at timestamp with time zone NOT NULL
);

INSERT INTO run_statuses(status) VALUES ('policy_checked') ON CONFLICT DO NOTHING;
INSERT INTO run_statuses(status) VALUES ('policy_soft_failed') ON CONFLICT DO NOTHING;
INSERT INTO run_statuses(status) VALUES ('policy_failed') ON CONFLICT DO NOTHING;

---- create above / drop below ----

DELETE FROM run_statuses WHERE status IN ('policy_checked', 'policy_soft_failed', 'policy_failed');

DROP TABLE policy_checks;
DROP TABLE policy_set_workspaces;
DROP TABLE policies;
DROP TABLE policy_sets;
