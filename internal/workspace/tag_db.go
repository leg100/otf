package workspace

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

type (
	// pgresult represents the result of a database query for a tag.
	tagresult struct {
		TagID            resource.TfeID
		Name             pgtype.Text
		OrganizationName resource.OrganizationName
		InstanceCount    int64
	}
)

// toTag converts a database result into a tag
func (r tagresult) toTag() *Tag {
	return &Tag{
		ID:            r.TagID,
		Name:          r.Name.String,
		Organization:  r.OrganizationName,
		InstanceCount: int(r.InstanceCount),
	}
}

func (db *pgdb) listTags(ctx context.Context, organization resource.OrganizationName, opts ListTagsOptions) (*resource.Page[*Tag], error) {
	rows, err := q.FindTags(ctx, db.Conn(ctx), FindTagsParams{
		OrganizationName: organization,
		Limit:            sql.GetLimit(opts.PageOptions),
		Offset:           sql.GetOffset(opts.PageOptions),
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	count, err := q.CountTags(ctx, db.Conn(ctx), organization)
	if err != nil {
		return nil, sql.Error(err)
	}

	items := make([]*Tag, len(rows))
	for i, r := range rows {
		items[i] = tagresult(r).toTag()
	}
	return resource.NewPage(items, opts.PageOptions, internal.Int64(count)), nil
}

func (db *pgdb) deleteTags(ctx context.Context, organization resource.OrganizationName, tagIDs []resource.TfeID) error {
	err := db.Tx(ctx, func(ctx context.Context, conn sql.Connection) error {
		for _, tid := range tagIDs {
			_, err := q.DeleteTag(ctx, conn, DeleteTagParams{
				TagID:            tid,
				OrganizationName: organization,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return sql.Error(err)
}

func (db *pgdb) addTag(ctx context.Context, organization resource.OrganizationName, name string, tagID resource.TfeID) error {
	err := q.InsertTag(ctx, db.Conn(ctx), InsertTagParams{
		TagID:            tagID,
		Name:             sql.String(name),
		OrganizationName: organization,
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) findTagByName(ctx context.Context, organization resource.OrganizationName, name string) (*Tag, error) {
	tag, err := q.FindTagByName(ctx, db.Conn(ctx), FindTagByNameParams{
		Name:             sql.String(name),
		OrganizationName: organization,
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	return tagresult(tag).toTag(), nil
}

func (db *pgdb) findTagByID(ctx context.Context, organization resource.OrganizationName, id resource.TfeID) (*Tag, error) {
	tag, err := q.FindTagByID(ctx, db.Conn(ctx), FindTagByIDParams{
		TagID:            id,
		OrganizationName: organization,
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	return tagresult(tag).toTag(), nil
}

func (db *pgdb) tagWorkspace(ctx context.Context, workspaceID, tagID resource.TfeID) error {
	_, err := q.InsertWorkspaceTag(ctx, db.Conn(ctx), InsertWorkspaceTagParams{
		TagID:       tagID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) deleteWorkspaceTag(ctx context.Context, workspaceID, tagID resource.TfeID) error {
	_, err := q.DeleteWorkspaceTag(ctx, db.Conn(ctx), DeleteWorkspaceTagParams{
		WorkspaceID: workspaceID,
		TagID:       tagID,
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) listWorkspaceTags(ctx context.Context, workspaceID resource.TfeID, opts ListWorkspaceTagsOptions) (*resource.Page[*Tag], error) {
	rows, err := q.FindWorkspaceTags(ctx, db.Conn(ctx), FindWorkspaceTagsParams{
		WorkspaceID: workspaceID,
		Limit:       sql.GetLimit(opts.PageOptions),
		Offset:      sql.GetOffset(opts.PageOptions),
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	count, err := q.CountWorkspaceTags(ctx, db.Conn(ctx), workspaceID)
	if err != nil {
		return nil, sql.Error(err)
	}

	items := make([]*Tag, len(rows))
	for i, r := range rows {
		items[i] = tagresult(r).toTag()
	}
	return resource.NewPage(items, opts.PageOptions, internal.Int64(count)), nil
}
