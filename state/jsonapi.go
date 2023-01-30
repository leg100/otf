package state

import (
	"time"

	"github.com/leg100/otf/http/jsonapi"
)

// StateVersion represents a Terraform Enterprise state version.

// jsonapiVersion is a state version suitable for marshaling into JSONAPI
type jsonapiVersion struct {
	ID           string    `jsonapi:"primary,state-versions"`
	CreatedAt    time.Time `jsonapi:"attr,created-at,iso8601"`
	DownloadURL  string    `jsonapi:"attr,hosted-state-download-url"`
	Serial       int64     `jsonapi:"attr,serial"`
	VCSCommitSHA string    `jsonapi:"attr,vcs-commit-sha"`
	VCSCommitURL string    `jsonapi:"attr,vcs-commit-url"`

	// Relations
	Outputs []*jsonapiVersionOutput `jsonapi:"relation,outputs"`
}

// jsonapiVersionList is a list of state versions suitable for marshaling into
// JSONAPI
type jsonapiVersionList struct {
	*jsonapi.Pagination
	Items []*jsonapiVersion
}

type jsonapiVersionOutput struct {
	ID        string `jsonapi:"primary,state-version-outputs"`
	Name      string `jsonapi:"attr,name"`
	Sensitive bool   `jsonapi:"attr,sensitive"`
	Type      string `jsonapi:"attr,type"`
	Value     string `jsonapi:"attr,value"`
}

// jsonapiCreateVersionOptions are options for creating a state version via
// JSONAPI
type jsonapiCreateVersionOptions struct {
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
	// Run *Run `jsonapi:"relation,run,omitempty"`
}

func unmarshalJSONAPI(japi *jsonapiVersion) *Version {
	return &Version{
		id:        japi.ID,
		createdAt: japi.CreatedAt,
		serial:    japi.Serial,
		// TODO: unmarshal outputs
	}
}
