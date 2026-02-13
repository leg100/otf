package workspace

import (
	"errors"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWorkspace(t *testing.T) {
	var (
		org1          = organization.NewTestName(t)
		agentPoolID   = testutils.ParseID(t, "apool-123")
		vcsProviderID = testutils.ParseID(t, "vcs-123")
		latestVersion = "1.9.0"
		factory       = &factory{
			defaultEngine: engine.Default,
			engines: &fakeReleasesService{
				latestVersion: latestVersion,
			},
		}
	)

	tests := []struct {
		name      string
		opts      CreateOptions
		wantError error
		test      func(t *testing.T, got *Workspace)
	}{
		{
			name: "default",
			opts: CreateOptions{
				Name:         new("my-workspace"),
				Organization: &org1,
			},
			test: func(t *testing.T, got *Workspace) {
				assert.Equal(t, "my-workspace", got.Name)
				assert.Equal(t, org1, got.Organization)
				assert.Equal(t, &Version{semver: latestVersion}, got.EngineVersion)
			},
		},
		{
			name: "missing name",
			opts: CreateOptions{
				Organization: &org1,
			},
			wantError: internal.ErrRequiredName,
		},
		{
			name: "missing organization",
			opts: CreateOptions{
				Name: new("my-workspace"),
			},
			wantError: internal.ErrRequiredOrg,
		},
		{
			name: "invalid name",
			opts: CreateOptions{
				Name: new("%*&^"),
			},
			wantError: internal.ErrInvalidName,
		},
		{
			name: "specifying both tags regex and trigger patterns",
			opts: CreateOptions{
				Name:            new("my-workspace"),
				Organization:    &org1,
				TriggerPatterns: []string{"/foo/**/*.tf"},
				ConnectOptions: &ConnectOptions{
					RepoPath:      new(vcs.NewMustRepo("leg100", "otf")),
					VCSProviderID: &vcsProviderID,
					TagsRegex:     new("\\d+"),
				},
			},
			wantError: ErrTagsRegexAndTriggerPatterns,
		},
		{
			name: "specifying trigger patterns but empty string for tags regex is ok",
			opts: CreateOptions{
				Name:            new("my-workspace"),
				Organization:    &org1,
				TriggerPatterns: []string{"/foo/**/*.tf"},
				ConnectOptions: &ConnectOptions{
					RepoPath:      new(vcs.NewMustRepo("leg100", "otf")),
					VCSProviderID: &vcsProviderID,
					TagsRegex:     new(""),
				},
			},
			wantError: nil,
		},
		{
			name: "specifying both tags regex and always trigger",
			opts: CreateOptions{
				Name:          new("my-workspace"),
				Organization:  &org1,
				AlwaysTrigger: new(true),
				ConnectOptions: &ConnectOptions{
					TagsRegex: new("\\d+"),
				},
			},
			wantError: ErrTagsRegexAndAlwaysTrigger,
		},
		{
			name: "specifying both trigger patterns and always trigger",
			opts: CreateOptions{
				Name:            new("my-workspace"),
				Organization:    &org1,
				AlwaysTrigger:   new(true),
				TriggerPatterns: []string{"/foo/**/*.tf"},
			},
			wantError: ErrTriggerPatternsAndAlwaysTrigger,
		},
		{
			name: "invalid trigger pattern",
			opts: CreateOptions{
				Name:            new("my-workspace"),
				Organization:    &org1,
				TriggerPatterns: []string{"/foo/[**/*.tf"},
			},
			wantError: ErrInvalidTriggerPattern,
		},
		{
			name: "invalid tags regex",
			opts: CreateOptions{
				Name:         new("my-workspace"),
				Organization: &org1,
				ConnectOptions: &ConnectOptions{
					RepoPath:      new(vcs.NewMustRepo("leg100", "otf")),
					VCSProviderID: &vcsProviderID,
					TagsRegex:     new("{**"),
				},
			},
			wantError: ErrInvalidTagsRegex,
		},
		{
			name: "agent execution mode with agent pool ID",
			opts: CreateOptions{
				Name:          new("my-workspace"),
				Organization:  &org1,
				ExecutionMode: new(AgentExecutionMode),
				AgentPoolID:   &agentPoolID,
			},
			wantError: nil,
		},
		{
			name: "agent execution mode without agent pool ID",
			opts: CreateOptions{
				Name:          new("my-workspace"),
				Organization:  &org1,
				ExecutionMode: new(AgentExecutionMode),
			},
			wantError: ErrAgentExecutionModeWithoutPool,
		},
		{
			name: "default remote execution mode with agent pool ID",
			opts: CreateOptions{
				Name:         new("my-workspace"),
				Organization: &org1,
				AgentPoolID:  &agentPoolID,
			},
			wantError: ErrNonAgentExecutionModeWithPool,
		},
		{
			name: "local execution mode with agent pool ID",
			opts: CreateOptions{
				Name:          new("my-workspace"),
				Organization:  &org1,
				ExecutionMode: new(LocalExecutionMode),
				AgentPoolID:   &agentPoolID,
			},
			wantError: ErrNonAgentExecutionModeWithPool,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := factory.NewWorkspace(t.Context(), tt.opts)
			assert.True(t, errors.Is(err, tt.wantError), "got: %s", err)
			if tt.test != nil {
				tt.test(t, got)
			}
		})
	}
}

func TestWorkspace_UpdateError(t *testing.T) {
	org1 := organization.NewTestName(t)
	agentPoolID := testutils.ParseID(t, "apool-123")
	vcsProviderID := testutils.ParseID(t, "vcs-123")

	tests := []struct {
		name string
		ws   *Workspace
		opts UpdateOptions
		want error
	}{
		{
			name: "invalid name",
			ws:   &Workspace{Name: "dev", Organization: org1},
			opts: UpdateOptions{
				Name: new("%*&^"),
			},
			want: internal.ErrInvalidName,
		},
		{
			name: "specifying both tags regex and trigger patterns",
			ws:   &Workspace{Name: "dev", Organization: org1},
			opts: UpdateOptions{
				Name:            new("my-workspace"),
				TriggerPatterns: []string{"/foo/**/*.tf"},
				ConnectOptions: &ConnectOptions{
					RepoPath:      new(vcs.NewMustRepo("leg100", "otf")),
					VCSProviderID: &vcsProviderID,
					TagsRegex:     new("\\d+"),
				},
			},
			want: ErrTagsRegexAndTriggerPatterns,
		},
		{
			name: "specifying trigger patterns but empty string for tags regex is ok",
			ws:   &Workspace{Name: "dev", Organization: org1},
			opts: UpdateOptions{
				Name:            new("my-workspace"),
				TriggerPatterns: []string{"/foo/**/*.tf"},
				ConnectOptions: &ConnectOptions{
					RepoPath:      new(vcs.NewMustRepo("leg100", "otf")),
					VCSProviderID: &vcsProviderID,
					TagsRegex:     new(""),
				},
			},
			want: nil,
		},
		{
			name: "specifying both tags regex and always trigger",
			ws:   &Workspace{Name: "dev", Organization: org1},
			opts: UpdateOptions{
				Name:          new("my-workspace"),
				AlwaysTrigger: new(true),
				ConnectOptions: &ConnectOptions{
					TagsRegex: new("\\d+"),
				},
			},
			want: ErrTagsRegexAndAlwaysTrigger,
		},
		{
			name: "specifying both trigger patterns and always trigger",
			ws:   &Workspace{Name: "dev", Organization: org1},
			opts: UpdateOptions{
				Name:            new("my-workspace"),
				AlwaysTrigger:   new(true),
				TriggerPatterns: []string{"/foo/**/*.tf"},
			},
			want: ErrTriggerPatternsAndAlwaysTrigger,
		},
		{
			name: "invalid trigger pattern",
			ws:   &Workspace{Name: "dev", Organization: org1},
			opts: UpdateOptions{
				Name:            new("my-workspace"),
				TriggerPatterns: []string{"/foo/[**/*.tf"},
			},
			want: ErrInvalidTriggerPattern,
		},
		{
			name: "invalid tags regex",
			ws:   &Workspace{Name: "dev", Organization: org1},
			opts: UpdateOptions{
				Name: new("my-workspace"),
				ConnectOptions: &ConnectOptions{
					RepoPath:      new(vcs.NewMustRepo("leg100", "otf")),
					VCSProviderID: &vcsProviderID,
					TagsRegex:     new("{**"),
				},
			},
			want: ErrInvalidTagsRegex,
		},
		{
			name: "agent execution mode with agent pool ID",
			ws:   &Workspace{Name: "dev", Organization: org1},
			opts: UpdateOptions{
				ExecutionMode: new(AgentExecutionMode),
				AgentPoolID:   &agentPoolID,
			},
			want: nil,
		},
		{
			name: "agent execution mode without agent pool ID",
			ws:   &Workspace{Name: "dev", Organization: org1},
			opts: UpdateOptions{
				ExecutionMode: new(AgentExecutionMode),
			},
			want: ErrAgentExecutionModeWithoutPool,
		},
		{
			name: "existing agent execution mode with updated agent pool ID",
			ws:   &Workspace{Name: "dev", Organization: org1, ExecutionMode: AgentExecutionMode, AgentPoolID: &agentPoolID},
			opts: UpdateOptions{
				AgentPoolID: &agentPoolID,
			},
			want: nil,
		},
		{
			name: "existing remote execution mode with updated agent pool ID",
			ws:   &Workspace{Name: "dev", Organization: org1, ExecutionMode: RemoteExecutionMode},
			opts: UpdateOptions{
				AgentPoolID: &agentPoolID,
			},
			want: ErrNonAgentExecutionModeWithPool,
		},
		{
			name: "set local execution mode with agent pool ID",
			ws:   &Workspace{Name: "dev", Organization: org1, ExecutionMode: RemoteExecutionMode},
			opts: UpdateOptions{
				ExecutionMode: new(LocalExecutionMode),
				AgentPoolID:   &agentPoolID,
			},
			want: ErrNonAgentExecutionModeWithPool,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.ws.Update(tt.opts)
			assert.True(t, errors.Is(err, tt.want), "got: %s", err)
		})
	}
}

func TestWorkspace_Update(t *testing.T) {
	org1 := organization.NewTestName(t)
	tests := []struct {
		name string
		ws   *Workspace
		opts UpdateOptions
		want func(t *testing.T, got *Workspace)
	}{
		{
			name: "default",
			ws:   &Workspace{Name: "dev", Organization: org1},
			opts: UpdateOptions{
				Name: new("my-workspace"),
			},
			want: func(t *testing.T, got *Workspace) {
				assert.Equal(t, "my-workspace", got.Name)
			},
		},
		{
			name: "set trigger patterns",
			ws:   &Workspace{Name: "dev", Organization: org1},
			opts: UpdateOptions{
				TriggerPatterns: []string{"/foo/**/*.tf"},
			},
			want: func(t *testing.T, got *Workspace) {
				assert.Equal(t, []string{"/foo/**/*.tf"}, got.TriggerPatterns)
			},
		},
		{
			name: "trigger patterns to tags regex",
			ws: &Workspace{
				Name:            "dev",
				Organization:    org1,
				TriggerPatterns: []string{"/foo/**/*.tf"},
				Connection:      &Connection{},
			},
			opts: UpdateOptions{
				ConnectOptions: &ConnectOptions{
					TagsRegex: new("\\d+"),
				},
			},
			want: func(t *testing.T, got *Workspace) {
				assert.Nil(t, got.TriggerPatterns)
				assert.Equal(t, "\\d+", got.Connection.TagsRegex)
			},
		},
		{
			name: "switch engine from terraform to tofu",
			ws: &Workspace{
				Name:         "dev",
				Organization: org1,
				Engine:       engine.Terraform,
			},
			opts: UpdateOptions{Engine: engine.Tofu},
			want: func(t *testing.T, got *Workspace) {
				assert.Equal(t, engine.Tofu, got.Engine)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.ws.Update(tt.opts)
			require.NoError(t, err)
			tt.want(t, tt.ws)
		})
	}
}

func TestWorkspace_UpdateConnection(t *testing.T) {
	org1 := organization.NewTestName(t)
	vcsProviderID := testutils.ParseID(t, "vcs-123")

	tests := []struct {
		name string
		ws   *Workspace
		opts UpdateOptions
		want *bool
	}{
		{
			name: "connect",
			ws:   &Workspace{Name: "dev", Organization: org1},
			opts: UpdateOptions{
				Name: new("my-workspace"),
				ConnectOptions: &ConnectOptions{
					RepoPath:      new(vcs.NewMustRepo("leg100", "otf")),
					VCSProviderID: &vcsProviderID,
				},
			},
			want: new(true),
		},
		{
			name: "disconnect",
			ws: &Workspace{
				Name:         "dev",
				Organization: org1,
				Connection:   &Connection{},
			},
			opts: UpdateOptions{
				Name:       new("my-workspace"),
				Disconnect: true,
			},
			want: new(false),
		},
		{
			name: "modify connection",
			ws: &Workspace{
				Name:         "dev",
				Organization: org1,
				Connection: &Connection{
					Repo:          vcs.NewMustRepo("leg100", "otf"),
					VCSProviderID: testutils.ParseID(t, "vcs-123"),
				},
			},
			opts: UpdateOptions{
				Name: new("my-workspace"),
				ConnectOptions: &ConnectOptions{
					RepoPath: new(vcs.NewMustRepo("leg100", "otf-demo")),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ws.Update(tt.opts)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

var (
	bobby         = user.MustUsername("bobby")
	burglarTestID = user.MustUsername("burglar")
	runTestID1    = resource.NewTfeID(resource.RunKind)
	runTestID2    = resource.NewTfeID(resource.RunKind)
)

func TestWorkspace_Lock(t *testing.T) {
	t.Run("lock an unlocked lock", func(t *testing.T) {
		ws := &Workspace{}
		assert.False(t, ws.Locked())
		err := ws.Enlock(bobby)
		require.NoError(t, err)
		assert.True(t, ws.Locked())
	})
	t.Run("replace run lock with another run lock", func(t *testing.T) {
		ws := &Workspace{Lock: runTestID1}
		err := ws.Enlock(runTestID2)
		require.NoError(t, err)
		assert.True(t, ws.Locked())
	})
	t.Run("user cannot lock a locked workspace", func(t *testing.T) {
		ws := &Workspace{Lock: runTestID1}
		err := ws.Enlock(bobby)
		require.Equal(t, ErrWorkspaceAlreadyLocked, err)
	})
}

func TestWorkspace_Unlock(t *testing.T) {
	t.Run("cannot unlock workspace already unlocked", func(t *testing.T) {
		err := (&Workspace{}).Unlock(bobby, false)
		require.Equal(t, ErrWorkspaceAlreadyUnlocked, err)
	})
	t.Run("user can unlock their own lock", func(t *testing.T) {
		ws := &Workspace{Lock: bobby}
		err := ws.Unlock(bobby, false)
		require.NoError(t, err)
		assert.False(t, ws.Locked())
	})
	t.Run("user cannot unlock another user's lock", func(t *testing.T) {
		ws := &Workspace{Lock: bobby}
		err := ws.Unlock(burglarTestID, false)
		require.Equal(t, ErrWorkspaceLockedByDifferentUser, err)
	})
	t.Run("user can unlock a lock by force", func(t *testing.T) {
		ws := &Workspace{Lock: bobby}
		err := ws.Unlock(burglarTestID, true)
		require.NoError(t, err)
		assert.False(t, ws.Locked())
	})
	t.Run("run can unlock its own lock", func(t *testing.T) {
		ws := &Workspace{Lock: runTestID1}
		err := ws.Unlock(runTestID1, false)
		require.NoError(t, err)
		assert.False(t, ws.Locked())
	})
}
