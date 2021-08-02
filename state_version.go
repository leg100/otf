package ots

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/leg100/go-tfe"
	"gorm.io/gorm"
)

var (
	ErrInvalidStateVersionGetOptions = errors.New("invalid state version get options")
)

// StateVersion represents a Terraform Enterprise state version.
type StateVersion struct {
	ID string

	gorm.Model

	Serial       int64
	VCSCommitSHA string
	VCSCommitURL string

	BlobID string

	// State version belongs to a workspace
	Workspace *Workspace

	// Run that created this state version. Optional.
	// Run     *Run

	Outputs []*StateVersionOutput

	// State version has many outputs
	StateVersionOutputs []StateVersionOutput
}

// StateVersionList represents a list of state versions.
type StateVersionList struct {
	*tfe.Pagination
	Items []*StateVersion
}

type StateVersionService interface {
	Create(workspaceID string, opts tfe.StateVersionCreateOptions) (*StateVersion, error)
	Current(workspaceID string) (*StateVersion, error)
	Get(id string) (*StateVersion, error)
	Download(id string) ([]byte, error)
	List(opts tfe.StateVersionListOptions) (*StateVersionList, error)
}

type StateVersionStore interface {
	Create(sv *StateVersion) (*StateVersion, error)
	Get(opts StateVersionGetOptions) (*StateVersion, error)
	List(opts tfe.StateVersionListOptions) (*StateVersionList, error)
}

// StateVersionGetOptions are options for retrieving a single StateVersion.
// Either ID *or* WorkspaceID must be specfiied.
type StateVersionGetOptions struct {
	// ID of state version to retrieve
	ID *string

	// Get current state version belonging to workspace with this ID
	WorkspaceID *string
}

type StateVersionFactory struct {
	WorkspaceService WorkspaceService
	BlobStore        BlobStore
}

func NewStateVersionID() string {
	return fmt.Sprintf("sv-%s", GenerateRandomString(16))
}

func (f *StateVersionFactory) NewStateVersion(workspaceID string, opts tfe.StateVersionCreateOptions) (*StateVersion, error) {
	sv := StateVersion{
		Serial: *opts.Serial,
		ID:     NewStateVersionID(),
	}

	ws, err := f.WorkspaceService.GetByID(workspaceID)
	if err != nil {
		return nil, err
	}
	sv.Workspace = ws

	decoded, err := base64.StdEncoding.DecodeString(*opts.State)
	if err != nil {
		return nil, err
	}

	blobID, err := f.BlobStore.Put(decoded)
	if err != nil {
		return nil, err
	}
	sv.BlobID = blobID

	state, err := Parse(decoded)
	if err != nil {
		return nil, err
	}

	for k, v := range state.Outputs {
		sv.Outputs = append(sv.Outputs, &StateVersionOutput{
			ID:    NewStateVersionOutputID(),
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
