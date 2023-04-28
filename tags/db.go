package tags

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

type (
	// pgdb is a tag database using postgres
	pgdb struct {
		otf.DB // provides access to generated SQL queries
	}

	// pgresult represents the result of a database query for a tag.
	pgresult struct {
		TagID            pgtype.Text `json:"tag_id"`
		Name             pgtype.Text `json:"name"`
		OrganizationName pgtype.Text `json:"organization_name"`
		InstanceCount    int         `json:"instance_count"`
	}
)

// toTag converts a database result into a tag
func (r pgresult) toTag() *Tag {
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
		items = append(items, pgresult(r).toTag())
	}

	return &TagList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, count),
	}, nil
}

func (db *pgdb) deleteTags(ctx context.Context, organization string, tagIDs []string) error {
	err := db.Tx(ctx, func(tx otf.DB) error {
		for _, tid := range tagIDs {
			_, err := tx.DeleteTag(ctx, sql.String(organization), sql.String(tid))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return sql.Error(err)
}

func (db *pgdb) tagWorkspaces(ctx context.Context, tagID string, workspaceIDs []string) error {
	err := db.Tx(ctx, func(tx otf.DB) error {
		for _, wid := range workspaceIDs {
			_, err := tx.InsertWorkspaceTag(ctx, sql.String(tagID), sql.String(wid))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return sql.Error(err)
}

func (db *pgdb) addTag(ctx context.Context, workspaceID string, tag *Tag) error {
	_, err := db.InsertTag(ctx, pggen.InsertTagParams{
		TagID:       sql.String(tag.ID),
		Name:        sql.String(tag.Name),
		WorkspaceID: sql.String(workspaceID),
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

func (db *pgdb) findTagByName(ctx context.Context, name string) (*Tag, error) {
	tag, err := db.FindTagByName(ctx, sql.String(name))
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgresult(tag).toTag(), nil
}

func (db *pgdb) findTagByID(ctx context.Context, id string) (*Tag, error) {
	tag, err := db.FindTagByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgresult(tag).toTag(), nil
}

func (db *pgdb) addWorkspaceTag(ctx context.Context, workspaceID, tagID string) error {
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
		items = append(items, pgresult(r).toTag())
	}

	return &TagList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, count),
	}, nil
}

// tx constructs a new pgdb within a transaction.
func (db *pgdb) tx(ctx context.Context, callback func(*pgdb) error) error {
	return db.Tx(ctx, func(tx otf.DB) error {
		return callback(&pgdb{tx})
	})
}

// lock webhooks table within a transaction, providing a callback within which
// caller can use the transaction.
func (db *pgdb) lock(ctx context.Context, callback func(*pgdb) error) error {
	return db.Tx(ctx, func(tx otf.DB) error {
		if _, err := tx.Exec(ctx, "LOCK tags"); err != nil {
			return err
		}
		return callback(&pgdb{tx})
	})
}
