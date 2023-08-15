-- +goose Up

--
-- Create workspace_variables table, and move workspace_id from variables to
-- workspace_variables
--
CREATE TABLE IF NOT EXISTS workspace_variables (
    workspace_id TEXT REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    variable_id TEXT REFERENCES variables ON UPDATE CASCADE ON DELETE CASCADE NOT NULL
);

UPDATE workspace_variables
SET workspace_id = v.workspace_id,
    variable_id = v.variable_id
FROM variables v;

ALTER TABLE variables DROP column workspace_id;

--
-- Create variable_sets and variable_sets_variables tables
--

CREATE TABLE IF NOT EXISTS variable_sets (
    variable_set_id TEXT NOT NULL,
    global BOOL NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    PRIMARY KEY (variable_set_id)
);

CREATE TABLE IF NOT EXISTS variable_set_variables (
    variable_set_id TEXT REFERENCES variable_sets ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    variable_id TEXT REFERENCES variables ON UPDATE CASCADE ON DELETE CASCADE NOT NULL
);

-- +goose Down

--
-- Drop variable_sets and variable_set_variables tables
--
DROP TABLE IF EXISTS variable_set_variables;
DROP TABLE IF EXISTS variable_sets;

--
-- Move workspace_id from workspace_variables back to variables, and drop
-- workspace_variables table
--
ALTER TABLE variables ADD column workspace_id TEXT REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE;

UPDATE variables v
SET workspace_id = wv.workspace_id
FROM workspace_variables wv
WHERE v.variable_id = wv.variable_id;

ALTER TABLE variables ALTER COLUMN workspace_id SET NOT NULL;

DROP TABLE IF EXISTS workspace_variables;
