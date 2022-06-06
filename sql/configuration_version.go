package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

func (db *DB) CreateConfigurationVersion(ctx context.Context, cv *otf.ConfigurationVersion) error {
	tx, err := db.Begin(ctx)
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

func (db *DB) UploadConfigurationVersion(ctx context.Context, id string, fn func(*otf.ConfigurationVersion, otf.ConfigUploader) error) error {
	tx, err := db.Begin(ctx)
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

func (db *DB) ListConfigurationVersions(ctx context.Context, workspaceID string, opts otf.ConfigurationVersionListOptions) (*otf.ConfigurationVersionList, error) {
	batch := &pgx.Batch{}
	db.FindConfigurationVersionsByWorkspaceIDBatch(batch, pggen.FindConfigurationVersionsByWorkspaceIDParams{
		WorkspaceID: pgtype.Text{String: workspaceID, Status: pgtype.Present},
		Limit:       opts.GetLimit(),
		Offset:      opts.GetOffset(),
	})
	db.CountConfigurationVersionsByWorkspaceIDBatch(batch, pgtype.Text{String: workspaceID, Status: pgtype.Present})
	results := db.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := db.FindConfigurationVersionsByWorkspaceIDScan(results)
	if err != nil {
		return nil, err
	}
	count, err := db.CountConfigurationVersionsByWorkspaceIDScan(results)
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

func (db *DB) GetConfigurationVersion(ctx context.Context, opts otf.ConfigurationVersionGetOptions) (*otf.ConfigurationVersion, error) {
	if opts.ID != nil {
		result, err := db.FindConfigurationVersionByID(ctx,
			includeWorkspace(opts.Include),
			pgtype.Text{String: *opts.ID, Status: pgtype.Present},
		)
		if err != nil {
			return nil, err
		}
		return otf.UnmarshalConfigurationVersionDBResult(otf.ConfigurationVersionDBResult(result))
	} else if opts.WorkspaceID != nil {
		result, err := db.FindConfigurationVersionLatestByWorkspaceID(ctx,
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

func (db *DB) GetConfig(ctx context.Context, id string) ([]byte, error) {
	return db.DownloadConfigurationVersion(ctx, pgtype.Text{String: id, Status: pgtype.Present})
}

func (db *DB) DeleteConfigurationVersion(ctx context.Context, id string) error {
	result, err := db.DeleteConfigurationVersionByID(ctx, pgtype.Text{String: id, Status: pgtype.Present})
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}
	return nil
}
