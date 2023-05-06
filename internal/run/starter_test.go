package run

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/repo"
	"github.com/leg100/otf/internal/vcsprovider"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartRun(t *testing.T) {
	ctx := context.Background()

	t.Run("not connected to repo", func(t *testing.T) {
		ws := &workspace.Workspace{}
		cv := &configversion.ConfigurationVersion{}
		want := &Run{}
		starter := newTestStarter(fakeStarterService{
			run:       want,
			workspace: ws,
			cv:        cv,
		})

		got, err := starter.startRun(ctx, ws.ID, planOnly)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("connected to repo", func(t *testing.T) {
		ws := &workspace.Workspace{Connection: &repo.Connection{}}
		cv := &configversion.ConfigurationVersion{}
		provider := &vcsprovider.VCSProvider{}
		want := &Run{}
		starter := newTestStarter(fakeStarterService{
			run:       want,
			workspace: ws,
			cv:        cv,
			provider:  provider,
		})

		got, err := starter.startRun(ctx, ws.ID, planOnly)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})
}

type (
	fakeStarterService struct {
		run       *Run
		workspace *workspace.Workspace
		cv        *configversion.ConfigurationVersion
		provider  *vcsprovider.VCSProvider

		WorkspaceService
		VCSProviderService
		ConfigurationVersionService
		RunService
	}
)

func newTestStarter(svc fakeStarterService) *starter {
	return &starter{
		ConfigurationVersionService: &svc,
		VCSProviderService:          &svc,
		WorkspaceService:            &svc,
		RunService:                  &svc,
	}
}

func (f *fakeStarterService) GetWorkspace(context.Context, string) (*workspace.Workspace, error) {
	return f.workspace, nil
}

func (f *fakeStarterService) CreateConfigurationVersion(context.Context, string, configversion.ConfigurationVersionCreateOptions) (*configversion.ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *fakeStarterService) GetLatestConfigurationVersion(context.Context, string) (*configversion.ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *fakeStarterService) CloneConfigurationVersion(context.Context, string, configversion.ConfigurationVersionCreateOptions) (*configversion.ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *fakeStarterService) UploadConfig(context.Context, string, []byte) error {
	return nil
}

func (f *fakeStarterService) GetVCSClient(context.Context, string) (cloud.Client, error) {
	return &fakeStartRunCloudClient{}, nil
}

func (f *fakeStarterService) CreateRun(context.Context, string, RunCreateOptions) (*Run, error) {
	return f.run, nil
}

type fakeStartRunCloudClient struct {
	cloud.Client
}

func (f *fakeStartRunCloudClient) GetRepoTarball(context.Context, cloud.GetRepoTarballOptions) ([]byte, error) {
	return nil, nil
}
