CREATE TABLE policy_modules (
    policy_module_id text PRIMARY KEY,
    policy_set_id text NOT NULL REFERENCES policy_sets(policy_set_id) ON UPDATE CASCADE ON DELETE CASCADE,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    name text NOT NULL,
    path text NOT NULL,
    source text NOT NULL,
    UNIQUE (policy_set_id, path),
    UNIQUE (policy_set_id, name)
);

---- create above / drop below ----

DROP TABLE policy_modules;
