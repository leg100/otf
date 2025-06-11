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
		want := svc.createVCSProvider(t, ctx, nil, nil)

		got, err := svc.VCSProviders.Get(ctx, want.ID)
		require.NoError(t, err)

		vcsProviderIsEqual(t, want, got)
	})

	t.Run("list", func(t *testing.T) {
		svc, org, ctx := setup(t)
		provider1 := svc.createVCSProvider(t, ctx, org, &vcs.CreateOptions{Name: "alpha"})
		provider2 := svc.createVCSProvider(t, ctx, org, &vcs.CreateOptions{Name: "beta"})
		provider3 := svc.createVCSProvider(t, ctx, org, &vcs.CreateOptions{Name: "gamma"})

		got, err := svc.VCSProviders.List(ctx, org.Name)
		require.NoError(t, err)

		vcsProviderIsEqual(t, got[0], provider1)
		vcsProviderIsEqual(t, got[1], provider2)
		vcsProviderIsEqual(t, got[2], provider3)
	})

	t.Run("delete", func(t *testing.T) {
		svc, _, ctx := setup(t)
		want := svc.createVCSProvider(t, ctx, nil, nil)

		got, err := svc.VCSProviders.Delete(ctx, want.ID)
		require.NoError(t, err)

		vcsProviderIsEqual(t, want, got)
	})
}

func vcsProviderIsEqual(t *testing.T, want, got *vcs.Provider) {
	assert.Equal(t, want.ID, got.ID)
	assert.Equal(t, want.Name, got.Name)
	assert.Equal(t, want.CreatedAt, got.CreatedAt)
	assert.Equal(t, want.Organization, got.Organization)
	assert.Equal(t, want.Kind.ID, got.Kind.ID)
	assert.Equal(t, want.Kind.Hostname, got.Kind.Hostname)
	assert.Equal(t, want.Kind.TokenKind, got.Kind.TokenKind)
	assert.Equal(t, want.Kind.SkipRepohook, got.Kind.SkipRepohook)
}
