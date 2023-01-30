package state

import (
	"errors"
	"fmt"
	"time"

	"github.com/leg100/otf"
)

// Version is a specific version of terraform state. It includes important
// metadata as well as the state data itself.
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/state-versions
type Version struct {
	id          string
	createdAt   time.Time
	serial      int64
	state       []byte                // json-encoded state file
	outputs     []*StateVersionOutput // State version has many outputs
	workspaceID string                // state version belongs to a workspace
}

func (v *Version) ID() string                     { return v.id }
func (v *Version) CreatedAt() time.Time           { return v.createdAt }
func (v *Version) String() string                 { return v.id }
func (v *Version) Serial() int64                  { return v.serial }
func (v *Version) State() []byte                  { return v.state }
func (v *Version) Outputs() []*StateVersionOutput { return v.outputs }

// ToJSONAPI assembles a struct suitable for marshalling into json-api
func (v *Version) ToJSONAPI() any {
	j := &jsonapiVersion{
		ID:          v.ID(),
		CreatedAt:   v.CreatedAt(),
		DownloadURL: fmt.Sprintf("/api/v2/state-versions/%s/download", v.ID()),
		Serial:      v.Serial(),
	}
	for _, out := range v.Outputs() {
		j.Outputs = append(j.Outputs, &jsonapiVersionOutput{
			ID:        out.ID(),
			Name:      out.Name,
			Sensitive: out.Sensitive,
			Type:      out.Type,
			Value:     out.Value,
		})
	}
	return j
}

// VersionList represents a list of state versions.
type VersionList struct {
	*otf.Pagination
	Items []*Version
}

// ToJSONAPI assembles a struct suitable for marshalling into json-api
func (l *VersionList) ToJSONAPI() any {
	jl := &jsonapiVersionList{
		Pagination: l.Pagination.ToJSONAPI(),
	}
	for _, item := range l.Items {
		jl.Items = append(jl.Items, item.ToJSONAPI().(*jsonapiVersion))
	}
	return jl
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
