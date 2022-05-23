package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

var (
	_ otf.StateVersionStore = (*StateVersionDB)(nil)
)

type StateVersionDB struct {
	*pgxpool.Pool
}

func NewStateVersionDB(conn *pgxpool.Pool) *StateVersionDB {
	return &StateVersionDB{
		Pool: conn,
	}
}

// Create persists a StateVersion to the DB.
func (s StateVersionDB) Create(workspaceID string, sv *otf.StateVersion) error {
	ctx := context.Background()

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := pggen.NewQuerier(tx)

	result, err := q.InsertStateVersion(ctx, pggen.InsertStateVersionParams{
		ID:          sv.ID(),
		Serial:      int(sv.Serial),
		State:       sv.State,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return err
	}
	sv.CreatedAt = result.CreatedAt
	sv.UpdatedAt = result.UpdatedAt

	// Insert state_version_outputs
	for _, svo := range sv.Outputs {
		result, err := q.InsertStateVersionOutput(ctx, pggen.InsertStateVersionOutputParams{
			ID:             svo.ID(),
			Name:           svo.Name,
			Sensitive:      svo.Sensitive,
			Type:           svo.Type,
			Value:          svo.Value,
			StateVersionID: sv.ID(),
		})
		if err != nil {
			return err
		}
		svo.CreatedAt = result.CreatedAt
		svo.UpdatedAt = result.UpdatedAt
	}

	return tx.Commit(ctx)
}

func (s StateVersionDB) List(opts otf.StateVersionListOptions) (*otf.StateVersionList, error) {
	if opts.Workspace == nil {
		return nil, fmt.Errorf("missing required option: workspace")
	}
	if opts.Organization == nil {
		return nil, fmt.Errorf("missing required option: organization")
	}

	q := pggen.NewQuerier(s.Pool)
	batch := &pgx.Batch{}
	ctx := context.Background()

	q.FindStateVersionsByWorkspaceNameBatch(batch, pggen.FindStateVersionsByWorkspaceNameParams{
		WorkspaceName:    *opts.Workspace,
		OrganizationName: *opts.Organization,
		Limit:            opts.GetLimit(),
		Offset:           opts.GetOffset(),
	})
	q.CountStateVersionsByWorkspaceNameBatch(batch, *opts.Workspace, *opts.Organization)

	results := s.Pool.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := q.FindStateVersionsByWorkspaceNameScan(results)
	if err != nil {
		return nil, err
	}
	count, err := q.CountStateVersionsByWorkspaceNameScan(results)
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

func (s StateVersionDB) Get(opts otf.StateVersionGetOptions) (*otf.StateVersion, error) {
	ctx := context.Background()
	q := pggen.NewQuerier(s.Pool)

	if opts.ID != nil {
		result, err := q.FindStateVersionByID(ctx, *opts.ID)
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalStateVersionDBResult(otf.StateVersionDBRow(result))
	} else if opts.WorkspaceID != nil {
		result, err := q.FindStateVersionLatestByWorkspaceID(ctx, *opts.WorkspaceID)
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalStateVersionDBResult(otf.StateVersionDBRow(result))
	} else {
		return nil, fmt.Errorf("no state version spec provided")
	}
}

func (s StateVersionDB) GetState(id string) ([]byte, error) {
	ctx := context.Background()
	q := pggen.NewQuerier(s.Pool)

	return q.FindStateVersionStateByID(ctx, id)
}

// Delete deletes a state version from the DB
func (s StateVersionDB) Delete(id string) error {
	ctx := context.Background()
	q := pggen.NewQuerier(s.Pool)

	result, err := q.DeleteStateVersionByID(ctx, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return nil
}
