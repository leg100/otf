package workspace

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

func (db *pgdb) listTags(ctx context.Context, organization organization.Name, opts ListTagsOptions) (*resource.Page[*Tag], error) {
	rows := db.Query(ctx, `
SELECT
    t.tag_id, t.name, t.organization_name,
    (
        SELECT count(*)
        FROM workspace_tags wt
        WHERE wt.tag_id = t.tag_id
    ) AS instance_count
FROM tags t
WHERE t.organization_name = $1
`,
		organization,
	)
	items, err := sql.CollectRows(rows, db.scanWorkspaceTag)
	if err != nil {
		return nil, err
	}
	return resource.NewPage(items, opts.PageOptions, nil), nil
}

func (db *pgdb) deleteTags(ctx context.Context, organization organization.Name, tagIDs []resource.TfeID) error {
	err := db.Tx(ctx, func(ctx context.Context) error {
		for _, tid := range tagIDs {
			_, err := db.Exec(ctx, `
DELETE
FROM tags
WHERE tag_id            = $1
AND   organization_name = $2
`, tid, organization)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (db *pgdb) addTag(ctx context.Context, organization organization.Name, name string, tagID resource.TfeID) error {
	_, err := db.Exec(ctx, `
INSERT INTO tags (
    tag_id,
    name,
    organization_name
) VALUES (
    $1,
    $2,
    $3
) ON CONFLICT (organization_name, name) DO NOTHING
`, tagID, name, organization)
	if err != nil {
		return err
	}
	return nil
}

func (db *pgdb) findTagByName(ctx context.Context, organization organization.Name, name string) (*Tag, error) {
	row := db.Query(ctx, `
SELECT
    t.tag_id, t.name, t.organization_name,
    (
        SELECT count(*)
        FROM workspace_tags wt
        WHERE wt.tag_id = t.tag_id
    ) AS instance_count
FROM tags t
WHERE t.name = $1
AND   t.organization_name = $2
`, name, organization)
	tag, err := sql.CollectOneRow(row, db.scanWorkspaceTag)
	if err != nil {
		return tag, err
	}
	return tag, nil
}

func (db *pgdb) findTagByID(ctx context.Context, organization organization.Name, id resource.TfeID) (*Tag, error) {
	row := db.Query(ctx, `
SELECT
    t.tag_id, t.name, t.organization_name,
    (
        SELECT count(*)
        FROM workspace_tags wt
        WHERE wt.tag_id = t.tag_id
    ) AS instance_count
FROM tags t
WHERE t.tag_id = $1
AND   t.organization_name = $2
`,
		id,
		organization,
	)
	tag, err := sql.CollectOneRow(row, db.scanWorkspaceTag)
	if err != nil {
		return tag, err
	}
	return tag, nil
}

func (db *pgdb) tagWorkspace(ctx context.Context, workspaceID, tagID resource.TfeID) error {
	result, err := db.Exec(ctx, `
INSERT INTO workspace_tags (
    workspace_id,
    tag_id
) SELECT $1, $2
  FROM workspaces w
  JOIN tags t ON (t.organization_name = w.organization_name)
  WHERE w.workspace_id = $1
  AND t.tag_id = $2
  RETURNING tag_id
`,
		workspaceID,
		tagID)
	if result.RowsAffected() == 0 {
		return internal.ErrResourceNotFound
	}
	return err
}

func (db *pgdb) deleteWorkspaceTag(ctx context.Context, workspaceID, tagID resource.TfeID) error {
	_, err := db.Exec(ctx, `
DELETE
FROM workspace_tags
WHERE workspace_id  = $1
AND   tag_id        = $2
`,
		workspaceID,
		tagID)
	return err
}

func (db *pgdb) listWorkspaceTags(ctx context.Context, workspaceID resource.TfeID, opts ListWorkspaceTagsOptions) (*resource.Page[*Tag], error) {
	rows := db.Query(ctx, `
SELECT
    t.tag_id, t.name, t.organization_name,
    (
        SELECT count(*)
        FROM workspace_tags wt
        WHERE wt.tag_id = t.tag_id
    ) AS instance_count
FROM workspace_tags wt
JOIN tags t USING (tag_id)
WHERE wt.workspace_id = $1
`,

		workspaceID,
	)
	items, err := sql.CollectRows(rows, db.scanWorkspaceTag)
	if err != nil {
		return nil, err
	}
	return resource.NewPage(items, opts.PageOptions, nil), nil
}

func (db *pgdb) scanWorkspaceTag(row pgx.CollectableRow) (*Tag, error) {
	var tag Tag
	err := row.Scan(&tag.ID, &tag.Name, &tag.Organization, &tag.InstanceCount)
	return &tag, err
}
