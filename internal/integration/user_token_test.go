package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserToken(t *testing.T) {
	integrationTest(t)

	// perform all actions as superuser
	ctx := authz.AddSubjectToContext(context.Background(), &user.SiteAdmin)

	t.Run("create", func(t *testing.T) {
		svc, _, _ := setup(t)
		// create user and then add them to context so that it is their token
		// that is created.
		ctx := authz.AddSubjectToContext(ctx, svc.createUser(t))
		_, _, err := svc.Users.CreateToken(ctx, user.CreateUserTokenOptions{
			Description: "lorem ipsum...",
		})
		require.NoError(t, err)
	})

	t.Run("list", func(t *testing.T) {
		svc, _, ctx := setup(t)
		svc.createToken(t, ctx, nil)
		svc.createToken(t, ctx, nil)
		svc.createToken(t, ctx, nil)

		got, err := svc.Users.ListTokens(ctx)
		require.NoError(t, err)

		assert.Equal(t, 3, len(got))
	})

	t.Run("delete", func(t *testing.T) {
		svc, _, ctx := setup(t)
		token, _ := svc.createToken(t, ctx, nil)

		err := svc.Users.DeleteToken(ctx, token.ID)
		require.NoError(t, err)
	})
}
