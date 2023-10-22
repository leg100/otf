package agent

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
	_ client = (*LocalClient)(nil)
	_ client = (*remoteClient)(nil)
)

type (
	// client allows the agent to communicate with the server endpoints.
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

	// LocalClient is the client for an internal agent.
	LocalClient struct {
		tokens.TokensService
		variable.VariableService
		state.StateService
		workspace.WorkspaceService
		internal.HostnameService
		configversion.ConfigurationVersionService
		run.RunService
		logs.LogsService
	}

	// remoteClient is the client for an external agent.
	remoteClient struct {
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

// New constructs a client that uses http to remotely invoke OTF
// services.
func newClient(config ExternalConfig) (*remoteClient, error) {
	api, err := otfapi.NewClient(config.APIConfig)
	if err != nil {
		return nil, err
	}

	return &remoteClient{
		Client:          api,
		stateClient:     &stateClient{JSONAPIClient: api},
		configClient:    &configClient{JSONAPIClient: api},
		variableClient:  &variableClient{JSONAPIClient: api},
		tokensClient:    &tokensClient{JSONAPIClient: api},
		workspaceClient: &workspaceClient{JSONAPIClient: api},
		runClient:       &runClient{JSONAPIClient: api, Config: config.APIConfig},
		logsClient:      &logsClient{JSONAPIClient: api},
	}, nil
}
