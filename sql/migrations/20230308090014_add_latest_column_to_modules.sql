-- +goose Up
ALTER TABLE modules
	ADD COLUMN latest TEXT,
	ADD CONSTRAINT modules_latest_version_fkey
	FOREIGN KEY (latest) REFERENCES module_versions (module_version_id) ON UPDATE CASCADE
;

-- +goose Down
ALTER TABLE modules DROP COLUMN latest TEXT
;
