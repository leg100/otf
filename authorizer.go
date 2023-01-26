package otf

import (
	"context"

	"github.com/leg100/otf/rbac"
)

// Authorizer is capable of granting or denying access to resources based on the
// subject contained within the context.
type Authorizer interface {
	CanAccessSite(ctx context.Context, action rbac.Action) (Subject, error)
	CanAccessOrganization(ctx context.Context, action rbac.Action, name string) (Subject, error)
	CanAccessWorkspaceByName(ctx context.Context, action rbac.Action, organization, workspace string) (Subject, error)
	CanAccessWorkspaceByID(ctx context.Context, action rbac.Action, workspaceID string) (Subject, error)
	CanAccessRun(ctx context.Context, action rbac.Action, runID string) (Subject, error)
	CanAccessStateVersion(ctx context.Context, action rbac.Action, svID string) (Subject, error)
	CanAccessConfigurationVersion(ctx context.Context, action rbac.Action, cvID string) (Subject, error)
}
