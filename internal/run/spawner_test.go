package run

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpawner(t *testing.T) {
	tests := []struct {
		name string
		ws   *workspace.Workspace
		// incoming event
		event vcs.Event
		// file paths to return from stubbed client.ListPullRequestFiles
		pullFiles []string
		// want spawned run
		spawn bool
	}{
		{
			name: "spawn run for push to default branch",
			ws:   &workspace.Workspace{Connection: &workspace.Connection{}},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:          vcs.EventTypePush,
					Action:        vcs.ActionCreated,
					Branch:        "main",
					DefaultBranch: "main",
				},
			},
			spawn: true,
		},
		{
			name: "skip run for push to non-default branch",
			ws:   &workspace.Workspace{Connection: &workspace.Connection{}},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:          vcs.EventTypePush,
					Action:        vcs.ActionCreated,
					Branch:        "dev",
					DefaultBranch: "main",
				},
			},
			spawn: false,
		},
		{
			name: "spawn run for push event for a workspace with user-specified branch",
			ws:   &workspace.Workspace{Connection: &workspace.Connection{Branch: "dev"}},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:   vcs.EventTypePush,
					Action: vcs.ActionCreated,
					Branch: "dev",
				},
			},
			spawn: true,
		},
		{
			name: "skip run for push event for a workspace with non-matching, user-specified branch",
			ws:   &workspace.Workspace{Connection: &workspace.Connection{Branch: "dev"}},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:   vcs.EventTypePush,
					Action: vcs.ActionCreated,
					Branch: "staging",
				},
			},
			spawn: false,
		},
		{
			name: "spawn run for opened pull request",
			ws:   &workspace.Workspace{Connection: &workspace.Connection{}},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type: vcs.EventTypePull, Action: vcs.ActionCreated,
				},
			},
			spawn: true,
		},
		{
			name: "spawn run for update to pull request",
			ws:   &workspace.Workspace{Connection: &workspace.Connection{}},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:   vcs.EventTypePull,
					Action: vcs.ActionUpdated,
				},
			},
			spawn: true,
		},
		{
			name: "skip run for push event for workspace with tags regex",
			ws:   &workspace.Workspace{Connection: &workspace.Connection{TagsRegex: "0.1.2"}},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{Type: vcs.EventTypePush, Action: vcs.ActionCreated},
			},
			spawn: false,
		},
		{
			name: "spawn run for tag event for workspace with matching tags regex",
			ws: &workspace.Workspace{Connection: &workspace.Connection{
				TagsRegex: `^\d+\.\d+\.\d+$`,
			}},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:   vcs.EventTypeTag,
					Action: vcs.ActionCreated,
					Tag:    "0.1.2",
				},
			},
			spawn: true,
		},
		{
			name: "skip run for tag event for workspace with non-matching tags regex",
			ws: &workspace.Workspace{Connection: &workspace.Connection{
				TagsRegex: `^\d+\.\d+\.\d+$`,
			}},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:   vcs.EventTypeTag,
					Action: vcs.ActionCreated,
					Tag:    "v0.1.2",
				},
			},
			spawn: false,
		},
		{
			name: "spawn run for push event for workspace with matching file trigger pattern",
			ws: &workspace.Workspace{
				TriggerPatterns: []string{"/foo/*.tf"},
				Connection:      &workspace.Connection{},
			},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:   vcs.EventTypePush,
					Action: vcs.ActionCreated,
					Paths:  []string{"/foo/bar.tf"},
				},
			},
			spawn: true,
		},
		{
			name: "skip run for push event for workspace with non-matching file trigger pattern",
			ws: &workspace.Workspace{
				TriggerPatterns: []string{"/foo/*.tf"},
				Connection:      &workspace.Connection{},
			},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:   vcs.EventTypePush,
					Action: vcs.ActionCreated,
					Paths:  []string{"README.md", ".gitignore"},
				},
			},
			spawn: false,
		},
		{
			name: "spawn run for pull event for workspace with matching file trigger pattern",
			ws: &workspace.Workspace{
				TriggerPatterns: []string{"/foo/*.tf"},
				Connection:      &workspace.Connection{},
			},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:   vcs.EventTypePull,
					Action: vcs.ActionUpdated,
				},
			},
			pullFiles: []string{"/foo/bar.tf"},
			spawn:     true,
		},
		{
			name: "skip run for pull event for workspace with non-matching file trigger pattern",
			ws: &workspace.Workspace{
				TriggerPatterns: []string{"/foo/*.tf"},
				Connection:      &workspace.Connection{},
			},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:   vcs.EventTypePull,
					Action: vcs.ActionUpdated,
				},
			},
			pullFiles: []string{"README.md", ".gitignore"},
			spawn:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			services := &fakeSpawnerServices{
				workspaces: []*workspace.Workspace{tt.ws},
				pullFiles:  tt.pullFiles,
			}
			spawner := Spawner{
				ConfigurationVersionService: services,
				WorkspaceService:            services,
				VCSProviderService:          services,
				RunService:                  services,
			}
			err := spawner.handleWithError(logr.Discard(), tt.event)
			require.NoError(t, err)

			assert.Equal(t, tt.spawn, services.spawned)
		})
	}
}

type fakeSpawnerServices struct {
	// workspaces to return from stubbed ListWorkspacesByRepoID()
	workspaces []*workspace.Workspace
	// created config versions
	created []*configversion.ConfigurationVersion
	// whether a run was spawned
	spawned bool
	// list of file paths to return from stubbed ListPullRequestFiles()
	pullFiles []string

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

func (f *fakeSpawnerServices) CreateRun(context.Context, string, CreateOptions) (*Run, error) {
	f.spawned = true
	return nil, nil
}

func (f *fakeSpawnerServices) GetVCSClient(context.Context, string) (vcs.Client, error) {
	return &fakeSpawnerCloudClient{pullFiles: f.pullFiles}, nil
}

type fakeSpawnerCloudClient struct {
	vcs.Client
	pullFiles []string
}

func (f *fakeSpawnerCloudClient) GetRepoTarball(context.Context, vcs.GetRepoTarballOptions) ([]byte, string, error) {
	return nil, "", nil
}

func (f *fakeSpawnerCloudClient) ListPullRequestFiles(ctx context.Context, repo string, pull int) ([]string, error) {
	return f.pullFiles, nil
}
