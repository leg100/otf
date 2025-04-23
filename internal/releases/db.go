package releases

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal/sql"
)

type db struct {
	*sql.DB
}

func (db *db) updateLatestVersion(ctx context.Context, engine, v string) error {
	_, err := db.Exec(ctx, `
INSERT INTO latest_engine_version (
    version,
    checkpoint,
	engine
) VALUES (
    @version,
    current_timestamp,
	@engine
) ON CONFLICT (engine) DO UPDATE
SET version		= @version,
	checkpoint	= current_timestamp
`, pgx.NamedArgs{
		"version": v,
		"engine":  engine,
	})
	return err
}

func (db *db) getLatest(ctx context.Context, engine string) (string, time.Time, error) {
	rows := db.QueryRow(ctx, `
SELECT version, checkpoint
FROM latest_engine_version
WHERE engine = $1
`, engine)
	var (
		version    string
		checkpoint time.Time
	)
	if err := rows.Scan(&version, &checkpoint); err != nil {
		return "", time.Time{}, err
	}
	return version, checkpoint, nil
}
