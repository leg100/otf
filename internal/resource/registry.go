package resource

import (
	"context"
	"fmt"
)

type Registry struct {
	parentResolvers map[Kind]ParentResolver
}

type ParentResolver func(ctx context.Context, id ID) (ID, error)

// RegisterParentResolver registers a means of resolving a resource's parent
// resource, e.g. the workspace of a run.
func (a *Registry) RegisterParentResolver(kind Kind, resolver ParentResolver) {
	a.parentResolvers[kind] = resolver
}

// GetParentOrganization retrieves the ID of the parent organization of the
// resource with the given ID. If the resource does not belong to an
// organization an error is returned.
func (a *Registry) GetParentOrganizationID(ctx context.Context, id ID) (ID, error) {
	return a.getParentKind(ctx, id, OrganizationKind)
}

// GetParentWorkspace retrieves the ID of the parent workspace of the resource
// with the given ID. If the resource does not belong to a workspace then an
// error is returned.
func (a *Registry) GetParentWorkspaceID(ctx context.Context, id ID) (ID, error) {
	return a.getParentKind(ctx, id, WorkspaceKind)
}

func (a *Registry) getParentKind(ctx context.Context, id ID, parentKind Kind) (ID, error) {
	for {
		resolver, ok := a.parentResolvers[id.Kind()]
		if !ok {
			break
		}
		parentID, err := resolver(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("resolving parent of resource %s: %w", id, err)
		}
		if parentID.Kind() == parentKind {
			return parentID, nil
		}
		// now try looking up parent of parent
		id = parentID
	}
	return nil, fmt.Errorf("no parent of kind %s found for resource: %s", parentKind, id)
}
