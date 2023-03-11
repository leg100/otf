package run

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartRun(t *testing.T) {
	ctx := context.Background()

	t.Run("not connected to repo", func(t *testing.T) {
		ws := &otf.Workspace{}
		cv := &otf.ConfigurationVersion{}
		want := &Run{}
		starter := newTestStarter(fakeStarterService{
			run:       want,
			workspace: ws,
			cv:        cv,
		})

		got, err := starter.startRun(ctx, ws.ID, otf.ConfigurationVersionCreateOptions{})
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("connected to repo", func(t *testing.T) {
		ws := &otf.Workspace{Connection: &otf.Connection{}}
		cv := &otf.ConfigurationVersion{}
		provider := &otf.VCSProvider{}
		want := &Run{}
		starter := newTestStarter(fakeStarterService{
			run:       want,
			workspace: ws,
			cv:        cv,
			provider:  provider,
		})

		got, err := starter.startRun(ctx, ws.ID, otf.ConfigurationVersionCreateOptions{})
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})
}

type (
	fakeStarterService struct {
		run       *Run
		workspace *otf.Workspace
		cv        *otf.ConfigurationVersion
		provider  *otf.VCSProvider

		otf.WorkspaceService
		otf.VCSProviderService
		otf.ConfigurationVersionService
		service
	}
)

func newTestStarter(svc fakeStarterService) *starter {
	return &starter{
		ConfigurationVersionService: &svc,
		VCSProviderService:          &svc,
		WorkspaceService:            &svc,
		service:                     &svc,
	}
}

func (f *fakeStarterService) GetWorkspace(context.Context, string) (*otf.Workspace, error) {
	return f.workspace, nil
}

func (f *fakeStarterService) CreateConfigurationVersion(context.Context, string, otf.ConfigurationVersionCreateOptions) (*otf.ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *fakeStarterService) GetLatestConfigurationVersion(context.Context, string) (*otf.ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *fakeStarterService) CloneConfigurationVersion(context.Context, string, otf.ConfigurationVersionCreateOptions) (*otf.ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *fakeStarterService) UploadConfig(context.Context, string, []byte) error {
	return nil
}

func (f *fakeStarterService) GetVCSClient(context.Context, string) (cloud.Client, error) {
	return &fakeStartRunCloudClient{}, nil
}

func (f *fakeStarterService) create(context.Context, string, RunCreateOptions) (*Run, error) {
	return f.run, nil
}

type fakeStartRunCloudClient struct {
	cloud.Client
}

func (f *fakeStartRunCloudClient) GetRepoTarball(context.Context, cloud.GetRepoTarballOptions) ([]byte, error) {
	return nil, nil
}
