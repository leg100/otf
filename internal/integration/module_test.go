package integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/module"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModule(t *testing.T) {
	integrationTest(t)

	t.Run("create", func(t *testing.T) {
		svc, org, ctx := setup(t, nil)

		_, err := svc.CreateModule(ctx, module.CreateOptions{
			Name:         uuid.NewString(),
			Provider:     uuid.NewString(),
			Organization: org.Name,
		})
		require.NoError(t, err)
	})

	t.Run("create connected module", func(t *testing.T) {
		svc, org, ctx := setup(t, nil, github.WithRepo("leg100/terraform-aws-stuff"))

		vcsprov := svc.createVCSProvider(t, ctx, org)
		mod := svc.createModule(t, ctx, org)

		mod, err := svc.PublishModule(ctx, module.PublishOptions{
			VCSProviderID: vcsprov.ID,
			Repo:          module.Repo("leg100/terraform-aws-stuff"),
		})
		require.NoError(t, err)

		// webhook should be registered with github
		require.True(t, svc.HasWebhook())

		t.Run("delete module", func(t *testing.T) {
			_, err := svc.DeleteModule(ctx, mod.ID)
			require.NoError(t, err)
		})

		// webhook should now have been deleted from github
		require.False(t, svc.HasWebhook())
	})

	t.Run("get", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		want := svc.createModule(t, ctx, nil)

		got, err := svc.GetModule(ctx, module.GetModuleOptions{
			Organization: want.Organization,
			Provider:     want.Provider,
			Name:         want.Name,
		})
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("get by id", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		want := svc.createModule(t, ctx, nil)

		got, err := svc.GetModuleByID(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("list", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		org := svc.createOrganization(t, ctx)
		module1 := svc.createModule(t, ctx, org)
		module2 := svc.createModule(t, ctx, org)
		module3 := svc.createModule(t, ctx, org)

		got, err := svc.ListModules(ctx, module.ListModulesOptions{
			Organization: org.Name,
		})
		require.NoError(t, err)

		assert.Contains(t, got, module1)
		assert.Contains(t, got, module2)
		assert.Contains(t, got, module3)
	})

	t.Run("delete", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		want := svc.createModule(t, ctx, nil)

		got, err := svc.DeleteModule(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})
}
