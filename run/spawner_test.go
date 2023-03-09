package run

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTriggerer(t *testing.T) {
	ctx := context.Background()
	org := otf.NewTestOrganization(t)
	provider := otf.NewTestVCSProvider(t, org)
	repo := otf.NewTestWorkspaceRepo(provider)

	tests := []struct {
		name    string
		ws      *otf.Workspace
		event   cloud.VCSEvent
		trigger bool
	}{
		{
			name:    "trigger run for push to default branch",
			ws:      otf.NewTestWorkspace(t, org, otf.WithRepo(repo)),
			event:   cloud.VCSPushEvent{Branch: "main", DefaultBranch: "main"},
			trigger: true,
		},
		{
			name:    "skip run for push to non-default branch",
			ws:      otf.NewTestWorkspace(t, org, otf.WithRepo(repo)),
			event:   cloud.VCSPushEvent{Branch: "dev", DefaultBranch: "main"},
			trigger: false,
		},
		{
			name:    "trigger run for push to user-specified branch",
			ws:      otf.NewTestWorkspace(t, org, otf.WithRepo(repo), otf.WithBranch("dev")),
			event:   cloud.VCSPushEvent{Branch: "dev"},
			trigger: true,
		},
		{
			name:    "skip run for push to branch not matching user-specified branch",
			ws:      otf.NewTestWorkspace(t, org, otf.WithRepo(repo), otf.WithBranch("dev")),
			event:   cloud.VCSPushEvent{Branch: "staging"},
			trigger: false,
		},
		{
			name:    "trigger run for opened pr",
			ws:      otf.NewTestWorkspace(t, org, otf.WithRepo(repo)),
			event:   cloud.VCSPullEvent{Action: cloud.VCSPullEventOpened},
			trigger: true,
		},
		{
			name:    "trigger run for push to pr",
			ws:      otf.NewTestWorkspace(t, org, otf.WithRepo(repo)),
			event:   cloud.VCSPullEvent{Action: cloud.VCSPullEventUpdated},
			trigger: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &fakeTriggererApp{
				workspaces: []*otf.Workspace{tt.ws},
			}
			triggerer := Triggerer{
				Application: app,
				Logger:      logr.Discard(),
			}
			err := triggerer.handle(ctx, tt.event)
			require.NoError(t, err)

			assert.Equal(t, tt.trigger, app.triggered)
		})
	}
}

type fakeTriggererApp struct {
	workspaces []*otf.Workspace
	created    []*otf.ConfigurationVersion // created config versions
	triggered  bool                        // whether a run was triggered

	otf.Application
}

func (f *fakeTriggererApp) ListWorkspacesByWebhookID(ctx context.Context, id uuid.UUID) ([]*otf.Workspace, error) {
	return f.workspaces, nil
}

func (f *fakeTriggererApp) CreateConfigurationVersion(ctx context.Context, wid string, opts otf.ConfigurationVersionCreateOptions) (*otf.ConfigurationVersion, error) {
	cv, err := otf.NewConfigurationVersion(wid, opts)
	if err != nil {
		return nil, err
	}
	f.created = append(f.created, cv)
	return cv, nil
}

func (f *fakeTriggererApp) UploadConfig(context.Context, string, []byte) error {
	return nil
}

func (f *fakeTriggererApp) CreateRun(context.Context, string, otf.RunCreateOptions) (*otf.Run, error) {
	f.triggered = true
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
