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
	ctx := context.Background()

	t.Run("create", func(t *testing.T) {
		svc := setup(t, nil)
		ctx := otf.AddSubjectToContext(ctx, svc.createUser(t, ctx))
		_, err := svc.CreateToken(ctx, &auth.TokenCreateOptions{
			Description: "lorem ipsum...",
		})
		require.NoError(t, err)
	})

	t.Run("list", func(t *testing.T) {
		svc := setup(t, nil)
		user := svc.createUser(t, ctx)

		_ = svc.createToken(t, ctx, user)
		_ = svc.createToken(t, ctx, user)
		_ = svc.createToken(t, ctx, user)

		ctx := otf.AddSubjectToContext(ctx, user)
		got, err := svc.ListTokens(ctx)
		require.NoError(t, err)

		assert.Equal(t, 3, len(got))
	})

	t.Run("delete", func(t *testing.T) {
		svc := setup(t, nil)
		user := svc.createUser(t, ctx)
		token := svc.createToken(t, ctx, user)

		ctx := otf.AddSubjectToContext(ctx, user)
		err := svc.DeleteToken(ctx, token.ID)
		require.NoError(t, err)
	})
}
