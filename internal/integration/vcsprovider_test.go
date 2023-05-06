package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/vcsprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVCSProvider(t *testing.T) {
	t.Parallel()

	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &auth.SiteAdmin)

	t.Run("create", func(t *testing.T) {
		svc := setup(t, nil)
		org := svc.createOrganization(t, ctx)

		_, err := svc.CreateVCSProvider(ctx, vcsprovider.CreateOptions{
			Organization: org.Name,
			Token:        uuid.NewString(),
			Name:         uuid.NewString(),
			Cloud:        "github",
		})
		require.NoError(t, err)
	})

	t.Run("get", func(t *testing.T) {
		svc := setup(t, nil)
		want := svc.createVCSProvider(t, ctx, nil)

		got, err := svc.GetVCSProvider(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("list", func(t *testing.T) {
		svc := setup(t, nil)
		org := svc.createOrganization(t, ctx)
		provider1 := svc.createVCSProvider(t, ctx, org)
		provider2 := svc.createVCSProvider(t, ctx, org)
		provider3 := svc.createVCSProvider(t, ctx, org)

		got, err := svc.ListVCSProviders(ctx, org.Name)
		require.NoError(t, err)

		assert.Contains(t, got, provider1)
		assert.Contains(t, got, provider2)
		assert.Contains(t, got, provider3)
	})

	t.Run("delete", func(t *testing.T) {
		svc := setup(t, nil)
		want := svc.createVCSProvider(t, ctx, nil)

		got, err := svc.DeleteVCSProvider(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})
}
