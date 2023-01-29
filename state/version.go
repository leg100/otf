package state

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/dto"
	jsonapi "github.com/leg100/otf/http/dto"
	"github.com/leg100/otf/sql/pggen"
)

// StateVersion represents a Terraform Enterprise state version.
type StateVersion struct {
	id          string
	createdAt   time.Time
	serial      int64
	state       []byte                // json-encoded state file
	outputs     []*StateVersionOutput // State version has many outputs
	workspaceID string                // state version belongs to a workspace
}

func (sv *StateVersion) ID() string                     { return sv.id }
func (sv *StateVersion) CreatedAt() time.Time           { return sv.createdAt }
func (sv *StateVersion) String() string                 { return sv.id }
func (sv *StateVersion) Serial() int64                  { return sv.serial }
func (sv *StateVersion) State() []byte                  { return sv.state }
func (sv *StateVersion) Outputs() []*StateVersionOutput { return sv.outputs }

// ToJSONAPI assembles a JSON-API DTO.
func (sv *StateVersion) ToJSONAPI() any {
	obj := &dto.StateVersion{
		ID:          sv.ID(),
		CreatedAt:   sv.CreatedAt(),
		DownloadURL: fmt.Sprintf("/api/v2/state-versions/%s/download", sv.ID()),
		Serial:      sv.Serial(),
	}
	for _, out := range sv.Outputs() {
		obj.Outputs = append(obj.Outputs, &dto.StateVersionOutput{
			ID:        out.ID(),
			Name:      out.Name,
			Sensitive: out.Sensitive,
			Type:      out.Type,
			Value:     out.Value,
		})
	}
	return obj
}

// StateVersionList represents a list of state versions.
type StateVersionList struct {
	*otf.Pagination
	Items []*StateVersion
}

// ToJSONAPI assembles a JSON-API DTO.
func (l *StateVersionList) ToJSONAPI() any {
	obj := &dto.StateVersionList{
		Pagination: l.Pagination.ToJSONAPI(),
	}
	for _, item := range l.Items {
		obj.Items = append(obj.Items, (&StateVersion{item}).ToJSONAPI().(*dto.StateVersion))
	}
	return obj
}

type StateVersionService interface {
	CreateStateVersion(ctx context.Context, workspaceID string, opts StateVersionCreateOptions) (*StateVersion, error)
	CurrentStateVersion(ctx context.Context, workspaceID string) (*StateVersion, error)
	GetStateVersion(ctx context.Context, id string) (*StateVersion, error)
	DownloadState(ctx context.Context, id string) ([]byte, error)
	ListStateVersions(ctx context.Context, opts StateVersionListOptions) (*StateVersionList, error)
}

type StateVersionStore interface {
	CreateStateVersion(ctx context.Context, workspaceID string, sv *StateVersion) error
	GetStateVersion(ctx context.Context, opts StateVersionGetOptions) (*StateVersion, error)
	GetState(ctx context.Context, id string) ([]byte, error)
	ListStateVersions(ctx context.Context, opts StateVersionListOptions) (*StateVersionList, error)
	DeleteStateVersion(ctx context.Context, id string) error
}

// StateVersionGetOptions are options for retrieving a single StateVersion.
// Either ID *or* WorkspaceID must be specfiied.
type StateVersionGetOptions struct {
	// ID of state version to retrieve
	ID *string
	// Get current state version belonging to workspace with this ID
	WorkspaceID *string
}

// StateVersionListOptions represents the options for listing state versions.
type StateVersionListOptions struct {
	otf.ListOptions
	Organization string `schema:"filter[organization][name],required"`
	Workspace    string `schema:"filter[workspace][name],required"`
}

// StateVersionCreateOptions represents the options for creating a state
// version. See dto.StateVersionCreateOptions for more details.
type StateVersionCreateOptions struct {
	Lineage *string
	Serial  *int64
	State   *string
	MD5     *string
	Run     *otf.Run
}

// NewStateVersion constructs a new state version.
func NewStateVersion(opts otf.CreateStateVersionOptions) (*StateVersion, error) {
	if opts.State == nil {
		return nil, errors.New("state file required")
	}
	if opts.WorkspaceID == nil {
		return nil, errors.New("workspace ID required")
	}

	state, err := otf.UnmarshalState(opts.State)
	if err != nil {
		return nil, err
	}
	sv := StateVersion{
		id:          otf.NewID("sv"),
		createdAt:   otf.CurrentTimestamp(),
		serial:      state.Serial,
		state:       opts.State,
		workspaceID: *opts.WorkspaceID,
	}
	for k, v := range state.Outputs {
		sv.outputs = append(sv.outputs, &StateVersionOutput{
			id:    otf.NewID("wsout"),
			Name:  k,
			Type:  v.Type,
			Value: v.Value,
		})
	}
	return &sv, nil
}

// StateVersionResult represents the result of a database query for a state
// version.
type StateVersionResult struct {
	StateVersionID      pgtype.Text                 `json:"state_version_id"`
	CreatedAt           pgtype.Timestamptz          `json:"created_at"`
	Serial              int                         `json:"serial"`
	State               []byte                      `json:"state"`
	WorkspaceID         pgtype.Text                 `json:"workspace_id"`
	StateVersionOutputs []pggen.StateVersionOutputs `json:"state_version_outputs"`
}

// UnmarshalStateVersionResult unmarshals a database result query into a state version.
func UnmarshalStateVersionResult(row StateVersionResult) (*StateVersion, error) {
	sv := StateVersion{
		id:        row.StateVersionID.String,
		createdAt: row.CreatedAt.Time.UTC(),
		serial:    int64(row.Serial),
		state:     row.State,
	}
	for _, r := range row.StateVersionOutputs {
		sv.outputs = append(sv.outputs, UnmarshalStateVersionOutputRow(r))
	}
	return &sv, nil
}

func UnmarshalStateVersionJSONAPI(dto *jsonapi.StateVersion) *StateVersion {
	return &StateVersion{
		id:        dto.ID,
		createdAt: dto.CreatedAt,
		serial:    dto.Serial,
		// TODO: unmarshal outputs
	}
}
