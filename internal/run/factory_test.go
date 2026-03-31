package run

import (
	"context"
	"testing"
	"time"

	"github.com/a-h/templ"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/configversion/source"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFactory(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		f := newTestFactory(
			&organization.Organization{},
			workspace.NewTestWorkspace(t, nil),
			&configversion.ConfigurationVersion{},
		)

		got, err := f.NewRun(t.Context(), resource.TfeID{}, CreateOptions{})
		require.NoError(t, err)

		assert.Equal(t, runstatus.Pending, got.Status)
		assert.NotZero(t, got.CreatedAt)
		assert.False(t, got.PlanOnly)
		assert.True(t, got.Refresh)
		assert.False(t, got.AutoApply)
		assert.Equal(t, "1.9.0", got.EngineVersion)
	})

	t.Run("speculative run", func(t *testing.T) {
		f := newTestFactory(
			&organization.Organization{},
			workspace.NewTestWorkspace(t, nil),
			&configversion.ConfigurationVersion{Speculative: true},
		)

		got, err := f.NewRun(t.Context(), resource.TfeID{}, CreateOptions{})
		require.NoError(t, err)

		assert.True(t, got.PlanOnly)
	})

	t.Run("plan-only run", func(t *testing.T) {
		f := newTestFactory(
			&organization.Organization{},
			workspace.NewTestWorkspace(t, nil),
			&configversion.ConfigurationVersion{},
		)

		got, err := f.NewRun(t.Context(), resource.TfeID{}, CreateOptions{PlanOnly: new(true)})
		require.NoError(t, err)

		assert.True(t, got.PlanOnly)
	})

	t.Run("workspace auto-apply", func(t *testing.T) {
		f := newTestFactory(
			&organization.Organization{},
			workspace.NewTestWorkspace(t, &workspace.CreateOptions{AutoApply: new(true)}),
			&configversion.ConfigurationVersion{},
		)

		got, err := f.NewRun(t.Context(), resource.TfeID{}, CreateOptions{})
		require.NoError(t, err)

		assert.True(t, got.AutoApply)
	})

	t.Run("run auto-apply", func(t *testing.T) {
		f := newTestFactory(
			&organization.Organization{},
			workspace.NewTestWorkspace(t, nil),
			&configversion.ConfigurationVersion{},
		)

		got, err := f.NewRun(t.Context(), resource.TfeID{}, CreateOptions{
			AutoApply: new(true),
		})
		require.NoError(t, err)

		assert.True(t, got.AutoApply)
	})

	t.Run("enable cost estimation", func(t *testing.T) {
		f := newTestFactory(
			&organization.Organization{CostEstimationEnabled: true},
			workspace.NewTestWorkspace(t, nil),
			&configversion.ConfigurationVersion{},
		)

		got, err := f.NewRun(t.Context(), resource.TfeID{}, CreateOptions{})
		require.NoError(t, err)

		assert.True(t, got.CostEstimationEnabled)
	})

	t.Run("pull from vcs", func(t *testing.T) {
		vcsProviderID := resource.NewTfeID(resource.VCSProviderKind)
		ws := workspace.NewTestWorkspace(t, &workspace.CreateOptions{
			ConnectOptions: &workspace.ConnectOptions{
				RepoPath:      &vcs.Repo{},
				VCSProviderID: &vcsProviderID,
			},
		})

		f := newTestFactory(
			&organization.Organization{},
			ws,
			&configversion.ConfigurationVersion{},
		)

		got, err := f.NewRun(t.Context(), resource.TfeID{}, CreateOptions{})
		require.NoError(t, err)

		// fake config version service sets the config version ID to the
		// workspace ID if it was newly created
		assert.Equal(t, ws.ID, got.ConfigurationVersionID)
	})

	t.Run("get latest engine version", func(t *testing.T) {
		f := newTestFactory(
			&organization.Organization{},
			workspace.NewTestWorkspace(t, &workspace.CreateOptions{
				EngineVersion: &workspace.Version{Latest: true},
			}),
			&configversion.ConfigurationVersion{},
		)
		f.client.(*fakeFactoryClient).latestVersion = "2.0.0"

		got, err := f.NewRun(t.Context(), resource.TfeID{}, CreateOptions{})
		require.NoError(t, err)

		assert.Equal(t, "2.0.0", got.EngineVersion)
	})
}

type (
	fakeFactoryClient struct {
		serviceClient

		org           *organization.Organization
		ws            *workspace.Workspace
		cv            *configversion.ConfigurationVersion
		latestVersion string
	}
)

func newTestFactory(org *organization.Organization, ws *workspace.Workspace, cv *configversion.ConfigurationVersion) *factory {
	return &factory{
		client: &fakeFactoryClient{
			org: org,
			ws:  ws,
			cv:  cv,
		},
	}
}

func (f *fakeFactoryClient) GetOrganization(context.Context, organization.Name) (*organization.Organization, error) {
	return f.org, nil
}

func (f *fakeFactoryClient) GetWorkspace(context.Context, resource.ID) (*workspace.Workspace, error) {
	return f.ws, nil
}

func (f *fakeFactoryClient) GetConfigVersion(context.Context, resource.TfeID) (*configversion.ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *fakeFactoryClient) GetLatestConfigVersion(context.Context, resource.TfeID) (*configversion.ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *fakeFactoryClient) CreateConfigVersion(ctx context.Context, workspaceID resource.TfeID, opts configversion.CreateOptions) (*configversion.ConfigurationVersion, error) {
	return &configversion.ConfigurationVersion{ID: workspaceID}, nil
}

func (f *fakeFactoryClient) UploadConfig(context.Context, resource.TfeID, []byte) error {
	return nil
}

func (f *fakeFactoryClient) GetSourceIcon(source source.Source) templ.Component {
	return templ.Raw("")
}

func (f *fakeFactoryClient) GetLatest(context.Context, *engine.Engine) (string, time.Time, error) {
	return f.latestVersion, time.Time{}, nil
}

func (f *fakeFactoryClient) GetVCSProvider(context.Context, resource.TfeID) (*vcs.Provider, error) {
	return &vcs.Provider{
		Client: &fakeFactoryCloudClient{},
	}, nil
}

type fakeFactoryCloudClient struct {
	vcs.Client
}

func (f *fakeFactoryCloudClient) GetRepoTarball(context.Context, vcs.GetRepoTarballOptions) ([]byte, string, error) {
	return nil, "", nil
}

func (f *fakeFactoryCloudClient) GetDefaultBranch(context.Context, string) (string, error) {
	return "", nil
}

func (f *fakeFactoryCloudClient) GetCommit(context.Context, vcs.Repo, string) (vcs.Commit, error) {
	return vcs.Commit{}, nil
}
