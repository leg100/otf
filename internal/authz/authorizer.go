package authz

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
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
	Authorize(ctx context.Context, action Action, id resource.ID, opts ...CanAccessOption) (Subject, error)
	CanAccess(ctx context.Context, action Action, id resource.ID) bool
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
type OrganizationResolver func(ctx context.Context, id resource.TfeID) (resource.ID, error)

// WorkspaceResolver takes the ID of a resource and returns the ID of the
// workspace it belongs to.
type WorkspaceResolver func(ctx context.Context, id resource.TfeID) (resource.TfeID, error)

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

// Options for configuring the individual calls of CanAccess.

type CanAccessOption func(*canAccessConfig)

// WithoutErrorLogging disables logging an unauthorized error. This can be
// useful if just checking if a user can do something.
func WithoutErrorLogging() CanAccessOption {
	return func(cfg *canAccessConfig) {
		cfg.disableLogs = true
	}
}

type canAccessConfig struct {
	disableLogs bool
}

// Authorize determines whether the subject can carry out an action on a
// resource. The subject is expected to be contained within the context. If the
// access request is nil then it's assumed the request is for access to the
// entire site (the highest level).
func (a *Authorizer) Authorize(ctx context.Context, action Action, resourceID resource.ID, opts ...CanAccessOption) (Subject, error) {
	if resourceID == nil {
		return nil, errors.New("authorization request resourceID parameter cannot be nil")
	}
	var cfg canAccessConfig
	for _, fn := range opts {
		fn(&cfg)
	}
	subj, err := SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	// Allow context to contain specific instruction to skip authorization.
	// Should only be used for testing purposes.
	if SkipAuthz(ctx) {
		return subj, nil
	}

	ar, err := a.generateAccessRequest(ctx, resourceID)
	if err == nil {
		// Subject determines whether it is allowed to access resource.
		if !subj.CanAccess(action, ar) {
			err = internal.ErrAccessNotPermitted
		}
	}
	if err != nil {
		if !cfg.disableLogs {
			// TODO: disambiguate between logging errors due to subject lacking
			// sufficient permissions, and errors due to an internal problem
			// with the authorization process
			a.Error(err, "authorization failure",
				"resource", resourceID,
				"action", action.String(),
				"subject", subj,
			)
		}
		return nil, err
	}
	return subj, nil
}

func (a *Authorizer) generateAccessRequest(ctx context.Context, resourceID resource.ID) (AccessRequest, error) {
	if resourceID == resource.SiteID {
		return AccessRequest{ID: resourceID}, nil
	}
	var ar AccessRequest
	// Check if resource kind is registered for its ID to be resolved to workspace
	// ID.
	if resolver, ok := a.workspaceResolvers[resourceID.Kind()]; ok {
		workspaceID, err := resolver(ctx, resourceID)
		if err != nil {
			return AccessRequest{}, fmt.Errorf("resolving workspace ID: %w", err)
		}
		// Authorize workspace ID instead
		resourceID = workspaceID
	}
	// If the resource kind is a workspace, then fetch its policy.
	if resourceID.Kind() == resource.WorkspaceKind {
		policy, err := a.GetWorkspacePolicy(ctx, resourceID)
		if err != nil {
			return AccessRequest{}, fmt.Errorf("fetching workspace policy: %w", err)
		}
		ar.WorkspacePolicy = &policy
	}
	// Resolve the organization. Apart from an organization or the "site"
	// resource, every resource belongs to an organization, so there should be a
	// resolver for each resource kind to resolve the resource ID to the
	// organization it belongs to.
	if resourceID.Kind() == resource.OrganizationKind {
		ar.Organization = resourceID
	} else {
		resolver, ok := a.organizationResolvers[resourceID.Kind()]
		if !ok {
			return AccessRequest{}, errors.New("resource kind is missing organization resolver")
		}
		organization, err := resolver(ctx, resourceID)
		if err != nil {
			return AccessRequest{}, fmt.Errorf("resolving organization: %w", err)
		}
		ar.Organization = organization
	}
	return ar, nil
}

// CanAccess is a helper to boil down an access request to a true/false
// decision, with any error encountered interpreted as false.
func (a *Authorizer) CanAccess(ctx context.Context, action Action, id resource.ID) bool {
	_, err := a.Authorize(ctx, action, id, WithoutErrorLogging())
	return err == nil
}

// AccessRequest is a request for access to a resource.
type AccessRequest struct {
	// ID of resource to which access is being requested.
	ID resource.ID
	// Organization is the ID of the organization the resource belongs to, or is
	// the same as the ID above if access is being requested to an organization.
	// If access is being requested to the "site", then this is nil.
	Organization resource.ID
	// WorkspacePolicy specifies workspace-specific permissions for the resource
	// specified by ID above. This is nil if the resource is not a workspace or
	// does not belong to a workspace.
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
	TeamID resource.TfeID
	Role   Role
}

//func (r AccessRequest) LogValue() slog.Value {
//	if r == nil {
//		return slog.StringValue("site")
//	} else {
//		var attrs []slog.Attr
//		if r.Organization != nil {
//			attrs = append(attrs, slog.Any("organization", *r.Organization))
//		}
//		if r.ID != nil {
//			attrs = append(attrs, slog.String("resource_id", r.ID.String()))
//		}
//		return slog.GroupValue(attrs...)
//	}
//}
