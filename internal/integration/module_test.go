package integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal/github/testserver"
	"github.com/leg100/otf/internal/module"
	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModule(t *testing.T) {
	integrationTest(t)

	t.Run("create", func(t *testing.T) {
		daemon, org, ctx := setup(t)

		_, err := daemon.Modules.CreateModule(ctx, module.CreateOptions{
			Name:         uuid.NewString(),
			Provider:     uuid.NewString(),
			Organization: org.Name,
		})
		require.NoError(t, err)
	})

	t.Run("create connected module", func(t *testing.T) {
		daemon, org, ctx := setup(t, withGithubOption(testserver.WithRepo(vcs.NewMustRepo("leg100", "terraform-aws-stuff"))))

		vcsprov := daemon.createVCSProvider(t, ctx, org, nil)

		mod, err := daemon.Modules.PublishModule(ctx, module.PublishOptions{
			VCSProviderID: vcsprov.ID,
			Repo:          module.Repo(vcs.NewMustRepo("leg100", "terraform-aws-stuff")),
		})
		require.NoError(t, err)

		// webhook should be registered with github
		hook := <-daemon.WebhookEvents
		require.Equal(t, testserver.WebhookCreated, hook.Action)

		t.Run("delete module", func(t *testing.T) {
			_, err := daemon.Modules.DeleteModule(ctx, mod.ID)
			require.NoError(t, err)
		})

		// webhook should now have been deleted from github
		hook = <-daemon.WebhookEvents
		require.Equal(t, testserver.WebhookDeleted, hook.Action)
	})

	t.Run("get", func(t *testing.T) {
		daemon, _, ctx := setup(t)
		want := daemon.createModule(t, ctx, nil)

		got, err := daemon.Modules.GetModule(ctx, module.GetModuleOptions{
			Organization: want.Organization,
			Provider:     want.Provider,
			Name:         want.Name,
		})
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("get by id", func(t *testing.T) {
		daemon, _, ctx := setup(t)
		want := daemon.createModule(t, ctx, nil)

		got, err := daemon.Modules.GetModuleByID(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("list", func(t *testing.T) {
		daemon, _, ctx := setup(t)
		org := daemon.createOrganization(t, ctx)
		module1 := daemon.createModule(t, ctx, org)
		module2 := daemon.createModule(t, ctx, org)
		module3 := daemon.createModule(t, ctx, org)

		got, err := daemon.Modules.ListModules(ctx, module.ListOptions{
			Organization: org.Name,
		})
		require.NoError(t, err)

		assert.Contains(t, got, module1)
		assert.Contains(t, got, module2)
		assert.Contains(t, got, module3)
	})

	t.Run("delete", func(t *testing.T) {
		daemon, _, ctx := setup(t)
		want := daemon.createModule(t, ctx, nil)

		got, err := daemon.Modules.DeleteModule(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})
}
