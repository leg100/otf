package configversion

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

var q = &Queries{}

type pgdb struct {
	*sql.DB // provides access to generated SQL queries
}

func (db *pgdb) CreateConfigurationVersion(ctx context.Context, cv *ConfigurationVersion) error {
	return db.Tx(ctx, func(ctx context.Context, conn sql.Connection) error {
		err := q.InsertConfigurationVersion(ctx, conn, InsertConfigurationVersionParams{
			ID:            cv.ID,
			CreatedAt:     sql.Timestamptz(cv.CreatedAt),
			AutoQueueRuns: sql.Bool(cv.AutoQueueRuns),
			Source:        sql.String(string(cv.Source)),
			Speculative:   sql.Bool(cv.Speculative),
			Status:        sql.String(string(cv.Status)),
			WorkspaceID:   cv.WorkspaceID,
		})
		if err != nil {
			return err
		}

		if cv.IngressAttributes != nil {
			ia := cv.IngressAttributes
			err := q.InsertIngressAttributes(ctx, conn, InsertIngressAttributesParams{
				Branch:                 sql.String(ia.Branch),
				CommitSHA:              sql.String(ia.CommitSHA),
				CommitURL:              sql.String(ia.CommitURL),
				PullRequestNumber:      sql.Int4(ia.PullRequestNumber),
				PullRequestURL:         sql.String(ia.PullRequestURL),
				PullRequestTitle:       sql.String(ia.PullRequestTitle),
				SenderUsername:         sql.String(ia.SenderUsername),
				SenderAvatarURL:        sql.String(ia.SenderAvatarURL),
				SenderHTMLURL:          sql.String(ia.SenderHTMLURL),
				Tag:                    sql.String(ia.Tag),
				Identifier:             sql.String(ia.Repo),
				IsPullRequest:          sql.Bool(ia.IsPullRequest),
				OnDefaultBranch:        sql.Bool(ia.OnDefaultBranch),
				ConfigurationVersionID: cv.ID,
			})
			if err != nil {
				return err
			}
		}

		// Insert timestamp for current status
		if err := db.insertCVStatusTimestamp(ctx, cv); err != nil {
			return fmt.Errorf("inserting configuration version status timestamp: %w", err)
		}
		return nil
	})
}

func (db *pgdb) UploadConfigurationVersion(ctx context.Context, id resource.ID, fn func(*ConfigurationVersion, ConfigUploader) error) error {
	return db.Tx(ctx, func(ctx context.Context, conn sql.Connection) error {
		// select ...for update
		result, err := q.FindConfigurationVersionByIDForUpdate(ctx, conn, id)
		if err != nil {
			return err
		}
		cv := pgRow(result).toConfigVersion()

		if err := fn(cv, &cvUploader{conn, cv.ID}); err != nil {
			return err
		}
		return nil
	})
}

func (db *pgdb) ListConfigurationVersions(ctx context.Context, workspaceID resource.ID, opts ListOptions) (*resource.Page[*ConfigurationVersion], error) {
	rows, err := q.FindConfigurationVersionsByWorkspaceID(ctx, db.Conn(ctx), FindConfigurationVersionsByWorkspaceIDParams{
		WorkspaceID: workspaceID,
		Limit:       sql.GetLimit(opts.PageOptions),
		Offset:      sql.GetOffset(opts.PageOptions),
	})
	if err != nil {
		return nil, err
	}
	count, err := q.CountConfigurationVersionsByWorkspaceID(ctx, db.Conn(ctx), workspaceID)
	if err != nil {
		return nil, err
	}

	items := make([]*ConfigurationVersion, len(rows))
	for i, r := range rows {
		items[i] = pgRow(r).toConfigVersion()
	}
	return resource.NewPage(items, opts.PageOptions, internal.Int64(count)), nil
}

func (db *pgdb) GetConfigurationVersion(ctx context.Context, opts ConfigurationVersionGetOptions) (*ConfigurationVersion, error) {
	if opts.ID != nil {
		result, err := q.FindConfigurationVersionByID(ctx, db.Conn(ctx), *opts.ID)
		if err != nil {
			return nil, sql.Error(err)
		}
		return pgRow(result).toConfigVersion(), nil
	} else if opts.WorkspaceID != nil {
		result, err := q.FindConfigurationVersionLatestByWorkspaceID(ctx, db.Conn(ctx), *opts.WorkspaceID)
		if err != nil {
			return nil, sql.Error(err)
		}
		return pgRow(result).toConfigVersion(), nil
	} else {
		return nil, fmt.Errorf("no configuration version spec provided")
	}
}

func (db *pgdb) GetConfig(ctx context.Context, id resource.ID) ([]byte, error) {
	cfg, err := q.DownloadConfigurationVersion(ctx, db.Conn(ctx), id)
	if err != nil {
		return nil, sql.Error(err)
	}
	return cfg, nil
}

func (db *pgdb) DeleteConfigurationVersion(ctx context.Context, id resource.ID) error {
	_, err := q.DeleteConfigurationVersionByID(ctx, db.Conn(ctx), id)
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) insertCVStatusTimestamp(ctx context.Context, cv *ConfigurationVersion) error {
	sts, err := cv.StatusTimestamp(cv.Status)
	if err != nil {
		return err
	}
	_, err = q.InsertConfigurationVersionStatusTimestamp(ctx, db.Conn(ctx), InsertConfigurationVersionStatusTimestampParams{
		ID:        cv.ID,
		Status:    sql.String(string(cv.Status)),
		Timestamp: sql.Timestamptz(sts),
	})
	return err
}

// pgRow represents the result of a database query for a configuration version.
type pgRow struct {
	ConfigurationVersionID resource.ID
	CreatedAt              pgtype.Timestamptz
	AutoQueueRuns          pgtype.Bool
	Source                 pgtype.Text
	Speculative            pgtype.Bool
	Status                 pgtype.Text
	WorkspaceID            resource.ID
	StatusTimestamps       []StatusTimestampModel
	IngressAttributes      *IngressAttributeModel
}

func (result pgRow) toConfigVersion() *ConfigurationVersion {
	cv := ConfigurationVersion{
		ID:               result.ConfigurationVersionID,
		CreatedAt:        result.CreatedAt.Time.UTC(),
		AutoQueueRuns:    result.AutoQueueRuns.Bool,
		Speculative:      result.Speculative.Bool,
		Source:           Source(result.Source.String),
		Status:           ConfigurationStatus(result.Status.String),
		StatusTimestamps: unmarshalStatusTimestampRows(result.StatusTimestamps),
		WorkspaceID:      result.WorkspaceID,
	}
	if result.IngressAttributes != nil {
		cv.IngressAttributes = NewIngressFromRow(result.IngressAttributes)
	}
	return &cv
}

func NewIngressFromRow(row *IngressAttributeModel) *IngressAttributes {
	return &IngressAttributes{
		Branch:            row.Branch.String,
		CommitSHA:         row.CommitSHA.String,
		CommitURL:         row.CommitURL.String,
		Repo:              row.Identifier.String,
		IsPullRequest:     row.IsPullRequest.Bool,
		PullRequestNumber: int(row.PullRequestNumber.Int32),
		PullRequestURL:    row.PullRequestURL.String,
		PullRequestTitle:  row.PullRequestTitle.String,
		SenderUsername:    row.SenderUsername.String,
		SenderAvatarURL:   row.SenderAvatarURL.String,
		SenderHTMLURL:     row.SenderHTMLURL.String,
		Tag:               row.Tag.String,
		OnDefaultBranch:   row.IsPullRequest.Bool,
	}
}

func unmarshalStatusTimestampRows(rows []StatusTimestampModel) (timestamps []ConfigurationVersionStatusTimestamp) {
	for _, ty := range rows {
		timestamps = append(timestamps, ConfigurationVersionStatusTimestamp{
			Status:    ConfigurationStatus(ty.Status.String),
			Timestamp: ty.Timestamp.Time.UTC(),
		})
	}
	return timestamps
}
