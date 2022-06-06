package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

// CreateStateVersion persists a StateVersion to the DB.
func (db *DB) CreateStateVersion(ctx context.Context, workspaceID string, sv *otf.StateVersion) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := pggen.NewQuerier(tx)

	_, err = q.InsertStateVersion(ctx, pggen.InsertStateVersionParams{
		ID:          pgtype.Text{String: sv.ID(), Status: pgtype.Present},
		CreatedAt:   sv.CreatedAt(),
		Serial:      int(sv.Serial()),
		State:       sv.State(),
		WorkspaceID: pgtype.Text{String: workspaceID, Status: pgtype.Present},
	})
	if err != nil {
		return err
	}

	// Insert state_version_outputs
	for _, svo := range sv.Outputs() {
		_, err := q.InsertStateVersionOutput(ctx, pggen.InsertStateVersionOutputParams{
			ID:             pgtype.Text{String: svo.ID(), Status: pgtype.Present},
			Name:           pgtype.Text{String: svo.Name, Status: pgtype.Present},
			Sensitive:      svo.Sensitive,
			Type:           pgtype.Text{String: svo.Type, Status: pgtype.Present},
			Value:          pgtype.Text{String: svo.Value, Status: pgtype.Present},
			StateVersionID: pgtype.Text{String: sv.ID(), Status: pgtype.Present},
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (db *DB) ListStateVersions(ctx context.Context, opts otf.StateVersionListOptions) (*otf.StateVersionList, error) {
	if opts.Workspace == nil {
		return nil, fmt.Errorf("missing required option: workspace")
	}
	if opts.Organization == nil {
		return nil, fmt.Errorf("missing required option: organization")
	}

	batch := &pgx.Batch{}

	db.FindStateVersionsByWorkspaceNameBatch(batch, pggen.FindStateVersionsByWorkspaceNameParams{
		WorkspaceName:    pgtype.Text{String: *opts.Workspace, Status: pgtype.Present},
		OrganizationName: pgtype.Text{String: *opts.Organization, Status: pgtype.Present},
		Limit:            opts.GetLimit(),
		Offset:           opts.GetOffset(),
	})
	db.CountStateVersionsByWorkspaceNameBatch(batch,
		pgtype.Text{String: *opts.Workspace, Status: pgtype.Present},
		pgtype.Text{String: *opts.Organization, Status: pgtype.Present},
	)

	results := db.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := db.FindStateVersionsByWorkspaceNameScan(results)
	if err != nil {
		return nil, err
	}
	count, err := db.CountStateVersionsByWorkspaceNameScan(results)
	if err != nil {
		return nil, err
	}

	var items []*otf.StateVersion
	for _, r := range rows {
		sv, err := otf.UnmarshalStateVersionDBResult(otf.StateVersionDBRow(r))
		if err != nil {
			return nil, err
		}
		items = append(items, sv)
	}

	return &otf.StateVersionList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, *count),
	}, nil
}

func (db *DB) GetStateVersion(ctx context.Context, opts otf.StateVersionGetOptions) (*otf.StateVersion, error) {
	if opts.ID != nil {
		result, err := db.FindStateVersionByID(ctx, pgtype.Text{String: *opts.ID, Status: pgtype.Present})
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalStateVersionDBResult(otf.StateVersionDBRow(result))
	} else if opts.WorkspaceID != nil {
		result, err := db.FindStateVersionLatestByWorkspaceID(ctx,
			pgtype.Text{String: *opts.WorkspaceID, Status: pgtype.Present},
		)
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalStateVersionDBResult(otf.StateVersionDBRow(result))
	} else {
		return nil, fmt.Errorf("no state version spec provided")
	}
}

func (db *DB) GetState(ctx context.Context, id string) ([]byte, error) {
	return db.FindStateVersionStateByID(ctx, pgtype.Text{String: id, Status: pgtype.Present})
}

// DeleteStateVersion deletes a state version from the DB
func (db *DB) DeleteStateVersion(ctx context.Context, id string) error {
	result, err := db.DeleteStateVersionByID(ctx, pgtype.Text{String: id, Status: pgtype.Present})
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return nil
}
