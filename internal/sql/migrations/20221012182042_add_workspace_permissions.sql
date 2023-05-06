-- +goose Up
CREATE TABLE IF NOT EXISTS workspace_roles (
    role TEXT PRIMARY KEY
);
INSERT INTO workspace_roles (role) VALUES
	('read'),
	('plan'),
	('write'),
	('admin');

CREATE TABLE IF NOT EXISTS workspace_permissions (
    workspace_id    TEXT REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    team_id         TEXT REFERENCES teams ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    role            TEXT REFERENCES workspace_roles ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                    UNIQUE (workspace_id, team_id)
);

-- +goose Down
DROP TABLE IF EXISTS workspace_permissions;
DROP TABLE IF EXISTS workspace_roles;
