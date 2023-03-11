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

func TestSpawner(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		ws      *otf.Workspace
		event   cloud.VCSEvent // incoming event
		spawned bool           // want spawned run
	}{
		{
			name:    "spawn run for push to default branch",
			ws:      &otf.Workspace{Connection: &otf.Connection{}},
			event:   cloud.VCSPushEvent{Branch: "main", DefaultBranch: "main"},
			spawned: true,
		},
		{
			name:    "skip run for push to non-default branch",
			ws:      &otf.Workspace{Connection: &otf.Connection{}},
			event:   cloud.VCSPushEvent{Branch: "dev", DefaultBranch: "main"},
			spawned: false,
		},
		{
			name:    "spawn run for push to user-specified branch",
			ws:      &otf.Workspace{Connection: &otf.Connection{}, Branch: "dev"},
			event:   cloud.VCSPushEvent{Branch: "dev"},
			spawned: true,
		},
		{
			name:    "skip run for push to branch not matching user-specified branch",
			ws:      &otf.Workspace{Connection: &otf.Connection{}, Branch: "dev"},
			event:   cloud.VCSPushEvent{Branch: "staging"},
			spawned: false,
		},
		{
			name:    "spawn run for opened pr",
			ws:      &otf.Workspace{Connection: &otf.Connection{}},
			event:   cloud.VCSPullEvent{Action: cloud.VCSPullEventOpened},
			spawned: true,
		},
		{
			name:    "spawn run for push to pr",
			ws:      &otf.Workspace{Connection: &otf.Connection{}},
			event:   cloud.VCSPullEvent{Action: cloud.VCSPullEventUpdated},
			spawned: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			services := &fakeSpawnerServices{
				workspaces: []*otf.Workspace{tt.ws},
			}
			spawner := spawner{
				ConfigurationVersionService: services,
				WorkspaceService:            services,
				VCSProviderService:          services,
				Logger:                      logr.Discard(),
				service:                     services,
			}
			err := spawner.handle(ctx, tt.event)
			require.NoError(t, err)

			assert.Equal(t, tt.spawned, services.spawned)
		})
	}
}

type fakeSpawnerServices struct {
	workspaces []*otf.Workspace
	created    []*otf.ConfigurationVersion // created config versions
	spawned    bool                        // whether a run was spawned

	otf.ConfigurationVersionService
	otf.WorkspaceService
	otf.VCSProviderService

	service
}

func (f *fakeSpawnerServices) ListWorkspacesByRepoID(ctx context.Context, id uuid.UUID) ([]*otf.Workspace, error) {
	return f.workspaces, nil
}

func (f *fakeSpawnerServices) CreateConfigurationVersion(ctx context.Context, wid string, opts otf.ConfigurationVersionCreateOptions) (*otf.ConfigurationVersion, error) {
	cv, err := otf.NewConfigurationVersion(wid, opts)
	if err != nil {
		return nil, err
	}
	f.created = append(f.created, cv)
	return cv, nil
}

func (f *fakeSpawnerServices) UploadConfig(context.Context, string, []byte) error {
	return nil
}

func (f *fakeSpawnerServices) create(context.Context, string, RunCreateOptions) (*Run, error) {
	f.spawned = true
	return nil, nil
}

func (f *fakeSpawnerServices) GetVCSClient(context.Context, string) (cloud.Client, error) {
	return &fakeSpawnerCloudClient{}, nil
}

type fakeSpawnerCloudClient struct {
	cloud.Client
}

func (f *fakeSpawnerCloudClient) GetRepoTarball(context.Context, cloud.GetRepoTarballOptions) ([]byte, error) {
	return nil, nil
}
