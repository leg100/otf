package workspace

import "github.com/leg100/otf"

// WorkspaceList represents a list of Workspaces.
type WorkspaceList struct {
	*otf.Pagination
	Items []*Workspace
}

// WorkspaceListOptions are options for paginating and filtering a list of
// Workspaces
type WorkspaceListOptions struct {
	// Pagination
	otf.ListOptions
	// Filter workspaces with name matching prefix.
	Prefix string `schema:"search[name],omitempty"`
	// Organization filters workspaces by organization name.
	Organization *string `schema:"organization_name,omitempty"`
	// Filter by those for which user has workspace-level permissions.
	UserID *string
}
