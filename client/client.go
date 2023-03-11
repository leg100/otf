// Package client allows remote interaction with the otf application
package client

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/http"
	"github.com/leg100/otf/state"
	"github.com/leg100/otf/variable"
	"github.com/leg100/otf/watch"
)

// Client is those service endpoints that support both in-process and remote
// invocation. Intended for use with the agent (the internal agent is
// in-process, while the external agent is remote) as well as the CLI.
type Client interface {
	CreateOrganization(ctx context.Context, opts otf.OrganizationCreateOptions) (otf.Organization, error)

	GetWorkspace(ctx context.Context, workspaceID string) (otf.Workspace, error)
	GetWorkspaceByName(ctx context.Context, organization, workspace string) (otf.Workspace, error)
	ListWorkspaces(ctx context.Context, opts otf.WorkspaceListOptions) (otf.WorkspaceList, error)
	UpdateWorkspace(ctx context.Context, workspaceID string, opts otf.UpdateWorkspaceOptions) (*otf.Workspace, error)

	ListVariables(ctx context.Context, workspaceID string) ([]otf.Variable, error)

	CreateAgentToken(ctx context.Context, opts otf.CreateAgentTokenOptions) (*otf.AgentToken, error)
	GetAgentToken(ctx context.Context, token string) (*otf.AgentToken, error)

	GetPlanFile(ctx context.Context, id string, format otf.PlanFormat) ([]byte, error)
	UploadPlanFile(ctx context.Context, id string, plan []byte, format otf.PlanFormat) error
	GetLockFile(ctx context.Context, id string) ([]byte, error)
	UploadLockFile(ctx context.Context, id string, lockFile []byte) error
	ListRuns(ctx context.Context, opts run.RunListOptions) (*run.RunList, error)
	GetRun(ctx context.Context, id string) (*run.Run, error)
	StartPhase(ctx context.Context, id string, phase otf.PhaseType, opts otf.PhaseStartOptions) (*run.Run, error)
	FinishPhase(ctx context.Context, id string, phase otf.PhaseType, opts otf.PhaseFinishOptions) (*run.Run, error)

	PutChunk(ctx context.Context, chunk otf.Chunk) error

	DownloadConfig(ctx context.Context, id string) ([]byte, error)

	Watch(context.Context, otf.WatchOptions) (<-chan otf.Event, error)

	// CreateRegistrySession creates a registry session for the given organization.
	CreateRegistrySession(ctx context.Context, organization string) (otf.RegistrySession, error)

	otf.WorkspaceLockService
	otf.StateVersionApp
	otf.HostnameService
}

type (
	stateClient    = state.Client
	variableClient = variable.Client
	authClient     = auth.Client
	watchClient    = watch.Client
)

type client struct {
	*http.Client
	http.Config

	stateClient
	variableClient
	authClient
	watchClient
}

// New constructs a client that uses the http to remotely invoke OTF
// services.
func New(config http.Config) (*client, error) {
	httpClient, err := http.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &client{
		Client:         httpClient,
		stateClient:    stateClient{httpClient},
		variableClient: variableClient{httpClient},
		authClient:     authClient{httpClient},
		watchClient:    watchClient{config},
	}, nil
}
