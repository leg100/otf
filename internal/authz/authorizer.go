package authz

import (
	"context"
	"errors"
	"fmt"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
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

type ParentResolver func(ctx context.Context, id resource.ID) (resource.ID, error)

// RegisterParentResolver registers with the authorizer a means of resolving the
// parent of a resource.
func (a *Authorizer) RegisterParentResolver(kind resource.Kind, resolver ParentResolver) {
	a.parentResolvers[kind] = resolver
}

// WorkspacePolicyGetter retrieves a workspace's policy.
type WorkspacePolicyGetter func(ctx context.Context, workspaceID resource.ID) (WorkspacePolicy, error)

// WorkspacePolicy checks whether a subject is permitted to carry out an action
// on a workspace.
type WorkspacePolicy interface {
	Check(subject resource.ID, action Action) bool
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
		// TODO: even in the event of error we return the subject, because some
		// callers rely upon this. This should be changed because it goes
		// against idiomatic go.
		return subj, err
	}
	return subj, nil
}

func (a *Authorizer) generateRequest(ctx context.Context, resourceID resource.ID) (Request, error) {
	req := Request{ID: resourceID}
	if resourceID == resource.SiteID {
		return req, nil
	}
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
		req.lineage = append(req.lineage, parentID)
		// now try looking up parent of parent
		resourceID = parentID
	}
	// If the requested resource is a workspace or belongs to a workspace then
	// fetch its workspace policy.
	if req.Workspace() != nil {
		checker, err := a.WorkspacePolicyGetter(ctx, req.Workspace())
		if err != nil {
			return Request{}, fmt.Errorf("fetching workspace policy: %w", err)
		}
		req.WorkspacePolicy = checker
	}
	return req, nil
}

// CanAccess is a helper to boil down an access request to a true/false
// decision, with any error encountered interpreted as false.
func (a *Authorizer) CanAccess(ctx context.Context, action Action, id resource.ID) bool {
	_, err := a.Authorize(ctx, action, id, WithoutErrorLogging())
	return err == nil
}
