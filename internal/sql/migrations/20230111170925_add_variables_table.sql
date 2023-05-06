-- +goose Up
CREATE TABLE IF NOT EXISTS variable_categories (
    category TEXT PRIMARY KEY
);
INSERT INTO variable_categories (category) VALUES
	('terraform'),
	('env');

CREATE TABLE IF NOT EXISTS variables (
    variable_id     TEXT,
    key             TEXT NOT NULL,
    value           TEXT NOT NULL,
    description     TEXT NOT NULL,
    category        TEXT REFERENCES variable_categories ON UPDATE CASCADE NOT NULL,
    sensitive       BOOL NOT NULL,
    hcl             BOOL NOT NULL,
    workspace_id    TEXT REFERENCES workspaces ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                    PRIMARY KEY (variable_id),
                    UNIQUE (workspace_id, category, key)
);

-- +goose Down
DROP TABLE IF EXISTS variables;
DROP TABLE IF EXISTS variable_categories;
