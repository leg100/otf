package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAddLatestToModules, downAddLatestToModules)
}

func upAddLatestToModules(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE modules
ADD COLUMN latest TEXT,
ADD CONSTRAINT modules_latest_version_fkey
	FOREIGN KEY (latest) REFERENCES module_versions (module_version_id) ON UPDATE CASCADE`)
	return err
}

func downAddLatestToModules(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE modules DROP COLUMN latest TEXT`)
	return err
}
