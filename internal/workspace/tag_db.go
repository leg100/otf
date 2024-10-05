package workspace

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
)

type (
	// pgresult represents the result of a database query for a tag.
	tagresult struct {
		TagID            pgtype.Text
		Name             pgtype.Text
		OrganizationName pgtype.Text
		InstanceCount    int64
	}
)

// toTag converts a database result into a tag
func (r tagresult) toTag() *Tag {
	return &Tag{
		ID:            r.TagID.String,
		Name:          r.Name.String,
		Organization:  r.OrganizationName.String,
		InstanceCount: int(r.InstanceCount),
	}
}

func (db *pgdb) listTags(ctx context.Context, organization string, opts ListTagsOptions) (*resource.Page[*Tag], error) {
	q := db.Querier(ctx)

	rows, err := q.FindTags(ctx, sqlc.FindTagsParams{
		OrganizationName: sql.String(organization),
		Limit:            opts.GetLimit(),
		Offset:           opts.GetOffset(),
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	count, err := q.CountTags(ctx, sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}

	items := make([]*Tag, len(rows))
	for i, r := range rows {
		items[i] = tagresult(r).toTag()
	}
	return resource.NewPage(items, opts.PageOptions, internal.Int64(count)), nil
}

func (db *pgdb) deleteTags(ctx context.Context, organization string, tagIDs []string) error {
	err := db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		for _, tid := range tagIDs {
			_, err := q.DeleteTag(ctx, sqlc.DeleteTagParams{
				TagID:            sql.String(tid),
				OrganizationName: sql.String(organization),
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return sql.Error(err)
}

func (db *pgdb) addTag(ctx context.Context, organization, name, id string) error {
	err := db.Querier(ctx).InsertTag(ctx, sqlc.InsertTagParams{
		TagID:            sql.String(id),
		Name:             sql.String(name),
		OrganizationName: sql.String(organization),
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) findTagByName(ctx context.Context, organization, name string) (*Tag, error) {
	tag, err := db.Querier(ctx).FindTagByName(ctx, sqlc.FindTagByNameParams{
		Name:             sql.String(name),
		OrganizationName: sql.String(organization),
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	return tagresult(tag).toTag(), nil
}

func (db *pgdb) findTagByID(ctx context.Context, organization, id string) (*Tag, error) {
	tag, err := db.Querier(ctx).FindTagByID(ctx, sqlc.FindTagByIDParams{
		TagID:            sql.String(id),
		OrganizationName: sql.String(organization),
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	return tagresult(tag).toTag(), nil
}

func (db *pgdb) tagWorkspace(ctx context.Context, workspaceID, tagID string) error {
	_, err := db.Querier(ctx).InsertWorkspaceTag(ctx, sqlc.InsertWorkspaceTagParams{
		TagID:       sql.String(tagID),
		WorkspaceID: sql.String(workspaceID),
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) deleteWorkspaceTag(ctx context.Context, workspaceID, tagID string) error {
	_, err := db.Querier(ctx).DeleteWorkspaceTag(ctx, sqlc.DeleteWorkspaceTagParams{
		WorkspaceID: sql.String(workspaceID),
		TagID:       sql.String(tagID),
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) listWorkspaceTags(ctx context.Context, workspaceID string, opts ListWorkspaceTagsOptions) (*resource.Page[*Tag], error) {
	q := db.Querier(ctx)

	rows, err := q.FindWorkspaceTags(ctx, sqlc.FindWorkspaceTagsParams{
		WorkspaceID: sql.String(workspaceID),
		Limit:       opts.GetLimit(),
		Offset:      opts.GetOffset(),
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	count, err := q.CountTags(ctx, sql.String(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}

	items := make([]*Tag, len(rows))
	for i, r := range rows {
		items[i] = tagresult(r).toTag()
	}
	return resource.NewPage(items, opts.PageOptions, internal.Int64(count)), nil
}
