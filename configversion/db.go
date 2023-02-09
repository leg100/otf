package configversion

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

type db interface {
	otf.Database

	// Creates a config version.
	CreateConfigurationVersion(ctx context.Context, cv *ConfigurationVersion) error
	// Get retrieves a config version.
	GetConfigurationVersion(ctx context.Context, opts ConfigurationVersionGetOptions) (*ConfigurationVersion, error)
	// GetConfig retrieves the config tarball for the given config version ID.
	GetConfig(ctx context.Context, id string) ([]byte, error)
	// List lists config versions for the given workspace.
	ListConfigurationVersions(ctx context.Context, workspaceID string, opts ConfigurationVersionListOptions) (*ConfigurationVersionList, error)
	// Delete deletes the config version from the store
	DeleteConfigurationVersion(ctx context.Context, id string) error
	// Upload uploads a config tarball for the given config version ID
	UploadConfigurationVersion(ctx context.Context, id string, fn func(cv *ConfigurationVersion, uploader ConfigUploader) error) error

	insertCVStatusTimestamp(ctx context.Context, cv *ConfigurationVersion) error
	tx(context.Context, func(db) error) error
}

type DB struct {
	otf.Database // provides access to generated SQL queries
}

func newPGDB(db otf.Database) *DB {
	return &DB{db}
}

func (pdb *DB) CreateConfigurationVersion(ctx context.Context, cv *ConfigurationVersion) error {
	return pdb.tx(ctx, func(tx db) error {
		_, err := tx.InsertConfigurationVersion(ctx, pggen.InsertConfigurationVersionParams{
			ID:            sql.String(cv.ID()),
			CreatedAt:     sql.Timestamptz(cv.CreatedAt()),
			AutoQueueRuns: cv.AutoQueueRuns(),
			Source:        sql.String(string(cv.Source())),
			Speculative:   cv.Speculative(),
			Status:        sql.String(string(cv.Status())),
			WorkspaceID:   sql.String(cv.WorkspaceID()),
		})
		if err != nil {
			return err
		}

		if cv.IngressAttributes() != nil {
			ia := cv.IngressAttributes()
			_, err := tx.InsertIngressAttributes(ctx, pggen.InsertIngressAttributesParams{
				Branch:                 sql.String(ia.Branch),
				CommitSHA:              sql.String(ia.CommitSHA),
				Identifier:             sql.String(ia.Identifier),
				IsPullRequest:          ia.IsPullRequest,
				OnDefaultBranch:        ia.OnDefaultBranch,
				ConfigurationVersionID: sql.String(cv.ID()),
			})
			if err != nil {
				return err
			}
		}

		// Insert timestamp for current status
		if err := tx.insertCVStatusTimestamp(ctx, cv); err != nil {
			return fmt.Errorf("inserting configuration version status timestamp: %w", err)
		}
		return nil
	})
}

func (pdb *DB) UploadConfigurationVersion(ctx context.Context, id string, fn func(*ConfigurationVersion, ConfigUploader) error) error {
	return pdb.tx(ctx, func(tx db) error {
		// select ...for update
		result, err := tx.FindConfigurationVersionByIDForUpdate(ctx, sql.String(id))
		if err != nil {
			return err
		}
		cv := pgRow(result).toConfigVersion()

		if err := fn(cv, newConfigUploader(tx, cv.ID())); err != nil {
			return err
		}
		return nil
	})
}

func (db *DB) ListConfigurationVersions(ctx context.Context, workspaceID string, opts ConfigurationVersionListOptions) (*ConfigurationVersionList, error) {
	batch := &pgx.Batch{}
	db.FindConfigurationVersionsByWorkspaceIDBatch(batch, pggen.FindConfigurationVersionsByWorkspaceIDParams{
		WorkspaceID: sql.String(workspaceID),
		Limit:       opts.GetLimit(),
		Offset:      opts.GetOffset(),
	})
	db.CountConfigurationVersionsByWorkspaceIDBatch(batch, sql.String(workspaceID))
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

	var items []*ConfigurationVersion
	for _, r := range rows {
		items = append(items, pgRow(r).toConfigVersion())
	}

	return &ConfigurationVersionList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, *count),
	}, nil
}

func (db *DB) GetConfigurationVersion(ctx context.Context, opts ConfigurationVersionGetOptions) (*ConfigurationVersion, error) {
	if opts.ID != nil {
		result, err := db.FindConfigurationVersionByID(ctx, sql.String(*opts.ID))
		if err != nil {
			return nil, sql.Error(err)
		}
		return pgRow(result).toConfigVersion(), nil
	} else if opts.WorkspaceID != nil {
		result, err := db.FindConfigurationVersionLatestByWorkspaceID(ctx, sql.String(*opts.WorkspaceID))
		if err != nil {
			return nil, sql.Error(err)
		}
		return pgRow(result).toConfigVersion(), nil
	} else {
		return nil, fmt.Errorf("no configuration version spec provided")
	}
}

func (db *DB) GetConfig(ctx context.Context, id string) ([]byte, error) {
	return db.DownloadConfigurationVersion(ctx, sql.String(id))
}

func (db *DB) DeleteConfigurationVersion(ctx context.Context, id string) error {
	_, err := db.DeleteConfigurationVersionByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *DB) insertCVStatusTimestamp(ctx context.Context, cv *ConfigurationVersion) error {
	sts, err := cv.StatusTimestamp(cv.Status())
	if err != nil {
		return err
	}
	_, err = db.InsertConfigurationVersionStatusTimestamp(ctx, pggen.InsertConfigurationVersionStatusTimestampParams{
		ID:        sql.String(cv.ID()),
		Status:    sql.String(string(cv.Status())),
		Timestamp: sql.Timestamptz(sts),
	})
	return err
}

// tx constructs a new pgdb within a transaction.
func (db *DB) tx(ctx context.Context, callback func(db) error) error {
	return db.Transaction(ctx, func(tx otf.Database) error {
		return callback(newPGDB(tx))
	})
}

// pgRow represents the result of a database query for a
// configuration version.
type pgRow struct {
	ConfigurationVersionID               pgtype.Text                                  `json:"configuration_version_id"`
	CreatedAt                            pgtype.Timestamptz                           `json:"created_at"`
	AutoQueueRuns                        bool                                         `json:"auto_queue_runs"`
	Source                               pgtype.Text                                  `json:"source"`
	Speculative                          bool                                         `json:"speculative"`
	Status                               pgtype.Text                                  `json:"status"`
	WorkspaceID                          pgtype.Text                                  `json:"workspace_id"`
	ConfigurationVersionStatusTimestamps []pggen.ConfigurationVersionStatusTimestamps `json:"configuration_version_status_timestamps"`
	IngressAttributes                    *pggen.IngressAttributes                     `json:"ingress_attributes"`
}

func (result pgRow) toConfigVersion() *ConfigurationVersion {
	cv := ConfigurationVersion{
		id:               result.ConfigurationVersionID.String,
		createdAt:        result.CreatedAt.Time.UTC(),
		autoQueueRuns:    result.AutoQueueRuns,
		speculative:      result.Speculative,
		source:           ConfigurationSource(result.Source.String),
		status:           ConfigurationStatus(result.Status.String),
		statusTimestamps: unmarshalStatusTimestampRows(result.ConfigurationVersionStatusTimestamps),
		workspaceID:      result.WorkspaceID.String,
	}
	if result.IngressAttributes != nil {
		cv.ingressAttributes = &otf.IngressAttributes{
			Branch:          result.IngressAttributes.Branch.String,
			CommitSHA:       result.IngressAttributes.CommitSHA.String,
			Identifier:      result.IngressAttributes.Identifier.String,
			IsPullRequest:   result.IngressAttributes.IsPullRequest,
			OnDefaultBranch: result.IngressAttributes.IsPullRequest,
		}
	}
	return &cv
}

func unmarshalStatusTimestampRows(rows []pggen.ConfigurationVersionStatusTimestamps) (timestamps []ConfigurationVersionStatusTimestamp) {
	for _, ty := range rows {
		timestamps = append(timestamps, ConfigurationVersionStatusTimestamp{
			Status:    ConfigurationStatus(ty.Status.String),
			Timestamp: ty.Timestamp.Time.UTC(),
		})
	}
	return timestamps
}
