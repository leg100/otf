package otf

import (
	"context"

	"github.com/leg100/otf/rbac"
)

// Authorizer is capable of granting or denying access to resources based on the
// subject contained within the context.
type Authorizer interface {
	CanAccessRun(ctx context.Context, action rbac.Action, runID string) (Subject, error)
	CanAccessStateVersion(ctx context.Context, action rbac.Action, svID string) (Subject, error)
	CanAccessConfigurationVersion(ctx context.Context, action rbac.Action, cvID string) (Subject, error)
}

type OrganizationAuthorizer interface {
	CanAccessOrganization(ctx context.Context, action rbac.Action, name string) (Subject, error)
}
