-- Because of the way variables are referenced via their parents (workspaces
-- and variable sets) indirectly via junction tables, when the parents are
-- deleted the variables are not also deleted, leaving them in an orphaned
-- state. This migration addresses that by removing the junction tables and
-- instead referencing the parents directly from foreign keys in the variables
-- table.
--
-- Remove any existing orphan variables (not referenced by either junction table)
DELETE FROM variables
WHERE variable_id NOT IN (SELECT variable_id FROM workspace_variables)
AND   variable_id NOT IN (SELECT variable_id FROM variable_set_variables);

-- Bypass junction tables by adding foreign keys directly to the variables table.
ALTER TABLE variables
    ADD COLUMN workspace_id TEXT REFERENCES workspaces(workspace_id) ON UPDATE CASCADE ON DELETE CASCADE,
    ADD COLUMN variable_set_id TEXT REFERENCES variable_sets(variable_set_id) ON UPDATE CASCADE ON DELETE CASCADE;

UPDATE variables
SET workspace_id = workspace_variables.workspace_id
FROM workspace_variables
WHERE workspace_variables.variable_id = variables.variable_id;

UPDATE variables
SET variable_set_id = variable_set_variables.variable_set_id
FROM variable_set_variables
WHERE variable_set_variables.variable_id = variables.variable_id;

-- Add check constraint enforcing that one of the keys must be not null.
ALTER TABLE variables
ADD CONSTRAINT variables_parent_not_null_check CHECK ((workspace_id IS NULL) <> (variable_set_id IS NULL));

-- Drop junction tables
DROP TABLE IF EXISTS workspace_variables;
DROP TABLE IF EXISTS variable_set_variables;

---- create above / drop below ----

CREATE TABLE IF NOT EXISTS variable_set_variables (
    variable_set_id text NOT NULL,
    variable_id text NOT NULL
);

CREATE TABLE IF NOT EXISTS workspace_variables (
    workspace_id text NOT NULL,
    variable_id text NOT NULL
);

INSERT INTO variable_set_variables (variable_set_id, variable_id)
SELECT variable_set_id, variable_id
FROM variables
WHERE variable_set_id IS NOT NULL
;

INSERT INTO workspace_variables (workspace_id, variable_id)
SELECT workspace_id, variable_id
FROM variables
WHERE workspace_id IS NOT NULL
;

ALTER TABLE variables
    DROP COLUMN workspace_id,
    DROP COLUMN variable_set_id;
