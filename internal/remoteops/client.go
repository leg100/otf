package remoteops

import (
	"context"

	"github.com/leg100/otf/internal"
	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/logs"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/tokens"
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

		tokens.RunTokenService
		internal.PutChunkService
	}

	// InProcClient is a client for in-process communication with the server.
	InProcClient struct {
		tokens.TokensService
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
		*tokensClient
		*workspaceClient
		*runClient
		*logsClient
	}

	stateClient     = state.Client
	configClient    = configversion.Client
	variableClient  = variable.Client
	tokensClient    = tokens.Client
	workspaceClient = workspace.Client
	runClient       = run.Client
	logsClient      = logs.Client
)

// New constructs a client that uses RPC to remotely invoke OTF services.
func newClient(cfg otfapi.Config) (*rpcClient, error) {
	api, err := otfapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &rpcClient{
		Client:          api,
		stateClient:     &stateClient{Client: api},
		configClient:    &configClient{Client: api},
		variableClient:  &variableClient{Client: api},
		tokensClient:    &tokensClient{Client: api},
		workspaceClient: &workspaceClient{Client: api},
		runClient:       &runClient{Client: api, Config: cfg},
		logsClient:      &logsClient{Client: api},
	}, nil
}
