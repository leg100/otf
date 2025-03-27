package releases

import (
	"context"
	"time"

	"github.com/leg100/otf/internal/sql"
)

type db struct {
	*sql.DB
}

func (db *db) updateLatestVersion(ctx context.Context, v string) error {
	return db.Lock(ctx, "latest_terraform_version", func(ctx context.Context, conn sql.Connection) error {
		count, err := db.Int(ctx, ` SELECT count(*) FROM latest_terraform_version`)
		if err != nil {
			return err
		}
		if count == 0 {
			_, err := db.Exec(ctx, `
INSERT INTO latest_terraform_version (
    version,
    checkpoint
) VALUES (
    $1,
    current_timestamp
)`, v)
			return err
		} else {
			_, err := db.Exec(ctx, `
UPDATE latest_terraform_version
SET version = $1, checkpoint = current_timestamp
`, v)
			return err
		}
	})
}

func (db *db) getLatest(ctx context.Context) (string, time.Time, error) {
	rows := db.QueryRow(ctx, `
SELECT version, checkpoint
FROM latest_terraform_version
`)
	var (
		version    string
		checkpoint time.Time
	)
	if err := rows.Scan(&version, &checkpoint); err != nil {
		return "", time.Time{}, err
	}
	return version, checkpoint, nil
}
