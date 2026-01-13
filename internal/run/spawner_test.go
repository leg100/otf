package run

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/resource"
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
			ws: &workspace.Workspace{
				Connection:         &workspace.Connection{},
				SpeculativeEnabled: true,
			},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type: vcs.EventTypePull, Action: vcs.ActionCreated,
				},
			},
			spawn: true,
		},
		{
			name: "spawn run for update to pull request",
			ws: &workspace.Workspace{
				Connection:         &workspace.Connection{},
				SpeculativeEnabled: true,
			},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:   vcs.EventTypePull,
					Action: vcs.ActionUpdated,
				},
			},
			spawn: true,
		},
		{
			name: "skip run for pull event for a workspace with speculative plans disabled",
			ws: &workspace.Workspace{
				Connection:         &workspace.Connection{},
				SpeculativeEnabled: false,
			},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:   vcs.EventTypePull,
					Action: vcs.ActionCreated,
				},
			},
			spawn: false,
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
				TriggerPatterns:    []string{"/foo/*.tf"},
				Connection:         &workspace.Connection{},
				SpeculativeEnabled: true,
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
			runClient := &fakeSpawnerRunClient{}
			spawner := Spawner{
				configs: &configversion.FakeService{},
				workspaces: &workspace.FakeService{
					Workspaces: []*workspace.Workspace{tt.ws},
				},
				runs: runClient,
				vcs: &fakeSpawnerVCSProviderClient{
					pullFiles: tt.pullFiles,
				},
			}
			err := spawner.handleWithError(logr.Discard(), tt.event)
			require.NoError(t, err)

			assert.Equal(t, tt.spawn, runClient.spawned)
		})
	}
}

type fakeSpawnerRunClient struct {
	// whether a run was spawned
	spawned bool
}

func (f *fakeSpawnerRunClient) Create(context.Context, resource.TfeID, CreateOptions) (*Run, error) {
	f.spawned = true
	return nil, nil
}

type fakeSpawnerVCSProviderClient struct {
	// list of file paths to return from stubbed ListPullRequestFiles()
	pullFiles []string
}

func (f *fakeSpawnerVCSProviderClient) Get(context.Context, resource.TfeID) (*vcs.Provider, error) {
	return &vcs.Provider{
		Client: &fakeSpawnerCloudClient{pullFiles: f.pullFiles},
	}, nil
}

type fakeSpawnerCloudClient struct {
	vcs.Client
	pullFiles []string
}

func (f *fakeSpawnerCloudClient) GetRepoTarball(context.Context, vcs.GetRepoTarballOptions) ([]byte, string, error) {
	return nil, "", nil
}

func (f *fakeSpawnerCloudClient) ListPullRequestFiles(ctx context.Context, repo vcs.Repo, pull int) ([]string, error) {
	return f.pullFiles, nil
}
