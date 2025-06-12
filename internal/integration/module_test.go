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
		svc, org, ctx := setup(t)

		_, err := svc.Modules.CreateModule(ctx, module.CreateOptions{
			Name:         uuid.NewString(),
			Provider:     uuid.NewString(),
			Organization: org.Name,
		})
		require.NoError(t, err)
	})

	t.Run("create connected module", func(t *testing.T) {
		svc, org, ctx := setup(t, withGithubOption(github.WithRepo("leg100/terraform-aws-stuff")))

		vcsprov := svc.createVCSProvider(t, ctx, org, nil)

		mod, err := svc.Modules.PublishModule(ctx, module.PublishOptions{
			VCSProviderID: vcsprov.ID,
			Repo:          module.Repo("leg100/terraform-aws-stuff"),
		})
		require.NoError(t, err)

		// webhook should be registered with github
		hook := <-svc.WebhookEvents
		require.Equal(t, github.WebhookCreated, hook.Action)

		t.Run("delete module", func(t *testing.T) {
			_, err := svc.Modules.DeleteModule(ctx, mod.ID)
			require.NoError(t, err)
		})

		// webhook should now have been deleted from github
		hook = <-svc.WebhookEvents
		require.Equal(t, github.WebhookDeleted, hook.Action)
	})

	t.Run("get", func(t *testing.T) {
		svc, _, ctx := setup(t)
		want := svc.createModule(t, ctx, nil)

		got, err := svc.Modules.GetModule(ctx, module.GetModuleOptions{
			Organization: want.Organization,
			Provider:     want.Provider,
			Name:         want.Name,
		})
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("get by id", func(t *testing.T) {
		svc, _, ctx := setup(t)
		want := svc.createModule(t, ctx, nil)

		got, err := svc.Modules.GetModuleByID(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("list", func(t *testing.T) {
		svc, _, ctx := setup(t)
		org := svc.createOrganization(t, ctx)
		module1 := svc.createModule(t, ctx, org)
		module2 := svc.createModule(t, ctx, org)
		module3 := svc.createModule(t, ctx, org)

		got, err := svc.Modules.ListModules(ctx, module.ListOptions{
			Organization: org.Name,
		})
		require.NoError(t, err)

		assert.Contains(t, got, module1)
		assert.Contains(t, got, module2)
		assert.Contains(t, got, module3)
	})

	t.Run("delete", func(t *testing.T) {
		svc, _, ctx := setup(t)
		want := svc.createModule(t, ctx, nil)

		got, err := svc.Modules.DeleteModule(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})
}
