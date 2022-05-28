package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

var (
	_ otf.ConfigurationVersionStore = (*ConfigurationVersionDB)(nil)
)

type ConfigurationVersionDB struct {
	*pgxpool.Pool
}

func NewConfigurationVersionDB(conn *pgxpool.Pool) *ConfigurationVersionDB {
	return &ConfigurationVersionDB{
		Pool: conn,
	}
}

func (db ConfigurationVersionDB) Create(cv *otf.ConfigurationVersion) error {
	ctx := context.Background()

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := pggen.NewQuerier(tx)

	_, err = q.InsertConfigurationVersion(ctx, pggen.InsertConfigurationVersionParams{
		ID:            pgtype.Text{String: cv.ID(), Status: pgtype.Present},
		CreatedAt:     cv.CreatedAt(),
		AutoQueueRuns: cv.AutoQueueRuns(),
		Source:        pgtype.Text{String: string(cv.Source()), Status: pgtype.Present},
		Speculative:   cv.Speculative(),
		Status:        pgtype.Text{String: string(cv.Status()), Status: pgtype.Present},
		WorkspaceID:   pgtype.Text{String: cv.Workspace.ID(), Status: pgtype.Present},
	})
	if err != nil {
		return err
	}

	// Insert timestamp for current status
	ts, err := q.InsertConfigurationVersionStatusTimestamp(ctx,
		pgtype.Text{String: cv.ID(), Status: pgtype.Present},
		pgtype.Text{String: string(cv.Status()), Status: pgtype.Present},
	)
	if err != nil {
		return err
	}
	cv.AddStatusTimestamp(otf.ConfigurationStatus(ts.Status.String), ts.Timestamp)

	return tx.Commit(ctx)
}

func (db ConfigurationVersionDB) Upload(ctx context.Context, id string, fn func(*otf.ConfigurationVersion, otf.ConfigUploader) error) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := pggen.NewQuerier(tx)

	// select ...for update
	result, err := q.FindConfigurationVersionByIDForUpdate(ctx, false, pgtype.Text{String: id, Status: pgtype.Present})
	if err != nil {
		return err
	}
	cv, err := otf.UnmarshalConfigurationVersionDBResult(otf.ConfigurationVersionDBResult(result))
	if err != nil {
		return err
	}

	if err := fn(cv, newConfigUploader(tx, cv.ID())); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (db ConfigurationVersionDB) List(workspaceID string, opts otf.ConfigurationVersionListOptions) (*otf.ConfigurationVersionList, error) {
	q := pggen.NewQuerier(db.Pool)
	batch := &pgx.Batch{}
	ctx := context.Background()

	q.FindConfigurationVersionsByWorkspaceIDBatch(batch, pggen.FindConfigurationVersionsByWorkspaceIDParams{
		WorkspaceID: pgtype.Text{String: workspaceID, Status: pgtype.Present},
		Limit:       opts.GetLimit(),
		Offset:      opts.GetOffset(),
	})
	q.CountConfigurationVersionsByWorkspaceIDBatch(batch, pgtype.Text{String: workspaceID, Status: pgtype.Present})

	results := db.Pool.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := q.FindConfigurationVersionsByWorkspaceIDScan(results)
	if err != nil {
		return nil, err
	}
	count, err := q.CountConfigurationVersionsByWorkspaceIDScan(results)
	if err != nil {
		return nil, err
	}

	var items []*otf.ConfigurationVersion
	for _, r := range rows {
		cv, err := otf.UnmarshalConfigurationVersionDBResult(otf.ConfigurationVersionDBResult(r))
		if err != nil {
			return nil, err
		}
		items = append(items, cv)
	}

	return &otf.ConfigurationVersionList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, *count),
	}, nil
}

func (db ConfigurationVersionDB) Get(opts otf.ConfigurationVersionGetOptions) (*otf.ConfigurationVersion, error) {
	ctx := context.Background()
	q := pggen.NewQuerier(db.Pool)

	if opts.ID != nil {
		result, err := q.FindConfigurationVersionByID(ctx,
			includeWorkspace(opts.Include),
			pgtype.Text{String: *opts.ID, Status: pgtype.Present},
		)
		if err != nil {
			return nil, err
		}
		return otf.UnmarshalConfigurationVersionDBResult(otf.ConfigurationVersionDBResult(result))
	} else if opts.WorkspaceID != nil {
		result, err := q.FindConfigurationVersionLatestByWorkspaceID(ctx,
			includeWorkspace(opts.Include),
			pgtype.Text{String: *opts.WorkspaceID, Status: pgtype.Present},
		)
		if err != nil {
			return nil, err
		}
		return otf.UnmarshalConfigurationVersionDBResult(otf.ConfigurationVersionDBResult(result))
	} else {
		return nil, fmt.Errorf("no configuration version spec provided")
	}
}

func (db ConfigurationVersionDB) GetConfig(ctx context.Context, id string) ([]byte, error) {
	q := pggen.NewQuerier(db.Pool)

	return q.DownloadConfigurationVersion(ctx, pgtype.Text{String: id, Status: pgtype.Present})
}

// Delete deletes a configuration version from the DB
func (db ConfigurationVersionDB) Delete(id string) error {
	q := pggen.NewQuerier(db.Pool)
	ctx := context.Background()

	result, err := q.DeleteConfigurationVersionByID(ctx, pgtype.Text{String: id, Status: pgtype.Present})
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return nil
}
