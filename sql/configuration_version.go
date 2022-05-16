package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leg100/otf"
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

func (db ConfigurationVersionDB) Create(cv *otf.ConfigurationVersion) (*otf.ConfigurationVersion, error) {
	ctx := context.Background()

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	q := NewQuerier(tx)

	result, err := q.InsertConfigurationVersion(ctx, InsertConfigurationVersionParams{
		ID:            cv.ID,
		AutoQueueRuns: cv.AutoQueueRuns,
		Source:        string(cv.Source),
		Speculative:   cv.Speculative,
		Status:        string(cv.Status),
		WorkspaceID:   cv.Workspace.ID,
	})
	if err != nil {
		return nil, err
	}
	cv.CreatedAt = result.CreatedAt
	cv.UpdatedAt = result.UpdatedAt

	// Insert timestamp for current status
	ts, err := q.InsertConfigurationVersionStatusTimestamp(ctx, cv.ID, string(cv.Status))
	if err != nil {
		return nil, err
	}
	cv.StatusTimestamps = append(cv.StatusTimestamps, otf.ConfigurationVersionStatusTimestamp{
		Status:    otf.ConfigurationStatus(ts.Status),
		Timestamp: ts.Timestamp,
	})

	return cv, tx.Commit(ctx)
}

func (db ConfigurationVersionDB) Upload(ctx context.Context, id string, fn func(*otf.ConfigurationVersion, otf.ConfigUploader) error) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := NewQuerier(tx)

	// select ...for update
	result, err := q.FindConfigurationVersionByIDForUpdate(ctx, id)
	if err != nil {
		return err
	}
	cv, err := otf.UnmarshalConfigurationVersionFromDB(result)
	if err != nil {
		return err
	}

	if err := fn(cv, newConfigUploader(tx, cv.ID)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (db ConfigurationVersionDB) List(workspaceID string, opts otf.ConfigurationVersionListOptions) (*otf.ConfigurationVersionList, error) {
	q := NewQuerier(db.Pool)
	batch := &pgx.Batch{}
	ctx := context.Background()

	q.FindConfigurationVersionsByWorkspaceIDBatch(batch, FindConfigurationVersionsByWorkspaceIDParams{
		WorkspaceID: workspaceID,
		Limit:       opts.GetLimit(),
		Offset:      opts.GetOffset(),
	})
	q.CountConfigurationVersionsByWorkspaceIDBatch(batch, workspaceID)

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

	cvs, err := otf.UnmarshalConfigurationVersionListFromDB(rows)
	if err != nil {
		return nil, err
	}

	return &otf.ConfigurationVersionList{
		Items:      cvs,
		Pagination: otf.NewPagination(opts.ListOptions, *count),
	}, nil
}

func (db ConfigurationVersionDB) Get(opts otf.ConfigurationVersionGetOptions) (*otf.ConfigurationVersion, error) {
	ctx := context.Background()
	q := NewQuerier(db.Pool)

	if opts.ID != nil {
		result, err := q.FindConfigurationVersionByID(ctx, *opts.ID)
		if err != nil {
			return nil, err
		}
		return otf.UnmarshalConfigurationVersionFromDB(result)
	} else if opts.WorkspaceID != nil {
		result, err := q.FindConfigurationVersionLatestByWorkspaceID(ctx, *opts.WorkspaceID)
		if err != nil {
			return nil, err
		}
		return otf.UnmarshalConfigurationVersionFromDB(result)
	} else {
		return nil, fmt.Errorf("no configuration version spec provided")
	}
}

func (db ConfigurationVersionDB) GetConfig(ctx context.Context, id string) ([]byte, error) {
	q := NewQuerier(db.Pool)

	return q.DownloadConfigurationVersion(ctx, id)
}

// Delete deletes a configuration version from the DB
func (db ConfigurationVersionDB) Delete(id string) error {
	q := NewQuerier(db.Pool)
	ctx := context.Background()

	result, err := q.DeleteConfigurationVersionByID(ctx, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return nil
}
