package sql

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

func (db *DB) insertJobStatusTimestamp(ctx context.Context, job otf.Job) error {
	ts, err := job.StatusTimestamp(job.Status())
	if err != nil {
		return err
	}
	_, err = db.InsertJobStatusTimestamp(ctx, pggen.InsertJobStatusTimestampParams{
		JobID:     pgtype.Text{String: job.JobID(), Status: pgtype.Present},
		Status:    pgtype.Text{String: string(job.Status()), Status: pgtype.Present},
		Timestamp: ts,
	})
	return err
}
