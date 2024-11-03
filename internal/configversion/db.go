package configversion

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
)

type pgdb struct {
	*sql.DB // provides access to generated SQL queries
}

func (db *pgdb) CreateConfigurationVersion(ctx context.Context, cv *ConfigurationVersion) error {
	return db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		err := q.InsertConfigurationVersion(ctx, sqlc.InsertConfigurationVersionParams{
			ID:            sql.ID(cv.ID),
			CreatedAt:     sql.Timestamptz(cv.CreatedAt),
			AutoQueueRuns: sql.Bool(cv.AutoQueueRuns),
			Source:        sql.String(string(cv.Source)),
			Speculative:   sql.Bool(cv.Speculative),
			Status:        sql.String(string(cv.Status)),
			WorkspaceID:   sql.ID(cv.WorkspaceID),
		})
		if err != nil {
			return err
		}

		if cv.IngressAttributes != nil {
			ia := cv.IngressAttributes
			err := q.InsertIngressAttributes(ctx, sqlc.InsertIngressAttributesParams{
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
				ConfigurationVersionID: sql.ID(cv.ID),
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
	return db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		// select ...for update
		result, err := q.FindConfigurationVersionByIDForUpdate(ctx, sql.ID(id))
		if err != nil {
			return err
		}
		cv := pgRow(result).toConfigVersion()

		if err := fn(cv, newConfigUploader(q, cv.ID)); err != nil {
			return err
		}
		return nil
	})
}

func (db *pgdb) ListConfigurationVersions(ctx context.Context, workspaceID resource.ID, opts ListOptions) (*resource.Page[*ConfigurationVersion], error) {
	q := db.Querier(ctx)
	rows, err := q.FindConfigurationVersionsByWorkspaceID(ctx, sqlc.FindConfigurationVersionsByWorkspaceIDParams{
		WorkspaceID: sql.ID(workspaceID),
		Limit:       sql.GetLimit(opts.PageOptions),
		Offset:      sql.GetOffset(opts.PageOptions),
	})
	if err != nil {
		return nil, err
	}
	count, err := q.CountConfigurationVersionsByWorkspaceID(ctx, sql.ID(workspaceID))
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
	q := db.Querier(ctx)
	if opts.ID != nil {
		result, err := q.FindConfigurationVersionByID(ctx, sql.ID(opts.ID))
		if err != nil {
			return nil, sql.Error(err)
		}
		return pgRow(result).toConfigVersion(), nil
	} else if opts.WorkspaceID != nil {
		result, err := q.FindConfigurationVersionLatestByWorkspaceID(ctx, sql.ID(opts.WorkspaceID))
		if err != nil {
			return nil, sql.Error(err)
		}
		return pgRow(result).toConfigVersion(), nil
	} else {
		return nil, fmt.Errorf("no configuration version spec provided")
	}
}

func (db *pgdb) GetConfig(ctx context.Context, id resource.ID) ([]byte, error) {
	cfg, err := db.Querier(ctx).DownloadConfigurationVersion(ctx, sql.ID(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return cfg, nil
}

func (db *pgdb) DeleteConfigurationVersion(ctx context.Context, id resource.ID) error {
	_, err := db.Querier(ctx).DeleteConfigurationVersionByID(ctx, sql.ID(id))
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
	_, err = db.Querier(ctx).InsertConfigurationVersionStatusTimestamp(ctx, sqlc.InsertConfigurationVersionStatusTimestampParams{
		ID:        sql.ID(cv.ID),
		Status:    sql.String(string(cv.Status)),
		Timestamp: sql.Timestamptz(sts),
	})
	return err
}

// pgRow represents the result of a database query for a configuration version.
type pgRow struct {
	ConfigurationVersionID pgtype.Text
	CreatedAt              pgtype.Timestamptz
	AutoQueueRuns          pgtype.Bool
	Source                 pgtype.Text
	Speculative            pgtype.Bool
	Status                 pgtype.Text
	WorkspaceID            pgtype.Text
	StatusTimestamps       []sqlc.ConfigurationVersionStatusTimestamp
	IngressAttributes      *sqlc.IngressAttribute
}

func (result pgRow) toConfigVersion() *ConfigurationVersion {
	cv := ConfigurationVersion{
		ID:               resource.ID{Kind: ConfigVersionKind, ID: result.ConfigurationVersionID.String},
		CreatedAt:        result.CreatedAt.Time.UTC(),
		AutoQueueRuns:    result.AutoQueueRuns.Bool,
		Speculative:      result.Speculative.Bool,
		Source:           Source(result.Source.String),
		Status:           ConfigurationStatus(result.Status.String),
		StatusTimestamps: unmarshalStatusTimestampRows(result.StatusTimestamps),
		WorkspaceID:      resource.ID{Kind: ConfigVersionKind, ID: result.WorkspaceID.String},
	}
	if result.IngressAttributes != nil {
		cv.IngressAttributes = NewIngressFromRow(result.IngressAttributes)
	}
	return &cv
}

func NewIngressFromRow(row *sqlc.IngressAttribute) *IngressAttributes {
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

func unmarshalStatusTimestampRows(rows []sqlc.ConfigurationVersionStatusTimestamp) (timestamps []ConfigurationVersionStatusTimestamp) {
	for _, ty := range rows {
		timestamps = append(timestamps, ConfigurationVersionStatusTimestamp{
			Status:    ConfigurationStatus(ty.Status.String),
			Timestamp: ty.Timestamp.Time.UTC(),
		})
	}
	return timestamps
}
