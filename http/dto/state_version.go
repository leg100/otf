package dto

import (
	"time"
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

// StateVersionList represents a list of state versions.
type StateVersionList struct {
	*Pagination
	Items []*StateVersion
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
