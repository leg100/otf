package authz

import (
	"context"

	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
)

// Authorizer is capable of granting or denying access to resources based on the
// subject contained within the context.
type Authorizer interface {
	CanAccess(ctx context.Context, action rbac.Action, id resource.ID) (Subject, error)
}
