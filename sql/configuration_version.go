package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

func (db *DB) CreateConfigurationVersion(ctx context.Context, cv *otf.ConfigurationVersion) error {
	return db.tx(ctx, func(tx *DB) error {
		_, err := tx.InsertConfigurationVersion(ctx, pggen.InsertConfigurationVersionParams{
			ID:            String(cv.ID()),
			CreatedAt:     Timestamptz(cv.CreatedAt()),
			AutoQueueRuns: cv.AutoQueueRuns(),
			Source:        String(string(cv.Source())),
			Speculative:   cv.Speculative(),
			Status:        String(string(cv.Status())),
			WorkspaceID:   String(cv.WorkspaceID()),
		})
		if err != nil {
			return err
		}

		if cv.IngressAttributes() != nil {
			ia := cv.IngressAttributes()
			_, err := tx.InsertIngressAttributes(ctx, pggen.InsertIngressAttributesParams{
				Branch:                 String(ia.Branch),
				CommitSHA:              String(ia.CommitSHA),
				Identifier:             String(ia.Identifier),
				IsPullRequest:          ia.IsPullRequest,
				OnDefaultBranch:        ia.OnDefaultBranch,
				ConfigurationVersionID: String(cv.ID()),
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

func (db *DB) UploadConfigurationVersion(ctx context.Context, id string, fn func(*otf.ConfigurationVersion, otf.ConfigUploader) error) error {
	return db.tx(ctx, func(tx *DB) error {
		// select ...for update
		result, err := tx.FindConfigurationVersionByIDForUpdate(ctx, String(id))
		if err != nil {
			return err
		}
		cv, err := otf.UnmarshalConfigurationVersionResult(otf.ConfigurationVersionResult(result))
		if err != nil {
			return err
		}

		if err := fn(cv, newConfigUploader(tx, cv.ID())); err != nil {
			return err
		}
		return nil
	})
}

func (db *DB) ListConfigurationVersions(ctx context.Context, workspaceID string, opts otf.ConfigurationVersionListOptions) (*otf.ConfigurationVersionList, error) {
	batch := &pgx.Batch{}
	db.FindConfigurationVersionsByWorkspaceIDBatch(batch, pggen.FindConfigurationVersionsByWorkspaceIDParams{
		WorkspaceID: String(workspaceID),
		Limit:       opts.GetLimit(),
		Offset:      opts.GetOffset(),
	})
	db.CountConfigurationVersionsByWorkspaceIDBatch(batch, String(workspaceID))
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
		cv, err := otf.UnmarshalConfigurationVersionResult(otf.ConfigurationVersionResult(r))
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
		result, err := db.FindConfigurationVersionByID(ctx, String(*opts.ID))
		if err != nil {
			return nil, Error(err)
		}
		return otf.UnmarshalConfigurationVersionResult(otf.ConfigurationVersionResult(result))
	} else if opts.WorkspaceID != nil {
		result, err := db.FindConfigurationVersionLatestByWorkspaceID(ctx, String(*opts.WorkspaceID))
		if err != nil {
			return nil, Error(err)
		}
		return otf.UnmarshalConfigurationVersionResult(otf.ConfigurationVersionResult(result))
	} else {
		return nil, fmt.Errorf("no configuration version spec provided")
	}
}

func (db *DB) GetConfig(ctx context.Context, id string) ([]byte, error) {
	return db.DownloadConfigurationVersion(ctx, String(id))
}

func (db *DB) DeleteConfigurationVersion(ctx context.Context, id string) error {
	_, err := db.DeleteConfigurationVersionByID(ctx, String(id))
	if err != nil {
		return Error(err)
	}
	return nil
}

func (db *DB) insertCVStatusTimestamp(ctx context.Context, cv *otf.ConfigurationVersion) error {
	sts, err := cv.StatusTimestamp(cv.Status())
	if err != nil {
		return err
	}
	_, err = db.InsertConfigurationVersionStatusTimestamp(ctx, pggen.InsertConfigurationVersionStatusTimestampParams{
		ID:        String(cv.ID()),
		Status:    String(string(cv.Status())),
		Timestamp: Timestamptz(sts),
	})
	return err
}
