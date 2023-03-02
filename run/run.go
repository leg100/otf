package run

import (
	"github.com/leg100/otf"
)

// RunListOptions are options for paginating and filtering a list of runs
type RunListOptions struct {
	otf.ListOptions
	// Filter by run statuses (with an implicit OR condition)
	Statuses []otf.RunStatus `schema:"statuses,omitempty"`
	// Filter by workspace ID
	WorkspaceID *string `schema:"workspace_id,omitempty"`
	// Filter by organization name
	Organization *string `schema:"organization_name,omitempty"`
	// Filter by workspace name
	WorkspaceName *string `schema:"workspace_name,omitempty"`
	// Filter by speculative or non-speculative
	Speculative *bool `schema:"-"`
	// A list of relations to include. See available resources:
	// https://www.terraform.io/docs/cloud/api/run.html#available-related-resources
	Include *string `schema:"include,omitempty"`
}
