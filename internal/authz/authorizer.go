package authz

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/organization"
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
	Authorize(ctx context.Context, action Action, req *AccessRequest, opts ...CanAccessOption) (Subject, error)
	CanAccess(ctx context.Context, action Action, req *AccessRequest) bool
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
func (a *Authorizer) Authorize(ctx context.Context, action Action, req *AccessRequest, opts ...CanAccessOption) (Subject, error) {
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
	// Wrapped in function in order to log error messages uniformly.
	err = func() error {
		if req != nil && req.ID != nil {
			// Check if resource kind is registered for its ID to be resolved to workspace
			// ID.
			if resolver, ok := a.workspaceResolvers[req.ID.Kind()]; ok {
				workspaceID, err := resolver(ctx, *req.ID)
				if err != nil {
					return fmt.Errorf("resolving workspace ID: %w", err)
				}
				// Authorize workspace ID instead
				req.ID = &workspaceID
			}
			// If the resource kind is a workspace, then fetch its policy.
			if req.ID.Kind() == resource.WorkspaceKind {
				policy, err := a.GetWorkspacePolicy(ctx, *req.ID)
				if err != nil {
					return fmt.Errorf("fetching workspace policy: %w", err)
				}
				req.WorkspacePolicy = &policy
			}
			// Resolve the organization if not already provided. Every resource
			// belongs to an organization, so there should be a resolver for each
			// resource kind to resolve the resource ID to the organization it
			// belongs to.
			if req.Organization == "" {
				resolver, ok := a.organizationResolvers[req.ID.Kind()]
				if !ok {
					return errors.New("resource kind is missing organization resolver")
				}
				organization, err := resolver(ctx, *req.ID)
				if err != nil {
					return fmt.Errorf("resolving organization: %w", err)
				}
				req.Organization = organization
			}
		}
		// Subject determines whether it is allowed to access resource.
		if !subj.CanAccess(action, req) {
			return internal.ErrAccessNotPermitted
		}
		return nil
	}()
	if err != nil && !cfg.disableLogs {
		a.Error(err, "authorization failure",
			"resource", req,
			"action", action.String(),
			"subject", subj,
		)
	}
	return subj, err
}

// CanAccess is a helper to boil down an access request to a true/false
// decision, with any error encountered interpreted as false.
func (a *Authorizer) CanAccess(ctx context.Context, action Action, req *AccessRequest) bool {
	_, err := a.Authorize(ctx, action, req, WithoutErrorLogging())
	return err == nil
}

// AccessRequest is a request for access to either an organization or an
// individual resource.
type AccessRequest struct {
	// Organization name to which access is being requested.
	Organization organization.Name
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
	Role   Role
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
