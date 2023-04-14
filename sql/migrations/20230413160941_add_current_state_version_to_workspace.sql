-- +goose Up
ALTER TABLE workspaces ADD COLUMN current_state_version_id TEXT;
ALTER TABLE workspaces ADD CONSTRAINT current_state_version_id_fk FOREIGN KEY (current_state_version_id) REFERENCES state_versions ON UPDATE CASCADE;

UPDATE workspaces w
SET current_state_version_id = sv.state_version_id
FROM (
    SELECT DISTINCT ON (workspace_id) workspace_id, state_version_id
    FROM state_versions
    ORDER BY workspace_id, serial DESC, created_at DESC
) sv
WHERE w.workspace_id = sv.workspace_id
;

-- +goose Down
ALTER TABLE workspaces DROP CONSTRAINT current_state_version_id_fk;
ALTER TABLE workspaces DROP COLUMN current_state_version_id;
