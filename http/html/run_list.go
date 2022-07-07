package html

import "github.com/leg100/otf"

// runList exposes a list of runs to a template
type runList struct {
	*otf.RunList
	opts otf.RunListOptions
}

// OrganizationName makes the organization name for a run listing availabe to a
// template
func (l runList) OrganizationName() string {
	return *l.opts.OrganizationName
}

// WorkspaceName makes the workspace name for a run listing availabel to a
// template.
func (l runList) WorkspaceName() string {
	return *l.opts.WorkspaceName
}
