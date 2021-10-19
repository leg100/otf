package otf

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
)

var (
	ErrInvalidStateVersionGetOptions = errors.New("invalid state version get options")
)

// StateVersion represents a Terraform Enterprise state version.
type StateVersion struct {
	ID string `db:"state_version_id"`

	Timestamps

	Serial       int64
	VCSCommitSHA string
	VCSCommitURL string

	// BlobID is ID of the binary object containing the state
	BlobID string

	// State version belongs to a workspace
	Workspace *Workspace `db:"workspaces"`

	// Run that created this state version. Optional.
	// Run     *Run

	// State version has many outputs
	Outputs []*StateVersionOutput `db:"state_version_outputs"`
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
	Create(sv *StateVersion) (*StateVersion, error)
	Get(opts StateVersionGetOptions) (*StateVersion, error)
	List(opts StateVersionListOptions) (*StateVersionList, error)
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

type StateVersionFactory struct {
	WorkspaceService WorkspaceService
	BlobStore        BlobStore
}

func (f *StateVersionFactory) NewStateVersion(workspaceID string, opts StateVersionCreateOptions) (*StateVersion, error) {
	sv := StateVersion{
		ID:         NewID("sv"),
		Timestamps: NewTimestamps(),
		Serial:     *opts.Serial,
	}

	ws, err := f.WorkspaceService.Get(context.Background(), WorkspaceSpecifier{ID: &workspaceID})
	if err != nil {
		return nil, err
	}
	sv.Workspace = ws

	decoded, err := base64.StdEncoding.DecodeString(*opts.State)
	if err != nil {
		return nil, err
	}

	sv.BlobID = NewBlobID()
	if err := f.BlobStore.Put(sv.BlobID, decoded); err != nil {
		return nil, err
	}

	state, err := Parse(decoded)
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

func (r *StateVersion) DownloadURL() string {
	return fmt.Sprintf("/state-versions/%s/download", r.ID)
}
