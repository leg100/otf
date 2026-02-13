package configversion

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal/configversion/source"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/vcs"
)

type pgdb struct {
	*sql.DB // provides access to generated SQL queries
}

func (db *pgdb) CreateConfigurationVersion(ctx context.Context, cv *ConfigurationVersion) error {
	return db.Tx(ctx, func(ctx context.Context) error {
		_, err := db.Exec(ctx, `
INSERT INTO configuration_versions (
    configuration_version_id,
    created_at,
    auto_queue_runs,
    source,
    speculative,
    status,
    workspace_id
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
`,
			cv.ID,
			cv.CreatedAt,
			cv.AutoQueueRuns,
			cv.Source,
			cv.Speculative,
			cv.Status,
			cv.WorkspaceID,
		)
		if err != nil {
			return err
		}

		if cv.IngressAttributes != nil {
			ia := cv.IngressAttributes
			_, err := db.Exec(ctx, `
INSERT INTO ingress_attributes (
    branch,
    commit_sha,
    commit_url,
    pull_request_number,
    pull_request_url,
    pull_request_title,
    sender_username,
    sender_avatar_url,
    sender_html_url,
    identifier,
    tag,
    is_pull_request,
    on_default_branch,
    configuration_version_id
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12,
    $13,
    $14
)
`,
				ia.Branch,
				ia.CommitSHA,
				ia.CommitURL,
				ia.PullRequestNumber,
				ia.PullRequestURL,
				ia.PullRequestTitle,
				ia.SenderUsername,
				ia.SenderAvatarURL,
				ia.SenderHTMLURL,
				ia.Repo,
				ia.Tag,
				ia.IsPullRequest,
				ia.OnDefaultBranch,
				cv.ID,
			)
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

func (db *pgdb) UploadConfigurationVersion(ctx context.Context, id resource.TfeID, config []byte) error {
	_, err := db.Exec(ctx, `
UPDATE configuration_versions
SET
    config = @config,
    status = 'uploaded'
WHERE configuration_version_id = @id
`, pgx.NamedArgs{
		"config": config,
		"id":     id,
	})
	return err
}

func (db *pgdb) ListConfigurationVersions(ctx context.Context, workspaceID resource.TfeID, opts ListOptions) (*resource.Page[*ConfigurationVersion], error) {
	rows := db.Query(ctx, `
SELECT
    cv.configuration_version_id,
    cv.created_at,
    cv.auto_queue_runs,
    cv.source,
    cv.speculative,
    cv.status,
    cv.workspace_id,
    (
        SELECT array_agg(cst.*)::configuration_version_status_timestamps[]
        FROM configuration_version_status_timestamps cst
        WHERE cst.configuration_version_id = cv.configuration_version_id
        GROUP BY cst.configuration_version_id
    ) AS status_timestamps,
    ia::"ingress_attributes" AS ingress_attributes
FROM configuration_versions cv
JOIN workspaces USING (workspace_id)
LEFT JOIN ingress_attributes ia USING (configuration_version_id)
WHERE workspaces.workspace_id = $1
LIMIT $2::int
OFFSET $3::int
`,
		workspaceID,
		sql.GetLimit(opts.PageOptions),
		sql.GetOffset(opts.PageOptions),
	)
	items, err := sql.CollectRows(rows, db.scan)
	if err != nil {
		return nil, err
	}
	count, err := db.Int(ctx, `
SELECT count(*)
FROM configuration_versions
WHERE configuration_versions.workspace_id = $1
`, workspaceID)
	if err != nil {
		return nil, err
	}
	return resource.NewPage(items, opts.PageOptions, new(count)), nil
}

func (db *pgdb) listOlderThan(ctx context.Context, t time.Time) ([]*ConfigurationVersion, error) {
	rows := db.Query(ctx, `
SELECT
    cv.configuration_version_id,
    cv.created_at,
    cv.auto_queue_runs,
    cv.source,
    cv.speculative,
    cv.status,
    cv.workspace_id,
    (
        SELECT array_agg(cst.*)::configuration_version_status_timestamps[]
        FROM configuration_version_status_timestamps cst
        WHERE cst.configuration_version_id = cv.configuration_version_id
        GROUP BY cst.configuration_version_id
    ) AS status_timestamps,
    ia::"ingress_attributes" AS ingress_attributes
FROM configuration_versions cv
JOIN workspaces USING (workspace_id)
LEFT JOIN ingress_attributes ia USING (configuration_version_id)
WHERE cv.created_at < $1::timestamptz
`,
		t,
	)
	return sql.CollectRows(rows, db.scan)
}

func (db *pgdb) get(ctx context.Context, id resource.ID) (*ConfigurationVersion, error) {
	row := db.Query(ctx, `
SELECT
    cv.configuration_version_id,
    cv.created_at,
    cv.auto_queue_runs,
    cv.source,
    cv.speculative,
    cv.status,
    cv.workspace_id,
    (
        SELECT array_agg(cst.*)::configuration_version_status_timestamps[]
        FROM configuration_version_status_timestamps cst
        WHERE cst.configuration_version_id = cv.configuration_version_id
        GROUP BY cst.configuration_version_id
    ) AS status_timestamps,
    ia::ingress_attributes AS ingress_attributes
FROM configuration_versions cv
JOIN workspaces USING (workspace_id)
LEFT JOIN ingress_attributes ia USING(configuration_version_id)
WHERE cv.configuration_version_id = $1
`, id)
	return sql.CollectOneRow(row, db.scan)
}

func (db *pgdb) getLatest(ctx context.Context, workspaceID resource.ID) (*ConfigurationVersion, error) {
	row := db.Query(ctx, `
SELECT
    cv.configuration_version_id,
    cv.created_at,
    cv.auto_queue_runs,
    cv.source,
    cv.speculative,
    cv.status,
    cv.workspace_id,
    (
        SELECT array_agg(cst.*)::configuration_version_status_timestamps[]
        FROM configuration_version_status_timestamps cst
        WHERE cst.configuration_version_id = cv.configuration_version_id
        GROUP BY cst.configuration_version_id
    ) AS status_timestamps,
    ia::ingress_attributes AS ingress_attributes
FROM configuration_versions cv
JOIN workspaces USING (workspace_id)
LEFT JOIN ingress_attributes ia USING(configuration_version_id)
WHERE cv.workspace_id = $1
ORDER BY cv.created_at DESC
`, workspaceID)
	return sql.CollectOneRow(row, db.scan)
}

func (db *pgdb) GetConfig(ctx context.Context, id resource.TfeID) ([]byte, error) {
	row := db.Query(ctx, `
SELECT config
FROM configuration_versions
WHERE configuration_version_id = $1
AND   status                   = 'uploaded'
`, id)
	return sql.CollectOneType[[]byte](row)
}

func (db *pgdb) DeleteConfigurationVersion(ctx context.Context, id resource.TfeID) error {
	_, err := db.Exec(ctx, `
DELETE
FROM configuration_versions
WHERE configuration_version_id = $1
RETURNING configuration_version_id
`, id)
	return err
}

func (db *pgdb) insertCVStatusTimestamp(ctx context.Context, cv *ConfigurationVersion) error {
	sts, err := cv.StatusTimestamp(cv.Status)
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, `
INSERT INTO configuration_version_status_timestamps (
    configuration_version_id,
    status,
    timestamp
) VALUES (
    $1,
    $2,
    $3
)`,
		cv.ID,
		cv.Status,
		sts,
	)
	return err
}

type IngressAttributesModel struct {
	Branch                 string
	CommitSHA              string         `db:"commit_sha"`
	Repo                   vcs.Repo       `db:"identifier"`
	IsPullRequest          bool           `db:"is_pull_request"`
	OnDefaultBranch        bool           `db:"on_default_branch"`
	ConfigurationVersionID resource.TfeID `db:"configuration_version_id"`
	CommitURL              string         `db:"commit_url"`
	PullRequestNumber      int            `db:"pull_request_number"`
	PullRequestURL         string         `db:"pull_request_url"`
	PullRequestTitle       string         `db:"pull_request_title"`
	Tag                    string         `db:"tag"`
	SenderUsername         string         `db:"sender_username"`
	SenderAvatarURL        string         `db:"sender_avatar_url"`
	SenderHTMLURL          string         `db:"sender_html_url"`
}

func (m IngressAttributesModel) ToIngressAttributes() *IngressAttributes {
	return &IngressAttributes{
		Branch:                 m.Branch,
		CommitSHA:              m.CommitSHA,
		Repo:                   m.Repo,
		IsPullRequest:          m.IsPullRequest,
		OnDefaultBranch:        m.OnDefaultBranch,
		ConfigurationVersionID: m.ConfigurationVersionID,
		CommitURL:              m.CommitURL,
		PullRequestNumber:      m.PullRequestNumber,
		PullRequestURL:         m.PullRequestURL,
		PullRequestTitle:       m.PullRequestTitle,
		Tag:                    m.Tag,
		SenderUsername:         m.SenderUsername,
		SenderAvatarURL:        m.SenderAvatarURL,
		SenderHTMLURL:          m.SenderHTMLURL,
	}
}

func (db *pgdb) scan(row pgx.CollectableRow) (*ConfigurationVersion, error) {
	type model struct {
		ID                resource.TfeID `db:"configuration_version_id"`
		CreatedAt         time.Time      `db:"created_at"`
		AutoQueueRuns     bool           `db:"auto_queue_runs"`
		Source            source.Source
		Speculative       bool
		Status            ConfigurationStatus
		StatusTimestamps  []StatusTimestamp       `db:"status_timestamps"`
		WorkspaceID       resource.TfeID          `db:"workspace_id"`
		IngressAttributes *IngressAttributesModel `db:"ingress_attributes"`
	}
	m, err := pgx.RowToAddrOfStructByName[model](row)
	if err != nil {
		return nil, err
	}
	cv := &ConfigurationVersion{
		ID:               m.ID,
		CreatedAt:        m.CreatedAt,
		AutoQueueRuns:    m.AutoQueueRuns,
		Source:           m.Source,
		Speculative:      m.Speculative,
		Status:           m.Status,
		StatusTimestamps: m.StatusTimestamps,
		WorkspaceID:      m.WorkspaceID,
	}
	if m.IngressAttributes != nil {
		cv.IngressAttributes = m.IngressAttributes.ToIngressAttributes()
	}
	return cv, nil
}
