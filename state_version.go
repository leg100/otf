package otf

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgtype"
	jsonapi "github.com/leg100/otf/http/dto"
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

// ToJSONAPI assembles a JSON-API DTO.
func (sv *StateVersion) ToJSONAPI(req *http.Request) any {
	dto := &jsonapi.StateVersion{
		ID:          sv.ID(),
		CreatedAt:   sv.CreatedAt(),
		DownloadURL: fmt.Sprintf("/state-versions/%s/download", sv.ID()),
		Serial:      sv.Serial(),
	}
	for _, out := range sv.Outputs() {
		dto.Outputs = append(dto.Outputs, &jsonapi.StateVersionOutput{
			ID:        out.ID(),
			Name:      out.Name,
			Sensitive: out.Sensitive,
			Type:      out.Type,
			Value:     out.Value,
		})
	}
	return dto
}

// StateVersionList represents a list of state versions.
type StateVersionList struct {
	*Pagination
	Items []*StateVersion
}

// ToJSONAPI assembles a JSON-API DTO.
func (l *StateVersionList) ToJSONAPI(req *http.Request) any {
	obj := &jsonapi.StateVersionList{
		Pagination: (*jsonapi.Pagination)(l.Pagination),
	}
	for _, item := range l.Items {
		obj.Items = append(obj.Items, item.ToJSONAPI(req).(*jsonapi.StateVersion))
	}
	return obj
}

type StateVersionService interface {
	Create(ctx context.Context, workspaceID string, opts StateVersionCreateOptions) (*StateVersion, error)
	Current(ctx context.Context, workspaceID string) (*StateVersion, error)
	Get(ctx context.Context, id string) (*StateVersion, error)
	Download(ctx context.Context, id string) ([]byte, error)
	List(ctx context.Context, opts StateVersionListOptions) (*StateVersionList, error)
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
	ListOptions
	Organization *string `schema:"filter[organization][name]"`
	Workspace    *string `schema:"filter[workspace][name]"`
}

// LogFields provides fields for logging
func (opts StateVersionListOptions) LogFields() (fields []interface{}) {
	if opts.Workspace != nil {
		fields = append(fields, "workspace", *opts.Workspace)
	}
	if opts.Organization != nil {
		fields = append(fields, "organization", *opts.Organization)
	}
	return fields
}

// StateVersionCreateOptions represents the options for creating a state
// version. See dto.StateVersionCreateOptions for more details.
type StateVersionCreateOptions struct {
	Lineage *string
	Serial  *int64
	State   *string
	MD5     *string
	Run     *Run
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
	StateVersionID      pgtype.Text                 `json:"state_version_id"`
	CreatedAt           time.Time                   `json:"created_at"`
	Serial              int                         `json:"serial"`
	State               []byte                      `json:"state"`
	WorkspaceID         pgtype.Text                 `json:"workspace_id"`
	StateVersionOutputs []pggen.StateVersionOutputs `json:"state_version_outputs"`
}

// UnmarshalStateVersionDBResult unmarshals a state version postgres record.
func UnmarshalStateVersionDBResult(row StateVersionDBRow) (*StateVersion, error) {
	sv := StateVersion{
		id:        row.StateVersionID.String,
		createdAt: row.CreatedAt,
		serial:    int64(row.Serial),
		state:     row.State,
	}
	for _, r := range row.StateVersionOutputs {
		sv.outputs = append(sv.outputs, UnmarshalStateVersionOutputDBType(r))
	}
	return &sv, nil
}
