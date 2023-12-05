package configversion

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

type pgdb struct {
	*sql.DB // provides access to generated SQL queries
}

func (db *pgdb) CreateConfigurationVersion(ctx context.Context, cv *ConfigurationVersion) error {
	return db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.InsertConfigurationVersion(ctx, pggen.InsertConfigurationVersionParams{
			ID:            sql.String(cv.ID),
			CreatedAt:     sql.Timestamptz(cv.CreatedAt),
			AutoQueueRuns: sql.Bool(cv.AutoQueueRuns),
			Source:        sql.String(string(cv.Source)),
			Speculative:   sql.Bool(cv.Speculative),
			Status:        sql.String(string(cv.Status)),
			WorkspaceID:   sql.String(cv.WorkspaceID),
		})
		if err != nil {
			return err
		}

		if cv.IngressAttributes != nil {
			ia := cv.IngressAttributes
			_, err := q.InsertIngressAttributes(ctx, pggen.InsertIngressAttributesParams{
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
				ConfigurationVersionID: sql.String(cv.ID),
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

func (db *pgdb) UploadConfigurationVersion(ctx context.Context, id string, fn func(*ConfigurationVersion, ConfigUploader) error) error {
	return db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		// select ...for update
		result, err := q.FindConfigurationVersionByIDForUpdate(ctx, sql.String(id))
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

func (db *pgdb) ListConfigurationVersions(ctx context.Context, workspaceID string, opts ConfigurationVersionListOptions) (*resource.Page[*ConfigurationVersion], error) {
	q := db.Conn(ctx)
	batch := &pgx.Batch{}
	q.FindConfigurationVersionsByWorkspaceIDBatch(batch, pggen.FindConfigurationVersionsByWorkspaceIDParams{
		WorkspaceID: sql.String(workspaceID),
		Limit:       opts.GetLimit(),
		Offset:      opts.GetOffset(),
	})
	q.CountConfigurationVersionsByWorkspaceIDBatch(batch, sql.String(workspaceID))
	results := db.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := q.FindConfigurationVersionsByWorkspaceIDScan(results)
	if err != nil {
		return nil, err
	}
	count, err := q.CountConfigurationVersionsByWorkspaceIDScan(results)
	if err != nil {
		return nil, err
	}

	items := make([]*ConfigurationVersion, len(rows))
	for i, r := range rows {
		items[i] = pgRow(r).toConfigVersion()
	}
	return resource.NewPage(items, opts.PageOptions, internal.Int64(count.Int)), nil
}

func (db *pgdb) GetConfigurationVersion(ctx context.Context, opts ConfigurationVersionGetOptions) (*ConfigurationVersion, error) {
	q := db.Conn(ctx)
	if opts.ID != nil {
		result, err := q.FindConfigurationVersionByID(ctx, sql.String(*opts.ID))
		if err != nil {
			return nil, sql.Error(err)
		}
		return pgRow(result).toConfigVersion(), nil
	} else if opts.WorkspaceID != nil {
		result, err := q.FindConfigurationVersionLatestByWorkspaceID(ctx, sql.String(*opts.WorkspaceID))
		if err != nil {
			return nil, sql.Error(err)
		}
		return pgRow(result).toConfigVersion(), nil
	} else {
		return nil, fmt.Errorf("no configuration version spec provided")
	}
}

func (db *pgdb) GetConfig(ctx context.Context, id string) ([]byte, error) {
	cfg, err := db.Conn(ctx).DownloadConfigurationVersion(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return cfg, nil
}

func (db *pgdb) DeleteConfigurationVersion(ctx context.Context, id string) error {
	_, err := db.Conn(ctx).DeleteConfigurationVersionByID(ctx, sql.String(id))
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
	_, err = db.Conn(ctx).InsertConfigurationVersionStatusTimestamp(ctx, pggen.InsertConfigurationVersionStatusTimestampParams{
		ID:        sql.String(cv.ID),
		Status:    sql.String(string(cv.Status)),
		Timestamp: sql.Timestamptz(sts),
	})
	return err
}

// pgRow represents the result of a database query for a configuration version.
type pgRow struct {
	ConfigurationVersionID               pgtype.Text                                  `json:"configuration_version_id"`
	CreatedAt                            pgtype.Timestamptz                           `json:"created_at"`
	AutoQueueRuns                        pgtype.Bool                                  `json:"auto_queue_runs"`
	Source                               pgtype.Text                                  `json:"source"`
	Speculative                          pgtype.Bool                                  `json:"speculative"`
	Status                               pgtype.Text                                  `json:"status"`
	WorkspaceID                          pgtype.Text                                  `json:"workspace_id"`
	ConfigurationVersionStatusTimestamps []pggen.ConfigurationVersionStatusTimestamps `json:"configuration_version_status_timestamps"`
	IngressAttributes                    *pggen.IngressAttributes                     `json:"ingress_attributes"`
}

func (result pgRow) toConfigVersion() *ConfigurationVersion {
	cv := ConfigurationVersion{
		ID:               result.ConfigurationVersionID.String,
		CreatedAt:        result.CreatedAt.Time.UTC(),
		AutoQueueRuns:    result.AutoQueueRuns.Bool,
		Speculative:      result.Speculative.Bool,
		Source:           Source(result.Source.String),
		Status:           ConfigurationStatus(result.Status.String),
		StatusTimestamps: unmarshalStatusTimestampRows(result.ConfigurationVersionStatusTimestamps),
		WorkspaceID:      result.WorkspaceID.String,
	}
	if result.IngressAttributes != nil {
		cv.IngressAttributes = NewIngressFromRow(result.IngressAttributes)
	}
	return &cv
}

func NewIngressFromRow(row *pggen.IngressAttributes) *IngressAttributes {
	return &IngressAttributes{
		Branch:            row.Branch.String,
		CommitSHA:         row.CommitSHA.String,
		CommitURL:         row.CommitURL.String,
		Repo:              row.Identifier.String,
		IsPullRequest:     row.IsPullRequest.Bool,
		PullRequestNumber: int(row.PullRequestNumber.Int),
		PullRequestURL:    row.PullRequestURL.String,
		PullRequestTitle:  row.PullRequestTitle.String,
		SenderUsername:    row.SenderUsername.String,
		SenderAvatarURL:   row.SenderAvatarURL.String,
		SenderHTMLURL:     row.SenderHTMLURL.String,
		Tag:               row.Tag.String,
		OnDefaultBranch:   row.IsPullRequest.Bool,
	}
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
