package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
)

var (
	_ otf.StateVersionStore = (*StateVersionService)(nil)
)

type StateVersionService struct {
	*pgx.Conn
}

func NewStateVersionDB(conn *pgx.Conn) *StateVersionService {
	return &StateVersionService{
		Conn: conn,
	}
}

// Create persists a StateVersion to the DB.
func (s StateVersionService) Create(sv *otf.StateVersion) (*otf.StateVersion, error) {
	ctx := context.Background()

	tx, err := s.Conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	q := NewQuerier(tx)

	result, err := q.InsertStateVersion(ctx, InsertStateVersionParams{
		ID:     sv.ID,
		Serial: int32(sv.Serial),
		State:  sv.State,
		RunID:  sv.Run.ID,
	})
	if err != nil {
		return nil, err
	}
	sv.CreatedAt = result.CreatedAt
	sv.UpdatedAt = result.UpdatedAt

	// Insert state_version_outputs
	for _, svo := range sv.Outputs {
		result, err := q.InsertStateVersionOutput(ctx, InsertStateVersionOutputParams{
			ID:             svo.ID,
			Name:           svo.Name,
			Sensitive:      svo.Sensitive,
			Type:           svo.Type,
			Value:          svo.Value,
			StateVersionID: sv.ID,
		})
		if err != nil {
			return nil, err
		}
		svo.CreatedAt = result.CreatedAt
		svo.UpdatedAt = result.UpdatedAt
	}

	return sv, tx.Commit(ctx)
}

func (s StateVersionService) List(opts otf.StateVersionListOptions) (*otf.StateVersionList, error) {
	if opts.Workspace == nil {
		return nil, fmt.Errorf("missing required option: workspace")
	}
	if opts.Organization == nil {
		return nil, fmt.Errorf("missing required option: organization")
	}

	q := NewQuerier(s.Conn)
	ctx := context.Background()

	result, err := q.FindStateVersionsByWorkspaceName(ctx, FindStateVersionsByWorkspaceNameParams{
		WorkspaceName:    *opts.Workspace,
		OrganizationName: *opts.Organization,
		Limit:            opts.GetLimit(),
		Offset:           opts.GetOffset(),
	})
	if err != nil {
		return nil, err
	}

	var items []*otf.StateVersion
	for _, r := range result {
		items = append(items, convertStateVersionList(r))
	}

	return &otf.StateVersionList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, getCount(result)),
	}, nil
}

func (s StateVersionService) Get(opts otf.StateVersionGetOptions) (*otf.StateVersion, error) {
	ctx := context.Background()
	q := NewQuerier(s.Conn)

	if opts.ID != nil {
		result, err := q.FindStateVersionByID(ctx, *opts.ID)
		if err != nil {
			return nil, err
		}
		return convertStateVersion(stateVersionRow(result)), nil
	} else if opts.WorkspaceID != nil {
		result, err := q.FindStateVersionLatestByWorkspaceID(ctx, *opts.WorkspaceID)
		if err != nil {
			return nil, err
		}
		return convertStateVersion(stateVersionRow(result)), nil
	} else {
		return nil, fmt.Errorf("no state version spec provided")
	}
}

// Delete deletes a state version from the DB
func (s StateVersionService) Delete(id string) error {
	ctx := context.Background()
	q := NewQuerier(s.Conn)

	result, err := q.DeleteStateVersionByID(ctx, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return nil
}
