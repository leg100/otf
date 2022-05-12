package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
)

var (
	_ otf.ConfigurationVersionStore = (*ConfigurationVersionDB)(nil)
)

type ConfigurationVersionDB struct {
	*pgx.Conn
}

func NewConfigurationVersionDB(conn *pgx.Conn) *ConfigurationVersionDB {
	return &ConfigurationVersionDB{
		Conn: conn,
	}
}

func (db ConfigurationVersionDB) Create(cv *otf.ConfigurationVersion) (*otf.ConfigurationVersion, error) {
	ctx := context.Background()

	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	q := NewQuerier(tx)

	result, err := q.InsertConfigurationVersion(ctx, InsertConfigurationVersionParams{
		ID:            &cv.ID,
		AutoQueueRuns: &cv.AutoQueueRuns,
		Source:        otf.String(string(cv.Source)),
		Speculative:   &cv.Speculative,
		Status:        otf.String(string(cv.Status)),
		WorkspaceID:   &cv.Workspace.ID,
	})
	if err != nil {
		return nil, err
	}
	addResultToConfigurationVersion(cv, result)

	// Insert timestamp for current status
	ts, err := q.InsertConfigurationVersionStatusTimestamp(ctx, &cv.ID, otf.String(string(cv.Status)))
	if err != nil {
		return nil, err
	}
	cv.StatusTimestamps = append(cv.StatusTimestamps, convertConfigurationVersionStatusTimestamps(ts))

	return cv, tx.Commit(ctx)
}

func (db ConfigurationVersionDB) Update(id string, fn func(*otf.ConfigurationVersion, otf.ConfigurationVersionUpdater) error) error {
	ctx := context.Background()

	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := NewQuerier(tx)

	// select ...for update
	result, err := q.FindConfigurationVersionByIDForUpdate(ctx, &id)
	if err != nil {
		return err
	}
	cv := convertConfigurationVersion(result)

	if err := fn(cv, newConfigurationVersionUpdater(tx, cv.ID)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (db ConfigurationVersionDB) List(workspaceID string, opts otf.ConfigurationVersionListOptions) (*otf.ConfigurationVersionList, error) {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	result, err := q.FindConfigurationVersionsByWorkspaceID(ctx, FindConfigurationVersionsByWorkspaceIDParams{
		WorkspaceID: opts.WorkspaceID,
		Limit:       opts.GetLimit(),
		Offset:      opts.GetOffset(),
	})
	if err != nil {
		return nil, err
	}

	var items []*otf.ConfigurationVersion
	for _, r := range result {
		items = append(items, convertConfigurationVersion(r))
	}

	return &otf.ConfigurationVersionList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, getCount(result)),
	}, nil
}

func (db ConfigurationVersionDB) Get(opts otf.ConfigurationVersionGetOptions) (*otf.ConfigurationVersion, error) {
	return getConfigurationVersion(context.Background(), NewQuerier(db.Conn), opts)
}

func (db ConfigurationVersionDB) GetConfig(ctx context.Context, id string) ([]byte, error) {
	q := NewQuerier(db.Conn)

	return q.DownloadConfigurationVersion(ctx, &id)
}

// Delete deletes a configuration version from the DB
func (db ConfigurationVersionDB) Delete(id string) error {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	result, err := q.DeleteConfigurationVersionByID(ctx, &id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return nil
}

func getConfigurationVersion(ctx context.Context, q *DBQuerier, opts otf.ConfigurationVersionGetOptions) (*otf.ConfigurationVersion, error) {
	if opts.ID != nil {
		result, err := q.FindConfigurationVersionByID(ctx, opts.ID)
		if err != nil {
			return nil, err
		}
		return convertConfigurationVersion(result), nil
	} else if opts.WorkspaceID != nil {
		result, err := q.FindConfigurationVersionLatestByWorkspaceID(ctx, opts.WorkspaceID)
		if err != nil {
			return nil, err
		}
		return convertConfigurationVersion(result), nil
	} else {
		return nil, fmt.Errorf("no configuration version spec provided")
	}
}
