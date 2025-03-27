package run

import (
	"context"
	"testing"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/releases"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFactory(t *testing.T) {
	ctx := context.Background()

	t.Run("defaults", func(t *testing.T) {
		f := newTestFactory(
			&organization.Organization{},
			&workspace.Workspace{},
			&configversion.ConfigurationVersion{},
			"",
		)

		got, err := f.NewRun(ctx, nil, CreateOptions{})
		require.NoError(t, err)

		assert.Equal(t, runstatus.Pending, got.Status)
		assert.NotZero(t, got.CreatedAt)
		assert.False(t, got.PlanOnly)
		assert.True(t, got.Refresh)
		assert.False(t, got.AutoApply)
	})

	t.Run("speculative run", func(t *testing.T) {
		f := newTestFactory(
			&organization.Organization{},
			&workspace.Workspace{},
			&configversion.ConfigurationVersion{Speculative: true},
			"",
		)

		got, err := f.NewRun(ctx, nil, CreateOptions{})
		require.NoError(t, err)

		assert.True(t, got.PlanOnly)
	})

	t.Run("plan-only run", func(t *testing.T) {
		f := newTestFactory(
			&organization.Organization{},
			&workspace.Workspace{},
			&configversion.ConfigurationVersion{},
			"",
		)

		got, err := f.NewRun(ctx, nil, CreateOptions{PlanOnly: internal.Bool(true)})
		require.NoError(t, err)

		assert.True(t, got.PlanOnly)
	})

	t.Run("workspace auto-apply", func(t *testing.T) {
		f := newTestFactory(
			&organization.Organization{},
			&workspace.Workspace{AutoApply: true},
			&configversion.ConfigurationVersion{},
			"",
		)

		got, err := f.NewRun(ctx, nil, CreateOptions{})
		require.NoError(t, err)

		assert.True(t, got.AutoApply)
	})

	t.Run("run auto-apply", func(t *testing.T) {
		f := newTestFactory(
			&organization.Organization{},
			&workspace.Workspace{},
			&configversion.ConfigurationVersion{},
			"",
		)

		got, err := f.NewRun(ctx, nil, CreateOptions{
			AutoApply: internal.Bool(true),
		})
		require.NoError(t, err)

		assert.True(t, got.AutoApply)
	})

	t.Run("enable cost estimation", func(t *testing.T) {
		f := newTestFactory(
			&organization.Organization{CostEstimationEnabled: true},
			&workspace.Workspace{},
			&configversion.ConfigurationVersion{},
			"",
		)

		got, err := f.NewRun(ctx, nil, CreateOptions{})
		require.NoError(t, err)

		assert.True(t, got.CostEstimationEnabled)
	})

	t.Run("pull from vcs", func(t *testing.T) {
		workspaceID := resource.NewTfeID(resource.WorkspaceKind)

		f := newTestFactory(
			&organization.Organization{},
			&workspace.Workspace{
				ID:         workspaceID,
				Connection: &workspace.Connection{},
			},
			&configversion.ConfigurationVersion{},
			"",
		)

		got, err := f.NewRun(ctx, nil, CreateOptions{})
		require.NoError(t, err)

		// fake config version service sets the config version ID to the
		// workspace ID if it was newly created
		assert.Equal(t, workspaceID, got.ConfigurationVersionID)
	})

	t.Run("get latest version", func(t *testing.T) {
		f := newTestFactory(
			&organization.Organization{},
			&workspace.Workspace{TerraformVersion: releases.LatestVersionString},
			&configversion.ConfigurationVersion{},
			"1.2.3",
		)

		got, err := f.NewRun(ctx, nil, CreateOptions{})
		require.NoError(t, err)

		assert.Equal(t, "1.2.3", got.TerraformVersion)
	})
}

type (
	fakeFactoryOrganizationService struct {
		org *organization.Organization
	}
	fakeFactoryWorkspaceService struct {
		ws *workspace.Workspace
	}
	fakeFactoryConfigurationVersionService struct {
		cv *configversion.ConfigurationVersion
	}
	fakeFactoryVCSProviderService struct{}
	fakeFactoryCloudClient        struct {
		vcs.Client
	}
	fakeReleasesService struct {
		latestVersion string
	}
)

func newTestFactory(org *organization.Organization, ws *workspace.Workspace, cv *configversion.ConfigurationVersion, latestVersion string) *factory {
	return &factory{
		organizations: &fakeFactoryOrganizationService{org: org},
		workspaces:    &fakeFactoryWorkspaceService{ws: ws},
		configs:       &fakeFactoryConfigurationVersionService{cv: cv},
		vcs:           &fakeFactoryVCSProviderService{},
		releases:      &fakeReleasesService{latestVersion: latestVersion},
	}
}

func (f *fakeFactoryOrganizationService) Get(context.Context, resource.OrganizationName) (*organization.Organization, error) {
	return f.org, nil
}

func (f *fakeFactoryWorkspaceService) Get(context.Context, resource.ID) (*workspace.Workspace, error) {
	return f.ws, nil
}

func (f *fakeFactoryConfigurationVersionService) Get(context.Context, resource.ID) (*configversion.ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *fakeFactoryConfigurationVersionService) GetLatest(context.Context, resource.ID) (*configversion.ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *fakeFactoryConfigurationVersionService) Create(ctx context.Context, workspaceID resource.ID, opts configversion.CreateOptions) (*configversion.ConfigurationVersion, error) {
	return &configversion.ConfigurationVersion{ID: workspaceID.(resource.TfeID)}, nil
}

func (f *fakeFactoryConfigurationVersionService) UploadConfig(context.Context, resource.ID, []byte) error {
	return nil
}

func (f *fakeFactoryVCSProviderService) GetVCSClient(context.Context, resource.ID) (vcs.Client, error) {
	return &fakeFactoryCloudClient{}, nil
}

func (f *fakeFactoryCloudClient) GetRepoTarball(context.Context, vcs.GetRepoTarballOptions) ([]byte, string, error) {
	return nil, "", nil
}

func (f *fakeFactoryCloudClient) GetRepository(context.Context, string) (vcs.Repository, error) {
	return vcs.Repository{}, nil
}

func (f *fakeFactoryCloudClient) GetCommit(context.Context, string, string) (vcs.Commit, error) {
	return vcs.Commit{}, nil
}

func (f *fakeReleasesService) GetLatest(context.Context) (string, time.Time, error) {
	return f.latestVersion, time.Time{}, nil
}
