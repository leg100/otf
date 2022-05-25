package otf

import (
	"encoding/base64"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/leg100/otf/sql/pggen"
	"github.com/stretchr/testify/require"
)

var (
	ErrInvalidStateVersionGetOptions = errors.New("invalid state version get options")
)

// StateVersion represents a Terraform Enterprise state version.
type StateVersion struct {
	id        string
	createdAt time.Time
	serial    int64
	// state is the state file in json.
	state []byte
	// State version has many outputs
	outputs []*StateVersionOutput
}

func (sv *StateVersion) ID() string                     { return sv.id }
func (sv *StateVersion) CreatedAt() time.Time           { return sv.createdAt }
func (sv *StateVersion) String() string                 { return sv.id }
func (sv *StateVersion) Serial() int64                  { return sv.serial }
func (sv *StateVersion) State() []byte                  { return sv.state }
func (sv *StateVersion) Outputs() []*StateVersionOutput { return sv.outputs }

// StateVersionList represents a list of state versions.
type StateVersionList struct {
	*Pagination
	Items []*StateVersion
}

type StateVersionService interface {
	Create(workspaceID string, opts StateVersionCreateOptions) (*StateVersion, error)
	Current(workspaceID string) (*StateVersion, error)
	Get(id string) (*StateVersion, error)
	Download(id string) ([]byte, error)
	List(opts StateVersionListOptions) (*StateVersionList, error)
}

type StateVersionStore interface {
	Create(workspaceID string, sv *StateVersion) error
	Get(opts StateVersionGetOptions) (*StateVersion, error)
	GetState(id string) ([]byte, error)
	List(opts StateVersionListOptions) (*StateVersionList, error)
	Delete(id string) error
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
	ListOptions
	Organization *string `schema:"filter[organization][name]"`
	Workspace    *string `schema:"filter[workspace][name]"`
}

// StateVersionCreateOptions represents the options for creating a state
// version.
type StateVersionCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,state-versions"`
	// The lineage of the state.
	Lineage *string `jsonapi:"attr,lineage,omitempty"`
	// The MD5 hash of the state version.
	MD5 *string `jsonapi:"attr,md5"`
	// The serial of the state.
	Serial *int64 `jsonapi:"attr,serial"`
	// The base64 encoded state.
	State *string `jsonapi:"attr,state"`
	// Force can be set to skip certain validations. Wrong use of this flag can
	// cause data loss, so USE WITH CAUTION!
	Force *bool `jsonapi:"attr,force"`
	// Specifies the run to associate the state with.
	Run *Run `jsonapi:"relation,run,omitempty"`
}

// Valid validates state version create options
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
		id:        NewID("sv"),
		serial:    *opts.Serial,
		createdAt: CurrentTimestamp(),
		state:     decoded,
	}
	state, err := UnmarshalState(decoded)
	if err != nil {
		return nil, err
	}
	for k, v := range state.Outputs {
		sv.outputs = append(sv.outputs, &StateVersionOutput{
			id:    NewID("wsout"),
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
		Serial: Int64(1),
		State:  &encoded,
	})
	require.NoError(t, err)
	return sv
}

// StateVersionDBRow is the state version postgres record.
type StateVersionDBRow struct {
	StateVersionID      string                      `json:"state_version_id"`
	CreatedAt           time.Time                   `json:"created_at"`
	Serial              int                         `json:"serial"`
	State               []byte                      `json:"state"`
	WorkspaceID         string                      `json:"workspace_id"`
	StateVersionOutputs []pggen.StateVersionOutputs `json:"state_version_outputs"`
}

// UnmarshalStateVersionDBResult unmarshals a state version postgres record.
func UnmarshalStateVersionDBResult(row StateVersionDBRow) (*StateVersion, error) {
	sv := StateVersion{
		id:        row.StateVersionID,
		createdAt: row.CreatedAt,
		serial:    int64(row.Serial),
		state:     row.State,
	}
	for _, r := range row.StateVersionOutputs {
		sv.outputs = append(sv.outputs, UnmarshalStateVersionOutputDBType(r))
	}
	return &sv, nil
}
