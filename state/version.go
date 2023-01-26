package state

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	jsonapi "github.com/leg100/otf/http/dto"
	"github.com/leg100/otf/sql/pggen"
	"github.com/stretchr/testify/require"
)

// StateVersion represents a Terraform Enterprise state version.
type StateVersion struct {
	id        string
	createdAt time.Time
	serial    int64
	state     []byte                // json-encoded state file
	outputs   []*StateVersionOutput // State version has many outputs
}

func (sv *StateVersion) ID() string                     { return sv.id }
func (sv *StateVersion) CreatedAt() time.Time           { return sv.createdAt }
func (sv *StateVersion) String() string                 { return sv.id }
func (sv *StateVersion) Serial() int64                  { return sv.serial }
func (sv *StateVersion) State() []byte                  { return sv.state }
func (sv *StateVersion) Outputs() []*StateVersionOutput { return sv.outputs }

// StateVersionList represents a list of state versions.
type StateVersionList struct {
	*otf.Pagination
	Items []*StateVersion
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

// Valid validates state version create options
//
// TODO: perform validation, check md5, etc
func (opts *StateVersionCreateOptions) Valid() error {
	return nil
}

// NewStateVersion constructs a new state version.
func NewStateVersion(opts StateVersionCreateOptions) (*StateVersion, error) {
	if err := opts.Valid(); err != nil {
		return nil, fmt.Errorf("invalid create options: %w", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(*opts.State)
	if err != nil {
		return nil, err
	}
	sv := StateVersion{
		id:        otf.NewID("sv"),
		serial:    *opts.Serial,
		createdAt: otf.CurrentTimestamp(),
		state:     decoded,
	}
	state, err := otf.UnmarshalState(decoded)
	if err != nil {
		return nil, err
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

func NewTestStateVersion(t *testing.T, outputs ...StateOutput) *StateVersion {
	state := NewState(StateCreateOptions{}, outputs...)
	encoded, err := state.Marshal()
	require.NoError(t, err)

	sv, err := NewStateVersion(StateVersionCreateOptions{
		Serial: otf.Int64(1),
		State:  &encoded,
	})
	require.NoError(t, err)
	return sv
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
