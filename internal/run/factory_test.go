package run

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/repo"
	"github.com/leg100/otf/internal/vcsprovider"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFactory(t *testing.T) {
	ctx := context.Background()

	t.Run("defaults", func(t *testing.T) {
		f := newTestFactory(
			&workspace.Workspace{},
			&configversion.ConfigurationVersion{},
		)

		got, err := f.NewRun(ctx, "", RunCreateOptions{})
		require.NoError(t, err)

		assert.Equal(t, internal.RunPending, got.Status)
		assert.NotZero(t, got.CreatedAt)
		assert.False(t, got.PlanOnly)
		assert.True(t, got.Refresh)
		assert.False(t, got.AutoApply)
	})

	t.Run("speculative run", func(t *testing.T) {
		f := newTestFactory(
			&workspace.Workspace{},
			&configversion.ConfigurationVersion{Speculative: true},
		)

		got, err := f.NewRun(ctx, "", RunCreateOptions{})
		require.NoError(t, err)

		assert.True(t, got.PlanOnly)
	})

	t.Run("plan-only run", func(t *testing.T) {
		f := newTestFactory(
			&workspace.Workspace{},
			&configversion.ConfigurationVersion{},
		)

		got, err := f.NewRun(ctx, "", RunCreateOptions{PlanOnly: internal.Bool(true)})
		require.NoError(t, err)

		assert.True(t, got.PlanOnly)
	})

	t.Run("workspace auto-apply", func(t *testing.T) {
		f := newTestFactory(
			&workspace.Workspace{AutoApply: true},
			&configversion.ConfigurationVersion{},
		)

		got, err := f.NewRun(ctx, "", RunCreateOptions{})
		require.NoError(t, err)

		assert.True(t, got.AutoApply)
	})

	t.Run("run auto-apply", func(t *testing.T) {
		f := newTestFactory(
			&workspace.Workspace{},
			&configversion.ConfigurationVersion{},
		)

		got, err := f.NewRun(ctx, "", RunCreateOptions{
			AutoApply: internal.Bool(true),
		})
		require.NoError(t, err)

		assert.True(t, got.AutoApply)
	})

	t.Run("magic string - pull from vcs", func(t *testing.T) {
		f := newTestFactory(
			&workspace.Workspace{
				Connection: &workspace.Connection{
					Connection: &repo.Connection{},
				},
			},
			&configversion.ConfigurationVersion{},
		)

		got, err := f.NewRun(ctx, "", RunCreateOptions{
			ConfigurationVersionID: internal.String(PullVCSMagicString),
		})
		require.NoError(t, err)

		// fake config version service sets the config version ID to "created"
		// if it was newly created
		assert.Equal(t, "created", got.ConfigurationVersionID)
	})

	t.Run("pull from vcs", func(t *testing.T) {
		f := newTestFactory(
			&workspace.Workspace{
				Connection: &workspace.Connection{
					Connection: &repo.Connection{},
				},
			},
			&configversion.ConfigurationVersion{},
		)

		got, err := f.NewRun(ctx, "", RunCreateOptions{})
		require.NoError(t, err)

		// fake config version service sets the config version ID to "created"
		// if it was newly created
		assert.Equal(t, "created", got.ConfigurationVersionID)
	})

	t.Run("pull from vcs workspace not connected error", func(t *testing.T) {
		f := newTestFactory(
			&workspace.Workspace{}, // workspace with no connection
			&configversion.ConfigurationVersion{},
		)

		_, err := f.NewRun(ctx, "", RunCreateOptions{
			ConfigurationVersionID: internal.String(PullVCSMagicString),
		})
		require.Equal(t, err, workspace.ErrNoVCSConnection)
	})
}

type (
	fakeFactoryWorkspaceService struct {
		ws *workspace.Workspace
		workspace.Service
	}
	fakeFactoryConfigurationVersionService struct {
		cv *configversion.ConfigurationVersion
		configversion.Service
	}
	fakeFactoryVCSProviderService struct {
		vcsprovider.Service
	}
	fakeFactoryCloudClient struct {
		cloud.Client
	}
)

func newTestFactory(ws *workspace.Workspace, cv *configversion.ConfigurationVersion) *factory {
	return &factory{
		WorkspaceService:            &fakeFactoryWorkspaceService{ws: ws},
		ConfigurationVersionService: &fakeFactoryConfigurationVersionService{cv: cv},
		VCSProviderService:          &fakeFactoryVCSProviderService{},
	}
}

func (f *fakeFactoryWorkspaceService) GetWorkspace(context.Context, string) (*workspace.Workspace, error) {
	return f.ws, nil
}

func (f *fakeFactoryConfigurationVersionService) GetConfigurationVersion(context.Context, string) (*configversion.ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *fakeFactoryConfigurationVersionService) GetLatestConfigurationVersion(context.Context, string) (*configversion.ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *fakeFactoryConfigurationVersionService) CreateConfigurationVersion(context.Context, string, configversion.ConfigurationVersionCreateOptions) (*configversion.ConfigurationVersion, error) {
	return &configversion.ConfigurationVersion{ID: "created"}, nil
}

func (f *fakeFactoryConfigurationVersionService) UploadConfig(context.Context, string, []byte) error {
	return nil
}

func (f *fakeFactoryVCSProviderService) GetVCSClient(context.Context, string) (cloud.Client, error) {
	return &fakeFactoryCloudClient{}, nil
}

func (f *fakeFactoryCloudClient) GetRepoTarball(context.Context, cloud.GetRepoTarballOptions) ([]byte, error) {
	return nil, nil
}
