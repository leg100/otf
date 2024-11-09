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
		TagID            resource.ID
		Name             pgtype.Text
		OrganizationName pgtype.Text
		InstanceCount    int64
	}
)

// toTag converts a database result into a tag
func (r tagresult) toTag() *Tag {
	return &Tag{
		ID:            r.TagID,
		Name:          r.Name.String,
		Organization:  r.OrganizationName.String,
		InstanceCount: int(r.InstanceCount),
	}
}

func (db *pgdb) listTags(ctx context.Context, organization string, opts ListTagsOptions) (*resource.Page[*Tag], error) {
	q := db.Querier(ctx)

	rows, err := q.FindTags(ctx, sqlc.FindTagsParams{
		OrganizationName: sql.String(organization),
		Limit:            sql.GetLimit(opts.PageOptions),
		Offset:           sql.GetOffset(opts.PageOptions),
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

func (db *pgdb) deleteTags(ctx context.Context, organization string, tagIDs []resource.ID) error {
	err := db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		for _, tid := range tagIDs {
			_, err := q.DeleteTag(ctx, sqlc.DeleteTagParams{
				TagID:            tid,
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

func (db *pgdb) addTag(ctx context.Context, organization, name string, tagID resource.ID) error {
	err := db.Querier(ctx).InsertTag(ctx, sqlc.InsertTagParams{
		TagID:            tagID,
		Name:             sql.String(name),
		OrganizationName: sql.String(organization),
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) findTagByName(ctx context.Context, organization string, name string) (*Tag, error) {
	tag, err := db.Querier(ctx).FindTagByName(ctx, sqlc.FindTagByNameParams{
		Name:             sql.String(name),
		OrganizationName: sql.String(organization),
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	return tagresult(tag).toTag(), nil
}

func (db *pgdb) findTagByID(ctx context.Context, organization string, id resource.ID) (*Tag, error) {
	tag, err := db.Querier(ctx).FindTagByID(ctx, sqlc.FindTagByIDParams{
		TagID:            id,
		OrganizationName: sql.String(organization),
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	return tagresult(tag).toTag(), nil
}

func (db *pgdb) tagWorkspace(ctx context.Context, workspaceID, tagID resource.ID) error {
	_, err := db.Querier(ctx).InsertWorkspaceTag(ctx, sqlc.InsertWorkspaceTagParams{
		TagID:       tagID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) deleteWorkspaceTag(ctx context.Context, workspaceID, tagID resource.ID) error {
	_, err := db.Querier(ctx).DeleteWorkspaceTag(ctx, sqlc.DeleteWorkspaceTagParams{
		WorkspaceID: workspaceID,
		TagID:       tagID,
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) listWorkspaceTags(ctx context.Context, workspaceID resource.ID, opts ListWorkspaceTagsOptions) (*resource.Page[*Tag], error) {
	q := db.Querier(ctx)

	rows, err := q.FindWorkspaceTags(ctx, sqlc.FindWorkspaceTagsParams{
		WorkspaceID: workspaceID,
		Limit:       sql.GetLimit(opts.PageOptions),
		Offset:      sql.GetOffset(opts.PageOptions),
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	count, err := q.CountWorkspaceTags(ctx, workspaceID)
	if err != nil {
		return nil, sql.Error(err)
	}

	items := make([]*Tag, len(rows))
	for i, r := range rows {
		items[i] = tagresult(r).toTag()
	}
	return resource.NewPage(items, opts.PageOptions, internal.Int64(count)), nil
}
