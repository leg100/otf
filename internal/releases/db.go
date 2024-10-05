package releases

import (
	"context"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
)

type db struct {
	*sql.DB
}

func (db *db) updateLatestVersion(ctx context.Context, v string) error {
	return db.Lock(ctx, "latest_terraform_version", func(ctx context.Context, q *sqlc.Queries) error {
		rows, err := q.FindLatestTerraformVersion(ctx)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			err = q.InsertLatestTerraformVersion(ctx, sql.String(v))
			if err != nil {
				return err
			}
		} else {
			err = q.UpdateLatestTerraformVersion(ctx, sql.String(v))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *db) getLatest(ctx context.Context) (string, time.Time, error) {
	rows, err := db.Querier(ctx).FindLatestTerraformVersion(ctx)
	if err != nil {
		return "", time.Time{}, err
	}
	if len(rows) == 0 {
		return "", time.Time{}, internal.ErrResourceNotFound
	}
	return rows[0].Version.String, rows[0].Checkpoint.Time, nil
}
