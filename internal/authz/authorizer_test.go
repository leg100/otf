package authz

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorizer(t *testing.T) {
	authorizer := NewAuthorizer(logr.Discard())
	user := &Superuser{}
	ctx := AddSubjectToContext(context.Background(), user)

	got, err := authorizer.Authorize(ctx, ListUsersAction, resource.SiteID)
	require.NoError(t, err)

	assert.Equal(t, user, got)
}
