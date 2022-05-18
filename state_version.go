package otf

import (
	"encoding/base64"
	"errors"
	"fmt"
)

var (
	ErrInvalidStateVersionGetOptions = errors.New("invalid state version get options")
)

// StateVersion represents a Terraform Enterprise state version.
type StateVersion struct {
	ID string

	Timestamps

	Serial       int64
	VCSCommitSHA string
	VCSCommitURL string

	// State is the state file itself.
	State []byte

	// State version has many outputs
	Outputs []*StateVersionOutput
}

func (sv *StateVersion) GetID() string  { return sv.ID }
func (sv *StateVersion) String() string { return sv.ID }

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

	// Force can be set to skip certain validations. Wrong use
	// of this flag can cause data loss, so USE WITH CAUTION!
	Force *bool `jsonapi:"attr,force"`

	// Specifies the run to associate the state with.
	Run *Run `jsonapi:"relation,run,omitempty"`
}

type StateVersionFactory struct{}

// Valid validates state version create options
func (opts *StateVersionCreateOptions) Valid() error {
	return nil
}

func (f *StateVersionFactory) NewStateVersion(opts StateVersionCreateOptions) (*StateVersion, error) {
	if err := opts.Valid(); err != nil {
		return nil, fmt.Errorf("invalid create options: %w", err)
	}

	sv := StateVersion{
		ID:     NewID("sv"),
		Serial: *opts.Serial,
	}

	var err error
	sv.State, err = base64.StdEncoding.DecodeString(*opts.State)
	if err != nil {
		return nil, err
	}

	state, err := Parse(sv.State)
	if err != nil {
		return nil, err
	}

	for k, v := range state.Outputs {
		sv.Outputs = append(sv.Outputs, &StateVersionOutput{
			ID:    NewID("wsout"),
			Name:  k,
			Type:  v.Type,
			Value: v.Value,
		})
	}

	return &sv, nil
}

func (sv *StateVersion) DownloadURL() string {
	return fmt.Sprintf("/state-versions/%s/download", sv.ID)
}
