package sql

import (
	"context"
	"fmt"

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

func (db ConfigurationVersionDB) Create(cv *otf.ConfigurationVersion) (*otf.ConfigurationVersion, error) {
	ctx := context.Background()

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	q := pggen.NewQuerier(tx)

	result, err := q.InsertConfigurationVersion(ctx, pggen.InsertConfigurationVersionParams{
		ID:            cv.ID,
		AutoQueueRuns: cv.AutoQueueRuns(),
		Source:        string(cv.Source()),
		Speculative:   cv.Speculative(),
		Status:        string(cv.Status()),
		WorkspaceID:   cv.Workspace.ID,
	})
	if err != nil {
		return nil, err
	}
	cv.CreatedAt = result.CreatedAt
	cv.UpdatedAt = result.UpdatedAt

	// Insert timestamp for current status
	ts, err := q.InsertConfigurationVersionStatusTimestamp(ctx, cv.ID, string(cv.Status()))
	if err != nil {
		return nil, err
	}
	cv.AddStatusTimestamp(otf.ConfigurationStatus(ts.Status), ts.Timestamp)

	return cv, tx.Commit(ctx)
}

func (db ConfigurationVersionDB) Upload(ctx context.Context, id string, fn func(*otf.ConfigurationVersion, otf.ConfigUploader) error) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := pggen.NewQuerier(tx)

	// select ...for update
	result, err := q.FindConfigurationVersionByIDForUpdate(ctx, id)
	if err != nil {
		return err
	}
	cv, err := otf.UnmarshalConfigurationVersionDBResult(otf.ConfigurationVersionDBResult(result))
	if err != nil {
		return err
	}

	if err := fn(cv, newConfigUploader(tx, cv.ID)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (db ConfigurationVersionDB) List(workspaceID string, opts otf.ConfigurationVersionListOptions) (*otf.ConfigurationVersionList, error) {
	q := pggen.NewQuerier(db.Pool)
	batch := &pgx.Batch{}
	ctx := context.Background()

	q.FindConfigurationVersionsByWorkspaceIDBatch(batch, pggen.FindConfigurationVersionsByWorkspaceIDParams{
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
		result, err := q.FindConfigurationVersionByID(ctx, *opts.ID)
		if err != nil {
			return nil, err
		}
		return otf.UnmarshalConfigurationVersionDBResult(otf.ConfigurationVersionDBResult(result))
	} else if opts.WorkspaceID != nil {
		result, err := q.FindConfigurationVersionLatestByWorkspaceID(ctx, *opts.WorkspaceID)
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

	return q.DownloadConfigurationVersion(ctx, id)
}

// Delete deletes a configuration version from the DB
func (db ConfigurationVersionDB) Delete(id string) error {
	q := pggen.NewQuerier(db.Pool)
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
