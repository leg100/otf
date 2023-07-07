package workspace

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

type (
	// pgresult represents the result of a database query for a tag.
	tagresult struct {
		TagID            pgtype.Text `json:"tag_id"`
		Name             pgtype.Text `json:"name"`
		OrganizationName pgtype.Text `json:"organization_name"`
		InstanceCount    pgtype.Int8 `json:"instance_count"`
	}
)

// toTag converts a database result into a tag
func (r tagresult) toTag() *Tag {
	return &Tag{
		ID:            r.TagID.String,
		Name:          r.Name.String,
		Organization:  r.OrganizationName.String,
		InstanceCount: int(r.InstanceCount.Int),
	}
}

func (db *pgdb) listTags(ctx context.Context, organization string, opts ListTagsOptions) (*resource.Page[*Tag], error) {
	q := db.Conn(ctx)
	batch := &pgx.Batch{}

	q.FindTagsBatch(batch, pggen.FindTagsParams{
		OrganizationName: sql.String(organization),
		Limit:            opts.GetLimit(),
		Offset:           opts.GetOffset(),
	})
	q.CountTagsBatch(batch, sql.String(organization))
	results := db.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := q.FindTagsScan(results)
	if err != nil {
		return nil, sql.Error(err)
	}
	count, err := q.CountTagsScan(results)
	if err != nil {
		return nil, sql.Error(err)
	}

	var items []*Tag
	for _, r := range rows {
		items = append(items, tagresult(r).toTag())
	}

	return resource.NewPage(items, opts.PageOptions, internal.Int64(count.Int)), nil
}

func (db *pgdb) deleteTags(ctx context.Context, organization string, tagIDs []string) error {
	err := db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		for _, tid := range tagIDs {
			_, err := q.DeleteTag(ctx, sql.String(tid), sql.String(organization))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return sql.Error(err)
}

func (db *pgdb) addTag(ctx context.Context, organization, name, id string) error {
	_, err := db.Conn(ctx).InsertTag(ctx, pggen.InsertTagParams{
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
	tag, err := db.Conn(ctx).FindTagByName(ctx, sql.String(name), sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}
	return tagresult(tag).toTag(), nil
}

func (db *pgdb) findTagByID(ctx context.Context, organization, id string) (*Tag, error) {
	tag, err := db.Conn(ctx).FindTagByID(ctx, sql.String(id), sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}
	return tagresult(tag).toTag(), nil
}

func (db *pgdb) tagWorkspace(ctx context.Context, workspaceID, tagID string) error {
	_, err := db.Conn(ctx).InsertWorkspaceTag(ctx, sql.String(tagID), sql.String(workspaceID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) deleteWorkspaceTag(ctx context.Context, workspaceID, tagID string) error {
	_, err := db.Conn(ctx).DeleteWorkspaceTag(ctx, sql.String(workspaceID), sql.String(tagID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) listWorkspaceTags(ctx context.Context, workspaceID string, opts ListWorkspaceTagsOptions) (*resource.Page[*Tag], error) {
	q := db.Conn(ctx)
	batch := &pgx.Batch{}

	q.FindWorkspaceTagsBatch(batch, pggen.FindWorkspaceTagsParams{
		WorkspaceID: sql.String(workspaceID),
		Limit:       opts.GetLimit(),
		Offset:      opts.GetOffset(),
	})
	q.CountTagsBatch(batch, sql.String(workspaceID))
	results := db.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := q.FindWorkspaceTagsScan(results)
	if err != nil {
		return nil, sql.Error(err)
	}
	count, err := q.CountWorkspaceTagsScan(results)
	if err != nil {
		return nil, sql.Error(err)
	}

	var items []*Tag
	for _, r := range rows {
		items = append(items, tagresult(r).toTag())
	}

	return resource.NewPage(items, opts.PageOptions, internal.Int64(count.Int)), nil
}
