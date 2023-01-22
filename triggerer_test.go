package otf

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTriggerer(t *testing.T) {
	org := NewTestOrganization(t)
	provider := NewTestVCSProvider(t, org)
	hook := NewTestWebhook(t, cloud.NewTestRepo(), "fake-cloud")
	repo := NewTestWorkspaceRepo(provider, hook)
	app := &fakeTriggererApp{
		workspaces: []*Workspace{
			NewTestWorkspace(t, org, WithRepo(repo)),
			NewTestWorkspace(t, org, WithRepo(repo)),
			NewTestWorkspace(t, org, WithRepo(repo)),
		},
	}
	triggerer := Triggerer{
		Application: app,
		Logger:      logr.Discard(),
	}

	err := triggerer.handle(context.Background(), cloud.VCSPushEvent{
		Branch: "main",
	})
	require.NoError(t, err)

	assert.Equal(t, 3, len(app.created))
}

type fakeTriggererApp struct {
	workspaces []*Workspace
	created    []*ConfigurationVersion // created config versions

	Application
}

func (f *fakeTriggererApp) ListWorkspacesByWebhookID(ctx context.Context, id uuid.UUID) ([]*Workspace, error) {
	return f.workspaces, nil
}

func (f *fakeTriggererApp) CreateConfigurationVersion(ctx context.Context, wid string, opts ConfigurationVersionCreateOptions) (*ConfigurationVersion, error) {
	cv, err := NewConfigurationVersion(wid, opts)
	if err != nil {
		return nil, err
	}
	f.created = append(f.created, cv)
	return cv, nil
}

func (f *fakeTriggererApp) UploadConfig(context.Context, string, []byte) error {
	return nil
}

func (f *fakeTriggererApp) CreateRun(context.Context, string, RunCreateOptions) (*Run, error) {
	return nil, nil
}

func (f *fakeTriggererApp) GetVCSClient(context.Context, string) (cloud.Client, error) {
	return &fakeTriggererCloudClient{}, nil
}

type fakeTriggererCloudClient struct {
	cloud.Client
}

func (f *fakeTriggererCloudClient) GetRepoTarball(context.Context, cloud.GetRepoTarballOptions) ([]byte, error) {
	return nil, nil
}
