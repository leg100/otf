package ots

import (
	"fmt"
	"time"

	"github.com/google/jsonapi"
	tfe "github.com/hashicorp/go-tfe"
)

// StateVersion represents a Terraform Enterprise state version.
type StateVersion struct {
	ID           string    `jsonapi:"primary,state-versions"`
	CreatedAt    time.Time `jsonapi:"attr,created-at,iso8601"`
	DownloadURL  string    `jsonapi:"attr,hosted-state-download-url"`
	Serial       int64     `jsonapi:"attr,serial"`
	VCSCommitSHA string    `jsonapi:"attr,vcs-commit-sha"`
	VCSCommitURL string    `jsonapi:"attr,vcs-commit-url"`

	// Relations
	Run     *Run                  `jsonapi:"relation,run"`
	Outputs []*StateVersionOutput `jsonapi:"relation,outputs"`
}

type StateVersionList struct {
	*Pagination
	Items     []*StateVersion
	Workspace string
}

// StateVersionListOptions represents the options for listing state versions.
type StateVersionListOptions struct {
	ListOptions

	Organization *string `url:"filter[organization][name]"`
	Workspace    *string `url:"filter[workspace][name]"`
}

var _ Paginated = (*StateVersionList)(nil)

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
	// Run *Run `jsonapi:"relation,run,omitempty"`
}

type StateVersionService interface {
	CreateStateVersion(id string, opts *tfe.StateVersionCreateOptions) (*StateVersion, error)
	CurrentStateVersion(id string) (*StateVersion, error)
	GetStateVersion(id string) (*StateVersion, error)
	ListStateVersions() (*StateVersionList, error)
}

func (ws *StateVersion) JSONAPILinks() *jsonapi.Links {
	return &jsonapi.Links{
		"self": fmt.Sprintf("/v2/api/organizations/%s/workspaces/%s", ws.Organization.Name, ws.Name),
	}
}

func (l *StateVersionList) GetItems() interface{} {
	return l.Items
}

func (l *StateVersionList) GetPath(org string) string {
	return fmt.Sprintf("/v2/api/organizations/%s/workspaces", org)
}

func NewStateVersionID() string {
	return fmt.Sprintf("sv-%s", GenerateRandomString(16))
}
