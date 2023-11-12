package remoteops

import (
	"bytes"
	"context"

	"github.com/leg100/otf/internal"
	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/logs"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/workspace"
)

var (
	_ client = (*InProcClient)(nil)
	_ client = (*rpcClient)(nil)
)

type (
	// client allows the daemon to communicate with the server endpoints.
	client interface {
		GetWorkspace(ctx context.Context, workspaceID string) (*workspace.Workspace, error)
		ListEffectiveVariables(ctx context.Context, runID string) ([]*variable.Variable, error)
		GetPlanFile(ctx context.Context, id string, format run.PlanFormat) ([]byte, error)
		UploadPlanFile(ctx context.Context, id string, plan []byte, format run.PlanFormat) error
		GetLockFile(ctx context.Context, id string) ([]byte, error)
		UploadLockFile(ctx context.Context, id string, lockFile []byte) error
		ListRuns(ctx context.Context, opts run.ListOptions) (*resource.Page[*run.Run], error)
		StartPhase(ctx context.Context, id string, phase internal.PhaseType, opts run.PhaseStartOptions) (*run.Run, error)
		FinishPhase(ctx context.Context, id string, phase internal.PhaseType, opts run.PhaseFinishOptions) (*run.Run, error)
		DownloadConfig(ctx context.Context, id string) ([]byte, error)
		Watch(context.Context, run.WatchOptions) (<-chan pubsub.Event, error)
		CreateStateVersion(ctx context.Context, opts state.CreateStateVersionOptions) (*state.Version, error)
		DownloadCurrentState(ctx context.Context, workspaceID string) ([]byte, error)
		Hostname() string
		CreateRunToken(ctx context.Context, opts run.CreateRunTokenOptions) ([]byte, error)

		internal.PutChunkService
	}

	// InProcClient is a client for in-process communication with the server.
	InProcClient struct {
		variable.VariableService
		state.StateService
		workspace.WorkspaceService
		internal.HostnameService
		configversion.ConfigurationVersionService
		run.RunService
		logs.LogsService
	}

	// rpcClient is a client for communication via RPC with the server.
	rpcClient struct {
		*otfapi.Client
		otfapi.Config

		*stateClient
		*configClient
		*variableClient
		*workspaceClient
		*runClient
		*logsClient

		// rpcClient doesn't implement all of AgentTokenService so stub it out
		// here to ensure it satisfies interface implementation.
		AgentTokenService
	}

	stateClient     = state.Client
	configClient    = configversion.Client
	variableClient  = variable.Client
	workspaceClient = workspace.Client
	runClient       = run.Client
	logsClient      = logs.Client
)

// New constructs a client that uses RPC to remotely invoke OTF services.
func newClient(config AgentConfig) (*rpcClient, error) {
	api, err := otfapi.NewClient(config.APIConfig)
	if err != nil {
		return nil, err
	}
	return &rpcClient{
		Client:          api,
		stateClient:     &stateClient{Client: api},
		configClient:    &configClient{Client: api},
		variableClient:  &variableClient{Client: api},
		workspaceClient: &workspaceClient{Client: api},
		runClient:       &runClient{Client: api, Config: config.APIConfig},
		logsClient:      &logsClient{Client: api},
	}, nil
}

func (c *rpcClient) CreateAgentToken(ctx context.Context, opts CreateAgentTokenOptions) ([]byte, error) {
	req, err := c.NewRequest("POST", "agent/create", &opts)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := c.Do(ctx, req, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *rpcClient) GetAgentToken(ctx context.Context, token string) (*AgentToken, error) {
	req, err := c.NewRequest("GET", "agent/details", nil)
	if err != nil {
		return nil, err
	}
	var at AgentToken
	if err := c.Do(ctx, req, &at); err != nil {
		return nil, err
	}
	return &at, nil
}
