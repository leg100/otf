// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: tags.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal/resource"
)

const countTags = `-- name: CountTags :one
SELECT count(*)
FROM tags t
WHERE t.organization_name = $1
`

func (q *Queries) CountTags(ctx context.Context, organizationName pgtype.Text) (int64, error) {
	row := q.db.QueryRow(ctx, countTags, organizationName)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const countWorkspaceTags = `-- name: CountWorkspaceTags :one
SELECT count(*)
FROM workspace_tags wt
WHERE wt.workspace_id = $1
`

func (q *Queries) CountWorkspaceTags(ctx context.Context, workspaceID resource.ID) (int64, error) {
	row := q.db.QueryRow(ctx, countWorkspaceTags, workspaceID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const deleteTag = `-- name: DeleteTag :one
DELETE
FROM tags
WHERE tag_id            = $1
AND   organization_name = $2
RETURNING tag_id
`

type DeleteTagParams struct {
	TagID            pgtype.Text
	OrganizationName pgtype.Text
}

func (q *Queries) DeleteTag(ctx context.Context, arg DeleteTagParams) (pgtype.Text, error) {
	row := q.db.QueryRow(ctx, deleteTag, arg.TagID, arg.OrganizationName)
	var tag_id pgtype.Text
	err := row.Scan(&tag_id)
	return tag_id, err
}

const deleteWorkspaceTag = `-- name: DeleteWorkspaceTag :one
DELETE
FROM workspace_tags
WHERE workspace_id  = $1
AND   tag_id        = $2
RETURNING tag_id
`

type DeleteWorkspaceTagParams struct {
	WorkspaceID resource.ID
	TagID       resource.ID
}

func (q *Queries) DeleteWorkspaceTag(ctx context.Context, arg DeleteWorkspaceTagParams) (resource.ID, error) {
	row := q.db.QueryRow(ctx, deleteWorkspaceTag, arg.WorkspaceID, arg.TagID)
	var tag_id resource.ID
	err := row.Scan(&tag_id)
	return tag_id, err
}

const findTagByID = `-- name: FindTagByID :one
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
`

type FindTagByIDParams struct {
	TagID            pgtype.Text
	OrganizationName pgtype.Text
}

type FindTagByIDRow struct {
	TagID            pgtype.Text
	Name             pgtype.Text
	OrganizationName pgtype.Text
	InstanceCount    int64
}

func (q *Queries) FindTagByID(ctx context.Context, arg FindTagByIDParams) (FindTagByIDRow, error) {
	row := q.db.QueryRow(ctx, findTagByID, arg.TagID, arg.OrganizationName)
	var i FindTagByIDRow
	err := row.Scan(
		&i.TagID,
		&i.Name,
		&i.OrganizationName,
		&i.InstanceCount,
	)
	return i, err
}

const findTagByName = `-- name: FindTagByName :one
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
`

type FindTagByNameParams struct {
	Name             pgtype.Text
	OrganizationName pgtype.Text
}

type FindTagByNameRow struct {
	TagID            pgtype.Text
	Name             pgtype.Text
	OrganizationName pgtype.Text
	InstanceCount    int64
}

func (q *Queries) FindTagByName(ctx context.Context, arg FindTagByNameParams) (FindTagByNameRow, error) {
	row := q.db.QueryRow(ctx, findTagByName, arg.Name, arg.OrganizationName)
	var i FindTagByNameRow
	err := row.Scan(
		&i.TagID,
		&i.Name,
		&i.OrganizationName,
		&i.InstanceCount,
	)
	return i, err
}

const findTags = `-- name: FindTags :many
SELECT
    t.tag_id, t.name, t.organization_name,
    (
        SELECT count(*)
        FROM workspace_tags wt
        WHERE wt.tag_id = t.tag_id
    ) AS instance_count
FROM tags t
WHERE t.organization_name = $1
LIMIT $3::int
OFFSET $2::int
`

type FindTagsParams struct {
	OrganizationName pgtype.Text
	Offset           pgtype.Int4
	Limit            pgtype.Int4
}

type FindTagsRow struct {
	TagID            pgtype.Text
	Name             pgtype.Text
	OrganizationName pgtype.Text
	InstanceCount    int64
}

func (q *Queries) FindTags(ctx context.Context, arg FindTagsParams) ([]FindTagsRow, error) {
	rows, err := q.db.Query(ctx, findTags, arg.OrganizationName, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FindTagsRow
	for rows.Next() {
		var i FindTagsRow
		if err := rows.Scan(
			&i.TagID,
			&i.Name,
			&i.OrganizationName,
			&i.InstanceCount,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findWorkspaceTags = `-- name: FindWorkspaceTags :many
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
LIMIT $3::int
OFFSET $2::int
`

type FindWorkspaceTagsParams struct {
	WorkspaceID resource.ID
	Offset      pgtype.Int4
	Limit       pgtype.Int4
}

type FindWorkspaceTagsRow struct {
	TagID            pgtype.Text
	Name             pgtype.Text
	OrganizationName pgtype.Text
	InstanceCount    int64
}

func (q *Queries) FindWorkspaceTags(ctx context.Context, arg FindWorkspaceTagsParams) ([]FindWorkspaceTagsRow, error) {
	rows, err := q.db.Query(ctx, findWorkspaceTags, arg.WorkspaceID, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FindWorkspaceTagsRow
	for rows.Next() {
		var i FindWorkspaceTagsRow
		if err := rows.Scan(
			&i.TagID,
			&i.Name,
			&i.OrganizationName,
			&i.InstanceCount,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertTag = `-- name: InsertTag :exec
INSERT INTO tags (
    tag_id,
    name,
    organization_name
) VALUES (
    $1,
    $2,
    $3
) ON CONFLICT (organization_name, name) DO NOTHING
`

type InsertTagParams struct {
	TagID            pgtype.Text
	Name             pgtype.Text
	OrganizationName pgtype.Text
}

func (q *Queries) InsertTag(ctx context.Context, arg InsertTagParams) error {
	_, err := q.db.Exec(ctx, insertTag, arg.TagID, arg.Name, arg.OrganizationName)
	return err
}

const insertWorkspaceTag = `-- name: InsertWorkspaceTag :one
INSERT INTO workspace_tags (
    tag_id,
    workspace_id
) SELECT $1, $2
  FROM workspaces w
  JOIN tags t ON (t.organization_name = w.organization_name)
  WHERE w.workspace_id = $2
  AND t.tag_id = $1
RETURNING tag_id
`

type InsertWorkspaceTagParams struct {
	TagID       resource.ID
	WorkspaceID resource.ID
}

func (q *Queries) InsertWorkspaceTag(ctx context.Context, arg InsertWorkspaceTagParams) (resource.ID, error) {
	row := q.db.QueryRow(ctx, insertWorkspaceTag, arg.TagID, arg.WorkspaceID)
	var tag_id resource.ID
	err := row.Scan(&tag_id)
	return tag_id, err
}

const insertWorkspaceTagByName = `-- name: InsertWorkspaceTagByName :one
INSERT INTO workspace_tags (
    tag_id,
    workspace_id
) SELECT t.tag_id, $1
  FROM workspaces w
  JOIN tags t ON (t.organization_name = w.organization_name)
  WHERE t.name = $2
RETURNING tag_id
`

type InsertWorkspaceTagByNameParams struct {
	WorkspaceID resource.ID
	TagName     pgtype.Text
}

func (q *Queries) InsertWorkspaceTagByName(ctx context.Context, arg InsertWorkspaceTagByNameParams) (resource.ID, error) {
	row := q.db.QueryRow(ctx, insertWorkspaceTagByName, arg.WorkspaceID, arg.TagName)
	var tag_id resource.ID
	err := row.Scan(&tag_id)
	return tag_id, err
}
