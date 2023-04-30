package types

import "time"

// StateVersion is a state version suitable for marshaling into JSONAPI
type StateVersion struct {
	ID           string    `jsonapi:"primary,state-versions"`
	CreatedAt    time.Time `jsonapi:"attribute" json:"created-at"`
	DownloadURL  string    `jsonapi:"attribute" json:"hosted-state-download-url"`
	Serial       int64     `jsonapi:"attribute" json:"serial"`
	VCSCommitSHA string    `jsonapi:"attribute" json:"vcs-commit-sha"`
	VCSCommitURL string    `jsonapi:"attribute" json:"vcs-commit-url"`

	// Relations
	Outputs []*StateVersionOutput `jsonapi:"relationship" json:"outputs"`
}

// StateVersionList is a list of state versions suitable for marshaling into
// JSONAPI
type StateVersionList struct {
	*Pagination
	Items []*StateVersion
}

type StateVersionOutput struct {
	ID        string `jsonapi:"primary,state-version-outputs"`
	Name      string `jsonapi:"attribute" json:"name"`
	Sensitive bool   `jsonapi:"attribute" json:"sensitive"`
	Type      string `jsonapi:"attribute" json:"type"`
	Value     string `jsonapi:"attribute" json:"value"`
}

// StateVersionOutputList is a list of state version outputs suitable for marshaling into
// JSONAPI
type StateVersionOutputList struct {
	Items []*StateVersionOutput
}

// StateVersionCreateVersionOptions are options for creating a state version via
// JSONAPI
type StateVersionCreateVersionOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,state-versions"`
	// The lineage of the state.
	Lineage *string `jsonapi:"attribute" json:"lineage,omitempty"`
	// The MD5 hash of the state version.
	MD5 *string `jsonapi:"attribute" json:"md5"`
	// The serial of the state.
	Serial *int64 `jsonapi:"attribute" json:"serial"`
	// The base64 encoded state.
	State *string `jsonapi:"attribute" json:"state"`
	// Force can be set to skip certain validations. Wrong use of this flag can
	// cause data loss, so USE WITH CAUTION!
	Force *bool `jsonapi:"attribute" json:"force"`
	// Specifies the run to associate the state with.
	// Run *Run `jsonapi:"relationship" json:"run,omitempty"`
}

// RollbackStateVersionOptions are options for rolling back a state version
type RollbackStateVersionOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,state-versions"`
	// Specifies state version to rollback to. Only its ID is specified.
	RollbackStateVersion *StateVersion `jsonapi:"relationship" json:"state-version"`
}
