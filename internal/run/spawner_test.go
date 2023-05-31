package run

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/repo"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpawner(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		ws      *workspace.Workspace
		event   cloud.VCSEvent // incoming event
		spawned bool           // want spawned run
	}{
		{
			name:    "spawn run upon push to default branch",
			ws:      &workspace.Workspace{Connection: &repo.Connection{}},
			event:   cloud.VCSPushEvent{Branch: "main", DefaultBranch: "main"},
			spawned: true,
		},
		{
			name:    "skip run upon push to non-default branch",
			ws:      &workspace.Workspace{Connection: &repo.Connection{}},
			event:   cloud.VCSPushEvent{Branch: "dev", DefaultBranch: "main"},
			spawned: false,
		},
		{
			name:    "spawn run upon push to user-specified branch",
			ws:      &workspace.Workspace{Connection: &repo.Connection{}, Branch: "dev"},
			event:   cloud.VCSPushEvent{Branch: "dev"},
			spawned: true,
		},
		{
			name:    "skip run upon push to branch not matching user-specified branch",
			ws:      &workspace.Workspace{Connection: &repo.Connection{}, Branch: "dev"},
			event:   cloud.VCSPushEvent{Branch: "staging"},
			spawned: false,
		},
		{
			name:    "spawn run upon opened pr",
			ws:      &workspace.Workspace{Connection: &repo.Connection{}},
			event:   cloud.VCSPullEvent{Action: cloud.VCSPullEventOpened},
			spawned: true,
		},
		{
			name:    "spawn run upon push to pr",
			ws:      &workspace.Workspace{Connection: &repo.Connection{}},
			event:   cloud.VCSPullEvent{Action: cloud.VCSPullEventUpdated},
			spawned: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			services := &fakeSpawnerServices{
				workspaces: []*workspace.Workspace{tt.ws},
			}
			spawner := Spawner{
				ConfigurationVersionService: services,
				WorkspaceService:            services,
				VCSProviderService:          services,
				RunService:                  services,
				Logger:                      logr.Discard(),
			}
			err := spawner.handle(ctx, tt.event)
			require.NoError(t, err)

			assert.Equal(t, tt.spawned, services.spawned)
		})
	}
}

type fakeSpawnerServices struct {
	workspaces []*workspace.Workspace
	created    []*configversion.ConfigurationVersion // created config versions
	spawned    bool                                  // whether a run was spawned

	ConfigurationVersionService
	WorkspaceService
	VCSProviderService
	RunService
}

func (f *fakeSpawnerServices) ListWorkspacesByRepoID(ctx context.Context, id uuid.UUID) ([]*workspace.Workspace, error) {
	return f.workspaces, nil
}

func (f *fakeSpawnerServices) CreateConfigurationVersion(ctx context.Context, wid string, opts configversion.ConfigurationVersionCreateOptions) (*configversion.ConfigurationVersion, error) {
	cv, err := configversion.NewConfigurationVersion(wid, opts)
	if err != nil {
		return nil, err
	}
	f.created = append(f.created, cv)
	return cv, nil
}

func (f *fakeSpawnerServices) UploadConfig(context.Context, string, []byte) error {
	return nil
}

func (f *fakeSpawnerServices) CreateRun(context.Context, string, RunCreateOptions) (*Run, error) {
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
