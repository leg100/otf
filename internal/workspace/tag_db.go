package workspace

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

type (
	// pgresult represents the result of a database query for a tag.
	tagresult struct {
		TagID            pgtype.Text `json:"tag_id"`
		Name             pgtype.Text `json:"name"`
		OrganizationName pgtype.Text `json:"organization_name"`
		InstanceCount    int         `json:"instance_count"`
	}
)

// toTag converts a database result into a tag
func (r tagresult) toTag() *Tag {
	return &Tag{
		ID:            r.TagID.String,
		Name:          r.Name.String,
		Organization:  r.OrganizationName.String,
		InstanceCount: r.InstanceCount,
	}
}

func (db *pgdb) listTags(ctx context.Context, organization string, opts ListTagsOptions) (*TagList, error) {
	batch := &pgx.Batch{}

	db.FindTagsBatch(batch, pggen.FindTagsParams{
		OrganizationName: sql.String(organization),
		Limit:            opts.GetLimit(),
		Offset:           opts.GetOffset(),
	})
	db.CountTagsBatch(batch, sql.String(organization))
	results := db.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := db.FindTagsScan(results)
	if err != nil {
		return nil, sql.Error(err)
	}
	count, err := db.CountTagsScan(results)
	if err != nil {
		return nil, sql.Error(err)
	}

	var items []*Tag
	for _, r := range rows {
		items = append(items, tagresult(r).toTag())
	}

	return &TagList{
		Items:      items,
		Pagination: internal.NewPagination(opts.ListOptions, count),
	}, nil
}

func (db *pgdb) deleteTags(ctx context.Context, organization string, tagIDs []string) error {
	err := db.Tx(ctx, func(tx internal.DB) error {
		for _, tid := range tagIDs {
			_, err := tx.DeleteTag(ctx, sql.String(tid), sql.String(organization))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return sql.Error(err)
}

func (db *pgdb) addTag(ctx context.Context, organization, name, id string) error {
	_, err := db.InsertTag(ctx, pggen.InsertTagParams{
		TagID:            sql.String(id),
		Name:             sql.String(name),
		OrganizationName: sql.String(organization),
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) deleteTag(ctx context.Context, tag *Tag) error {
	_, err := db.DeleteTag(ctx, sql.String(tag.ID), sql.String(tag.Organization))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) findTagByName(ctx context.Context, organization, name string) (*Tag, error) {
	tag, err := db.FindTagByName(ctx, sql.String(name), sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}
	return tagresult(tag).toTag(), nil
}

func (db *pgdb) findTagByID(ctx context.Context, organization, id string) (*Tag, error) {
	tag, err := db.FindTagByID(ctx, sql.String(id), sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}
	return tagresult(tag).toTag(), nil
}

func (db *pgdb) tagWorkspace(ctx context.Context, workspaceID, tagID string) error {
	_, err := db.InsertWorkspaceTag(ctx, sql.String(tagID), sql.String(workspaceID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) deleteWorkspaceTag(ctx context.Context, workspaceID, tagID string) error {
	_, err := db.DeleteWorkspaceTag(ctx, sql.String(workspaceID), sql.String(tagID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) listWorkspaceTags(ctx context.Context, workspaceID string, opts ListWorkspaceTagsOptions) (*TagList, error) {
	batch := &pgx.Batch{}

	db.FindWorkspaceTagsBatch(batch, pggen.FindWorkspaceTagsParams{
		WorkspaceID: sql.String(workspaceID),
		Limit:       opts.GetLimit(),
		Offset:      opts.GetOffset(),
	})
	db.CountTagsBatch(batch, sql.String(workspaceID))
	results := db.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := db.FindWorkspaceTagsScan(results)
	if err != nil {
		return nil, sql.Error(err)
	}
	count, err := db.CountWorkspaceTagsScan(results)
	if err != nil {
		return nil, sql.Error(err)
	}

	var items []*Tag
	for _, r := range rows {
		items = append(items, tagresult(r).toTag())
	}

	return &TagList{
		Items:      items,
		Pagination: internal.NewPagination(opts.ListOptions, count),
	}, nil
}

// lockTags tags table within a transaction, providing a callback within which
// caller can use the transaction.
func (db *pgdb) lockTags(ctx context.Context, callback func(*pgdb) error) error {
	return db.Tx(ctx, func(tx internal.DB) error {
		if _, err := tx.Exec(ctx, "LOCK TABLE tags IN EXCLUSIVE MODE"); err != nil {
			return err
		}
		return callback(&pgdb{tx})
	})
}
