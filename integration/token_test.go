package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToken(t *testing.T) {
	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &otf.Superuser{})

	t.Run("create", func(t *testing.T) {
		svc := setup(t, nil)
		// create user and then add them to context so that it is their token
		// that is created.
		ctx := otf.AddSubjectToContext(ctx, svc.createUser(t, ctx))
		_, err := svc.CreateToken(ctx, &auth.TokenCreateOptions{
			Description: "lorem ipsum...",
		})
		require.NoError(t, err)
	})

	t.Run("list", func(t *testing.T) {
		svc := setup(t, nil)
		user := svc.createUser(t, ctx)
		// create user and then add them to context so that it is their token
		// that is created.
		ctx := otf.AddSubjectToContext(ctx, user)

		_ = svc.createToken(t, ctx, user)
		_ = svc.createToken(t, ctx, user)
		_ = svc.createToken(t, ctx, user)

		got, err := svc.ListTokens(ctx)
		require.NoError(t, err)

		assert.Equal(t, 3, len(got))
	})

	t.Run("delete", func(t *testing.T) {
		svc := setup(t, nil)
		user := svc.createUser(t, ctx)
		// create user and then add them to context so that it is their token
		// that is created.
		ctx := otf.AddSubjectToContext(ctx, user)
		token := svc.createToken(t, ctx, user)

		err := svc.DeleteToken(ctx, token.ID)
		require.NoError(t, err)
	})
}
