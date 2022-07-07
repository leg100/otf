package html

import (
	"github.com/leg100/otf"
)

// workspaceList exposes a list of workspaces to a template
type workspaceList struct {
	*otf.WorkspaceList
	otf.WorkspaceListOptions
}

// OrganizationName makes the organization name for a workspace listing
// available to a template
func (l workspaceList) OrganizationName() string {
	return *l.WorkspaceListOptions.OrganizationName
}
