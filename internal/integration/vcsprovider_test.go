package integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVCSProvider(t *testing.T) {
	integrationTest(t)

	t.Run("create", func(t *testing.T) {
		svc, org, ctx := setup(t)

		_, err := svc.VCSProviders.Create(ctx, vcs.CreateOptions{
			Organization: org.Name,
			Token:        internal.String(uuid.NewString()),
			KindID:       github.TokenKindID,
		})
		require.NoError(t, err)
	})

	t.Run("get", func(t *testing.T) {
		svc, _, ctx := setup(t)
		want := svc.createVCSProvider(t, ctx, nil)

		got, err := svc.VCSProviders.Get(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("list", func(t *testing.T) {
		svc, org, ctx := setup(t)
		provider1 := svc.createVCSProvider(t, ctx, org)
		provider2 := svc.createVCSProvider(t, ctx, org)
		provider3 := svc.createVCSProvider(t, ctx, org)

		got, err := svc.VCSProviders.List(ctx, org.Name)
		require.NoError(t, err)

		assert.Contains(t, got, provider1)
		assert.Contains(t, got, provider2)
		assert.Contains(t, got, provider3)
	})

	t.Run("delete", func(t *testing.T) {
		svc, _, ctx := setup(t)
		want := svc.createVCSProvider(t, ctx, nil)

		got, err := svc.VCSProviders.Delete(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})
}
