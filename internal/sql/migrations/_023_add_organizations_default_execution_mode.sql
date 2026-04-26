-- Add execution modes table and populate with all possible values.
CREATE TABLE execution_modes (
    mode TEXT
);

INSERT INTO execution_modes (mode) VALUES ('remote', 'agent', 'local');

ALTER TABLE execution_modes
    ALTER COLUMN default_execution_mode SET NOT NULL;


-- Add default execution mode column to organizations table, referencing the
-- execution modes table.

ALTER TABLE organizations
    ADD COLUMN default_execution_mode TEXT NOT NULL;

ALTER TABLE organizations
	ADD CONSTRAINT organization_default_execution_mode_fkey FOREIGN KEY (exection_mode) REFERENCES execution_modes(mode) ON UPDATE CASCADE;


-- Update workspaces table's execution mode column to instead reference the new
-- execution modes table, and allow nulls.
ALTER TABLE workspace
	ADD CONSTRAINT workspace_execution_mode_fkey FOREIGN KEY (exection_mode) REFERENCES execution_modes(mode) ON UPDATE CASCADE;

ALTER TABLE workspaces
    ALTER COLUMN execution_mode DROP NOT NULL;

---- create above / drop below ----
