package agent

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal"
	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/logs"
	"github.com/leg100/otf/internal/pubsub"
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
		GetRun(ctx context.Context, runID string) (*run.Run, error)
		ListEffectiveVariables(ctx context.Context, runID string) ([]*variable.Variable, error)
		GetPlanFile(ctx context.Context, id string, format run.PlanFormat) ([]byte, error)
		UploadPlanFile(ctx context.Context, id string, plan []byte, format run.PlanFormat) error
		GetLockFile(ctx context.Context, id string) ([]byte, error)
		UploadLockFile(ctx context.Context, id string, lockFile []byte) error
		DownloadConfig(ctx context.Context, id string) ([]byte, error)
		Watch(context.Context, run.WatchOptions) (<-chan pubsub.Event, error)
		CreateStateVersion(ctx context.Context, opts state.CreateStateVersionOptions) (*state.Version, error)
		DownloadCurrentState(ctx context.Context, workspaceID string) ([]byte, error)
		Hostname() string
		GetAgentToken(context.Context, string) (*tokens.AgentToken, error)

		tokens.RunTokenService
		internal.PutChunkService

		registerAgent(ctx context.Context, opts registerAgentOptions) (*Agent, error)
		getAgentJobs(ctx context.Context, agentID string) ([]*Job, error)
	}

	// InProcClient is a client for in-process communication with the server.
	InProcClient struct {
		tokens.TokensService
		variable.VariableService
		state.StateService
		internal.HostnameService
		configversion.ConfigurationVersionService
		run.RunService
		logs.LogsService
		Service
	}

	// rpcClient is a client for communication via RPC with the server.
	rpcClient struct {
		*otfapi.Client
		otfapi.Config

		*stateClient
		*configClient
		*variableClient
		*tokensClient
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

// NewClient constructs a client that uses RPC to call OTF services.
func NewClient(cfg otfapi.Config) (*rpcClient, error) {
	api, err := otfapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &rpcClient{
		Client:         api,
		stateClient:    &stateClient{Client: api},
		configClient:   &configClient{Client: api},
		variableClient: &variableClient{Client: api},
		tokensClient:   &tokensClient{Client: api},
		runClient:      &runClient{Client: api, Config: cfg},
		logsClient:     &logsClient{Client: api},
	}, nil
}

func (c *rpcClient) registerAgent(ctx context.Context, opts registerAgentOptions) (*Agent, error) {
	req, err := c.NewRequest("POST", "agents/register", &opts)
	if err != nil {
		return nil, err
	}
	var agent Agent
	if err := c.Do(ctx, req, &agent); err != nil {
		return nil, err
	}
	return &agent, nil
}
func (c *rpcClient) getAgentJobs(ctx context.Context, agentID string) ([]*Job, error) {
	u := fmt.Sprintf("agent/%s/jobs", agentID)
	req, err := c.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	var jobs []*Job
	// GET request blocks until:
	//
	// (a) job(s) are allocated to agent
	// (b) job(s) already allocated to agent are sent a cancelation signal
	// (c) a timeout is reached
	//
	// (c) can occur due to any intermediate proxies placed between otf-agent
	// and otfd, such as nginx, which has a default proxy_read_timeout of 60s.
	if err := c.Do(ctx, req, jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}
