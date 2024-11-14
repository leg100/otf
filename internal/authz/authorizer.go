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
	WorkspacePolicyGetter

	organizationResolvers map[resource.Kind]OrganizationResolver
	workspaceResolvers    map[resource.Kind]WorkspaceResolver
}

// Interface provides an interface for services to use to permit swapping out
// the authorizer for tests.
type Interface interface {
	// TODO: rename to Authorize
	CanAccess(ctx context.Context, action rbac.Action, req *AccessRequest) (Subject, error)
	// TODO: rename to CanAccess
	CanAccessDecision(ctx context.Context, action rbac.Action, req *AccessRequest) bool
}

func NewAuthorizer(logger logr.Logger) *Authorizer {
	return &Authorizer{
		Logger:                logger,
		organizationResolvers: make(map[resource.Kind]OrganizationResolver),
		workspaceResolvers:    make(map[resource.Kind]WorkspaceResolver),
	}
}

type WorkspacePolicyGetter interface {
	GetWorkspacePolicy(ctx context.Context, workspaceID resource.ID) (WorkspacePolicy, error)
}

// OrganizationResolver takes the ID of a resource and returns the name of the
// organization it belongs to.
type OrganizationResolver func(ctx context.Context, id resource.ID) (string, error)

// WorkspaceResolver takes the ID of a resource and returns the ID of the
// workspace it belongs to.
type WorkspaceResolver func(ctx context.Context, id resource.ID) (resource.ID, error)

// RegisterOrganizationResolver registers with the authorizer the ability to
// resolve access requests for a specific resource kind to the name of the
// organization the resource belongs to.
//
// This is necessary because authorization is determined not only on resource ID
// but on the name of the organization the resource belongs to.
func (a *Authorizer) RegisterOrganizationResolver(kind resource.Kind, resolver OrganizationResolver) {
	a.organizationResolvers[kind] = resolver
}

// RegisterWorkspaceResolver registers with the authorizer the ability to
// resolve access requests for a specific resource kind to the workspace ID the
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
	if req != nil && req.ID != nil {
		// Check if resource kind is registered for its ID to be resolved to workspace
		// ID.
		if resolver, ok := a.workspaceResolvers[req.ID.Kind()]; ok {
			workspaceID, err := resolver(ctx, *req.ID)
			if err != nil {
				a.Error(err, "authorization failure resolving workspace ID",
					"resource", req,
					"action", action.String(),
					"subject", subj,
				)
				return nil, err
			}
			// Authorize workspace ID instead
			req.ID = &workspaceID
		}
		// If the resource kind is a workspace, then fetch its policy.
		if req.ID.Kind() == resource.WorkspaceKind {
			policy, err := a.GetWorkspacePolicy(ctx, *req.ID)
			if err != nil {
				a.Error(err, "authorization failure fetching workspace policy",
					"resource", req,
					"action", action.String(),
					"subject", subj,
				)
				return nil, err
			}
			req.WorkspacePolicy = &policy
		}

		// Then check if the resource kind - including the case where the
		// resource kind has been resolved to a workspace - is registered for
		// its ID to be resolved to an oranization name. This is ony necessary
		// if the organization has not been specified in the access request.
		if req.Organization == "" {
			if resolver, ok := a.organizationResolvers[req.ID.Kind()]; ok {
				organization, err := resolver(ctx, *req.ID)
				if err != nil {
					a.Error(err, "authorization failure resolving organization",
						"resource", req,
						"action", action.String(),
						"subject", subj,
					)
					return nil, err
				}
				req.Organization = organization
			}
		}
	}
	// Allow subject to determine whether it is allowed to access resource.
	if subj.CanAccess(action, req) {
		return subj, nil
	}
	a.Error(nil, "unauthorized action",
		"resource", req,
		"action", action.String(),
		"subject", subj,
	)
	return nil, internal.ErrAccessNotPermitted
}

// CanAccessDecision is a helper to boil down an access request to a true/false
// decision, with any error encountered interpreted as false.
func (a *Authorizer) CanAccessDecision(ctx context.Context, action rbac.Action, req *AccessRequest) bool {
	_, err := a.CanAccess(ctx, action, req)
	return err == nil
}

// AccessRequest is a request for access to either an organization or an
// individual resource.
type AccessRequest struct {
	// Organization name to which access is being requested.
	Organization string
	// ID of resource to which access is being requested. If nil then the action
	// is being requested on the organization.
	ID *resource.ID
	// WorkspacePolicy specifies workspace-specific permissions for the resource
	// specified by ID. Only non-nil if ID refers to a workspace.
	WorkspacePolicy *WorkspacePolicy
}

// WorkspacePolicy binds workspace permissions to a workspace
type WorkspacePolicy struct {
	Permissions []WorkspacePermission
	// Whether workspace permits its state to be consumed by all workspaces in
	// the organization.
	GlobalRemoteState bool
}

// WorkspacePermission binds a role to a team.
type WorkspacePermission struct {
	TeamID resource.ID
	Role   rbac.Role
}

func (r *AccessRequest) LogValue() slog.Value {
	if r == nil {
		return slog.StringValue("site")
	} else {
		attrs := []slog.Attr{
			slog.String("organization", r.Organization),
		}
		if r.ID != nil {
			attrs = append(attrs, slog.String("resource_id", r.ID.String()))
		}
		return slog.GroupValue(attrs...)
	}
}
