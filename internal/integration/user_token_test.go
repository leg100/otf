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
		daemon, _, _ := setup(t)
		// create user and then add them to context so that it is their token
		// that is created.
		ctx := authz.AddSubjectToContext(ctx, daemon.createUser(t))
		_, _, err := daemon.Users.CreateToken(ctx, user.CreateUserTokenOptions{
			Description: "lorem ipsum...",
		})
		require.NoError(t, err)
	})

	t.Run("list", func(t *testing.T) {
		daemon, _, ctx := setup(t)
		daemon.createToken(t, ctx, nil)
		daemon.createToken(t, ctx, nil)
		daemon.createToken(t, ctx, nil)

		got, err := daemon.Users.ListTokens(ctx)
		require.NoError(t, err)

		assert.Equal(t, 3, len(got))
	})

	t.Run("delete", func(t *testing.T) {
		daemon, _, ctx := setup(t)
		token, _ := daemon.createToken(t, ctx, nil)

		err := daemon.Users.DeleteToken(ctx, token.ID)
		require.NoError(t, err)
	})
}
