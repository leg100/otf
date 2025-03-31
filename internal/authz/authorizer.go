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

	parentResolvers map[resource.Kind]ParentResolver
}

// Interface provides an interface for services to use to permit swapping out
// the authorizer for tests.
type Interface interface {
	Authorize(ctx context.Context, action Action, id resource.ID, opts ...CanAccessOption) (Subject, error)
	CanAccess(ctx context.Context, action Action, id resource.ID) bool
}

func NewAuthorizer(logger logr.Logger) *Authorizer {
	return &Authorizer{
		Logger:          logger,
		parentResolvers: make(map[resource.Kind]ParentResolver),
	}
}

type WorkspacePolicyGetter interface {
	GetWorkspacePolicy(ctx context.Context, workspaceID resource.TfeID) (WorkspacePolicy, error)
}

type ParentResolver func(ctx context.Context, id resource.ID) (resource.ID, error)

func (a *Authorizer) RegisterParentResolver(kind resource.Kind, resolver ParentResolver) {
	a.parentResolvers[kind] = resolver
}

// RegisterOrganizationResolver registers with the authorizer the ability to
// resolve access requests for a specific resource kind to the name of the
// organization the resource belongs to.
//
// This is necessary because authorization is determined not only on resource ID
// but on the name of the organization the resource belongs to.

// RegisterWorkspaceResolver registers with the authorizer the ability to
// resolve access requests for a specific resource kind to the workspace ID the
// resource belongs to.
//
// This is necessary because authorization is often determined based on
// workspace ID, and not the ID of a run, state version, etc.

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

	ar, err := a.generateRequest(ctx, resourceID)
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

func (a *Authorizer) generateRequest(ctx context.Context, resourceID resource.ID) (Request, error) {
	req := Request{ID: resourceID}
	if resourceID == resource.SiteID {
		return req, nil
	}
	var ar Request
	// Retrieve resource lineage
	for {
		parent, ok := a.parentResolvers[resourceID.Kind()]
		if !ok {
			break
		}
		parentID, err := parent(ctx, resourceID)
		if err != nil {
			return Request{}, fmt.Errorf("resolving ID: %w", err)
		}
		ar.lineage = append(ar.lineage, parentID)
		// now try looking up parent of parent
		resourceID = parentID
	}
	// If the requested resource is a workspace or belongs to a workspace then
	// fetch its workspace policy.
	if ar.Workspace() != nil {
		policy, err := a.GetWorkspacePolicy(ctx, ar.Workspace().(resource.TfeID))
		if err != nil {
			return Request{}, fmt.Errorf("fetching workspace policy: %w", err)
		}
		ar.WorkspacePolicy = &policy
	}
	return ar, nil
}

// CanAccess is a helper to boil down an access request to a true/false
// decision, with any error encountered interpreted as false.
func (a *Authorizer) CanAccess(ctx context.Context, action Action, id resource.ID) bool {
	_, err := a.Authorize(ctx, action, id, WithoutErrorLogging())
	return err == nil
}

type Request struct {
	// ID of resource to which access is being requested.
	resource.ID
	// WorkspacePolicy specifies workspace-specific permissions for the resource
	// specified by ID above. This is nil if the resource is not a workspace or
	// does not belong to a workspace.
	WorkspacePolicy *WorkspacePolicy
	// lineage are the parents of the resource.
	lineage []resource.ID
}

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
