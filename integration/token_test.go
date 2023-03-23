package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/stretchr/testify/require"
)

func TestToken(t *testing.T) {
	ctx := context.Background()

	t.Run("create", func(t *testing.T) {
		svc := setup(t, "")
		ctx := otf.AddSubjectToContext(ctx, svc.createUser(t, ctx))
		_, err := svc.CreateToken(ctx, &auth.TokenCreateOptions{
			Description: "lorem ipsum...",
		})
		require.NoError(t, err)
	})

	t.Run("delete", func(t *testing.T) {
		svc := setup(t, "")
		user := svc.createUser(t, ctx)
		token := svc.createToken(t, ctx, user, nil)

		ctx := otf.AddSubjectToContext(ctx, user)
		err := svc.DeleteToken(ctx, token.ID)
		require.NoError(t, err)
	})
}
