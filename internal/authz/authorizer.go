package authz

import (
	"context"
	"log/slog"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
)

// Authorizer intermediates authorization between subjects (entities requesting
// access) and resources (the entities to which access is being requested).
type Authorizer struct {
	logr.Logger

	resourceAuthorizers   map[resource.Kind]ResourceAuthorizer
	organizationResolvers map[resource.Kind]OrganizationResolver
	workspaceResolvers    map[resource.Kind]WorkspaceResolver
}

type ResourceAuthorizer func(ctx context.Context, action rbac.Action, id resource.ID) (bool, error)

// OrganizationResolver takes the ID of a resource and returns the name of the
// organization it belongs to.
type OrganizationResolver func(ctx context.Context, id resource.ID) (string, error)

// WorkspaceResolver takes the ID of a resource and returns the ID of the
// workspace it belongs to.
type WorkspaceResolver func(ctx context.Context, id resource.ID) (resource.ID, error)

// RegisterAuthorizer allows authorization to be determined for a specific
// resource kind. i.e. a workspace policy determining which teams can carry out actions on a
// workspace.
func (a *Authorizer) RegisterAuthorizer(kind resource.Kind, authorizer ResourceAuthorizer) {
	a.resourceAuthorizers[kind] = authorizer
}

// RegisterWorkspaceResolver registers with the authorizer the ability to
// resolve access requests to a specific resource kind to the workspace ID the
// resource belongs to.
//
// This is necessary because authorization is often determined based on
// workspace ID, and not the ID of a run, state version, etc.
func (a *Authorizer) RegisterWorkspaceResolver(kind resource.Kind, resolver WorkspaceResolver) {
	a.workspaceResolvers[kind] = resolver
}

// CanAccess determines whether the subject can carry out an action on a
// resource. The subject is expected to be contained within the context. If the
// access request is nil then it's assumed the request is for access to the
// entire site (the highest level).
//
// Authorization in OTF works as follows:
// (i) if the subject allows access then it is allowed.
// (ii) otherwise specific resource kinds can allow access to subject.
// (iii) otherwise access is denied.
func (a *Authorizer) CanAccess(ctx context.Context, action rbac.Action, req *AccessRequest) (Subject, error) {
	subj, err := SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	// Allow context to contain specific instruction to skip authorization.
	// Should only be used for testing purposes.
	if SkipAuthz(ctx) {
		return subj, nil
	}
	if req.ID != nil {
		// Check if resource kind is registered for its ID to be resolved to workspace
		// ID.
		if resolver, ok := a.workspaceResolvers[req.ID.Kind()]; ok {
			workspaceID, err := resolver(ctx, *req.ID)
			if err != nil {
				a.Error(err, "authorization failure",
					"resource", req,
					"action", action.String(),
					"subject", subj,
				)
				return nil, err
			}
			// Authorize workspace ID instead
			req.ID = &workspaceID
		}
	}
	// Allow subject to determine whether it is allowed to access resource.
	if subj.CanAccess(action, req) {
		return subj, nil
	}
	// Subject hasn't explicitly allowed access so delegate authorization to a
	// resource kind's specific authorizer, if one has been registered.
	if req.ID != nil {
		if authorizer, ok := a.resourceAuthorizers[req.ID.Kind()]; ok {
			allowed, err := authorizer(ctx, action, *req.ID)
			if err != nil {
				a.Error(err, "authorization failure",
					"resource", req,
					"action", action.String(),
					"subject", subj,
				)
				return nil, err
			}
			if allowed {
				return subj, nil
			}
		}
	}
	a.Error(nil, "unauthorized action",
		"resource", req,
		"action", action.String(),
		"subject", subj,
	)
	return nil, internal.ErrAccessNotPermitted
}

func (a *Authorizer) CanAccessDecision(ctx context.Context, action rbac.Action, req *AccessRequest) bool {
	_, err := a.CanAccess(ctx, action, req)
	return err != nil
}

// AccessRequest is a request for access to either an organization or an
// individual resource.
type AccessRequest struct {
	// Organization name to which access is being requested.
	Organization string
	// ID of resource to which access is being requested. If nil then the action
	// is being requested on the organization.
	ID *resource.ID
}

func (r *AccessRequest) LogValue() slog.Value {
	// Compose error message
	var resource string
	if r == nil {
		resource = "site"
	} else if r.Organization != nil {
		resource = *r.Organization
	} else if r.ID != nil {
		resource = r.ID.String()
	}
	return slog.GroupValue(slog.String("resource", resource))
}
