package authz

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/rbac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorizer(t *testing.T) {
	authorizer := NewAuthorizer(logr.Discard())
	user := &Superuser{}
	ctx := AddSubjectToContext(context.Background(), user)

	got, err := authorizer.CanAccess(ctx, rbac.ListUsersAction, nil)
	require.NoError(t, err)

	assert.Equal(t, user, got)
}
