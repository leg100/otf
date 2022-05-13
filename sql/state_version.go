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

type stateVersionRow interface {
	GetStateVersionID() *string
	GetSerial() *int32
	GetVcsCommitSha() *string
	GetVcsCommitUrl() *string
	GetState() []byte
	GetRunID() *string
	GetStateVersionOutputs() []StateVersionOutputs

	Timestamps
}

type stateVersionOutputRow interface {
	GetStateVersionOutputID() *string
	GetName() *string
	GetSensitive() *bool
	GetType() *string
	GetValue() *string
	GetStateVersionID() *string

	Timestamps
}

type stateVersionRowList interface {
	stateVersionRow

	GetFullCount() *int
}

func NewStateVersionDB(conn *pgx.Conn) *StateVersionService {
	return &StateVersionService{
		Conn: conn,
	}
}

// Create persists a StateVersion to the DB.
func (s StateVersionService) Create(sv *otf.StateVersion) (*otf.StateVersion, error) {
	q := NewQuerier(s.Conn)
	ctx := context.Background()

	_, err := q.InsertStateVersion(ctx, InsertStateVersionParams{
		ID:     sv.ID,
		Serial: int32(sv.Serial),
		State:  sv.State,
		RunID:  sv.Run.ID,
	})
	if err != nil {
		return nil, err
	}

	// Insert state_version_outputs
	for _, svo := range sv.Outputs {
		_, err = q.InsertStateVersionOutput(ctx, InsertStateVersionOutputParams{
			ID:             svo.ID,
			Name:           svo.Name,
			Sensitive:      svo.Sensitive,
			Type:           svo.Type,
			Value:          svo.Value,
			StateVersionID: svo.StateVersionID,
		})
		if err != nil {
			return nil, err
		}
	}

	// Return newly created state version to caller
	return getStateVersion(ctx, q, otf.StateVersionGetOptions{ID: &sv.ID})
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
		items = append(items, convertStateVersion(r))
	}

	return &otf.StateVersionList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, getCount(result)),
	}, nil
}

func (s StateVersionService) Get(opts otf.StateVersionGetOptions) (*otf.StateVersion, error) {
	ctx := context.Background()
	q := NewQuerier(s.Conn)

	return getStateVersion(ctx, q, opts)
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

func getStateVersion(ctx context.Context, q *DBQuerier, opts otf.StateVersionGetOptions) (*otf.StateVersion, error) {
	if opts.ID != nil {
		result, err := q.FindStateVersionByID(ctx, *opts.ID)
		if err != nil {
			return nil, err
		}
		return convertStateVersion(result), nil
	} else if opts.WorkspaceID != nil {
		result, err := q.FindStateVersionLatestByWorkspaceID(ctx, *opts.WorkspaceID)
		if err != nil {
			return nil, err
		}
		return convertStateVersion(result), nil
	} else {
		return nil, fmt.Errorf("no state version spec provided")
	}
}

func convertStateVersion(row stateVersionRow) *otf.StateVersion {
	sv := otf.StateVersion{
		ID:         *row.GetStateVersionID(),
		Timestamps: convertTimestamps(row),
		Serial:     int64(*row.GetSerial()),
		State:      row.GetState(),
	}

	for _, svo := range row.GetStateVersionOutputs() {
		sv.Outputs = append(sv.Outputs, convertStateVersionOutput(svo))
	}

	return &sv
}

func convertStateVersionOutput(row stateVersionOutputRow) *otf.StateVersionOutput {
	svo := otf.StateVersionOutput{
		ID:             *row.GetStateVersionID(),
		Sensitive:      *row.GetSensitive(),
		Timestamps:     convertTimestamps(row),
		Type:           *row.GetType(),
		Value:          *row.GetValue(),
		Name:           *row.GetName(),
		StateVersionID: *row.GetStateVersionID(),
	}

	return &svo
}
