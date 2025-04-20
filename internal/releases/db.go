package releases

import (
	"context"
	"time"

	"github.com/leg100/otf/internal/sql"
)

type db struct {
	*sql.DB
	engine Engine
}

func (db *db) updateLatestVersion(ctx context.Context, v string) error {
	// TODO: use an UPSERT instead
	return db.Lock(ctx, "latest_engine_version", func(ctx context.Context, conn sql.Connection) error {
		count, err := db.Int(ctx, ` SELECT count(*) FROM latest_engine_version`)
		if err != nil {
			return err
		}
		if count == 0 {
			_, err := db.Exec(ctx, `
INSERT INTO latest_engine_version (
    version,
    checkpoint,
	engine
) VALUES (
    $1,
    current_timestamp,
	$2
)`, v, db.engine.String())
			return err
		} else {
			_, err := db.Exec(ctx, `
UPDATE latest_engine_version
SET version = $1, checkpoint = current_timestamp, engine = $2
`, v, db.engine.String())
			return err
		}
	})
}

func (db *db) getLatest(ctx context.Context) (string, time.Time, error) {
	rows := db.QueryRow(ctx, `
SELECT version, checkpoint
FROM latest_engine_version
WHERE engine = $1
`, db.engine.String())
	var (
		version    string
		checkpoint time.Time
	)
	if err := rows.Scan(&version, &checkpoint); err != nil {
		return "", time.Time{}, err
	}
	return version, checkpoint, nil
}
