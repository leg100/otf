package releases

import (
	"context"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql"
)

var q = &Queries{}

type db struct {
	*sql.DB
}

func (db *db) updateLatestVersion(ctx context.Context, v string) error {
	return db.Lock(ctx, "latest_terraform_version", func(ctx context.Context, conn sql.Connection) error {
		rows, err := q.FindLatestTerraformVersion(ctx, conn)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			err = q.InsertLatestTerraformVersion(ctx, conn, sql.String(v))
			if err != nil {
				return err
			}
		} else {
			err = q.UpdateLatestTerraformVersion(ctx, conn, sql.String(v))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *db) getLatest(ctx context.Context) (string, time.Time, error) {
	rows, err := q.FindLatestTerraformVersion(ctx, db.Conn(ctx))
	if err != nil {
		return "", time.Time{}, err
	}
	if len(rows) == 0 {
		return "", time.Time{}, internal.ErrResourceNotFound
	}
	return rows[0].Version.String, rows[0].Checkpoint.Time, nil
}
