package authz

import "github.com/leg100/otf/internal/resource"

// Request for authorization.
type Request struct {
	// ID of resource to which access is being requested.
	resource.ID
	// WorkspacePolicy provides a means of checking workspace-specific
	// permissions for the resource specified by the ID above. If this is nil
	// then the resource is not a workspace or does not belong to a workspace.
	WorkspacePolicy WorkspacePolicy
	// lineage is the direct line of ancestors of the resource, starting with
	// the nearest ancestor (its parent) first.
	lineage []resource.ID
}

// Organization identifies the organization that the requested resource belongs
// to, or the organization itself if access to an organization is being
// requested.
func (r Request) Organization() resource.ID {
	if r.Kind() == resource.OrganizationKind {
		return r.ID
	}
	for _, id := range r.lineage {
		if id.Kind() == resource.OrganizationKind {
			return id
		}
	}
	return nil
}

// Workspace identifies the workspace that the requested resource belongs to, or
// the workspace itself if access to an workspace is being requested.
func (r Request) Workspace() resource.ID {
	if r.Kind() == resource.WorkspaceKind {
		return r.ID
	}
	for _, id := range r.lineage {
		if id.Kind() == resource.WorkspaceKind {
			return id
		}
	}
	return nil
}
