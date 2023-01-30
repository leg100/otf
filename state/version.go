package state

import (
	"errors"
	"fmt"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

// Version represents a Terraform Enterprise state version.
type Version struct {
	id          string
	createdAt   time.Time
	serial      int64
	state       []byte                // json-encoded state file
	outputs     []*StateVersionOutput // State version has many outputs
	workspaceID string                // state version belongs to a workspace
}

func (sv *Version) ID() string                     { return sv.id }
func (sv *Version) CreatedAt() time.Time           { return sv.createdAt }
func (sv *Version) String() string                 { return sv.id }
func (sv *Version) Serial() int64                  { return sv.serial }
func (sv *Version) State() []byte                  { return sv.state }
func (sv *Version) Outputs() []*StateVersionOutput { return sv.outputs }

// ToJSONAPI assembles a JSON-API DTO.
func (sv *Version) ToJSONAPI() any {
	obj := &jsonapiVersion{
		ID:          sv.ID(),
		CreatedAt:   sv.CreatedAt(),
		DownloadURL: fmt.Sprintf("/api/v2/state-versions/%s/download", sv.ID()),
		Serial:      sv.Serial(),
	}
	for _, out := range sv.Outputs() {
		obj.Outputs = append(obj.Outputs, &jsonapiVersionOutput{
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
	Items []*Version
}

// ToJSONAPI assembles a JSON-API DTO.
func (l *StateVersionList) ToJSONAPI() any {
	obj := &jsonapiVersionList{
		Pagination: l.Pagination.ToJSONAPI(),
	}
	for _, item := range l.Items {
		obj.Items = append(obj.Items, item.ToJSONAPI().(*jsonapiVersion))
	}
	return obj
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
func NewStateVersion(opts otf.CreateStateVersionOptions) (*Version, error) {
	if opts.State == nil {
		return nil, errors.New("state file required")
	}
	if opts.WorkspaceID == nil {
		return nil, errors.New("workspace ID required")
	}

	state, err := unmarshalState(opts.State)
	if err != nil {
		return nil, err
	}
	sv := Version{
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

func UnmarshalStateVersionJSONAPI(dto *jsonapi.StateVersion) *Version {
	return &Version{
		id:        dto.ID,
		createdAt: dto.CreatedAt,
		serial:    dto.Serial,
		// TODO: unmarshal outputs
	}
}
