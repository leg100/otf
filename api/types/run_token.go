package types

type CreateRunTokenOptions struct {
	// Type is a public field utilized by JSON:API to set the resource type via
	// the field tag.  It is not a user-defined value and does not need to be
	// set.  https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,run_tokens"`

	// Organization is the organization of the run
	Organization *string `jsonapi:"attribute" json:"organization_name"`

	// RunID is the ID of the run for which the token is being created.
	RunID *string `jsonapi:"attribute" json:"run_id"`
}
