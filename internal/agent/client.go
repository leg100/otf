package agent

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/leg100/otf/internal"
	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/logs"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/workspace"
)

const agentIDHeader = "otf-agent-id"

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
		GetWorkspace(ctx context.Context, workspaceID string) (*workspace.Workspace, error)
		Hostname() string

		internal.PutChunkService

		registerAgent(ctx context.Context, opts registerAgentOptions) (*Agent, error)
		getAgentJobs(ctx context.Context, agentID string) ([]*Job, error)
		updateAgentStatus(ctx context.Context, agentID string, status AgentStatus) error
		startJob(ctx context.Context, spec JobSpec) ([]byte, error)
		finishJob(ctx context.Context, spec JobSpec, opts finishJobOptions) error
	}

	// InProcClient is a client for in-process communication with the server.
	InProcClient struct {
		tokens.TokensService
		variable.VariableService
		state.StateService
		internal.HostnameService
		configversion.ConfigurationVersionService
		run.RunService
		workspace.WorkspaceService
		logs.LogsService
		Service
	}

	// rpcClient is a client for communication via RPC with the server.
	rpcClient struct {
		*otfapi.Client

		*stateClient
		*configClient
		*variableClient
		*runClient
		*logsClient
		*workspaceClient

		// agentID is the ID of the agent using the client
		agentID *string
		// address of OTF server peer
		address string
	}

	stateClient     = state.Client
	configClient    = configversion.Client
	variableClient  = variable.Client
	workspaceClient = workspace.Client
	runClient       = run.Client
	logsClient      = logs.Client
)

// NewRPCClient constructs a client that uses RPC to call OTF services. The
// agentID is added as an HTTP header to requests, allowing the server to
// identify the client; if nil then it is automatically added once the client
// successfully registers the agent.
func NewRPCClient(cfg otfapi.Config, agentID *string) (*rpcClient, error) {
	client, err := otfapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &rpcClient{
		Client:          client,
		address:         cfg.Address,
		stateClient:     &stateClient{Client: client},
		configClient:    &configClient{Client: client},
		variableClient:  &variableClient{Client: client},
		runClient:       &runClient{Client: client, Config: cfg},
		workspaceClient: &workspaceClient{Client: client},
		logsClient:      &logsClient{Client: client},
	}, nil
}

func (c *rpcClient) NewJobClient(token []byte, logger logr.Logger) (*rpcClient, error) {
	return NewRPCClient(otfapi.Config{
		Address:       c.address,
		Token:         string(token),
		RetryRequests: true,
		RetryLogHook: func(_ retryablehttp.Logger, r *http.Request, n int) {
			// ignore first un-retried requests
			if n == 0 {
				return
			}
			logger.Error(nil, "retrying request", "url", r.URL, "attempt", n)
		},
	}, c.agentID)
}

// NewRequest constructs a new API request
func (c *rpcClient) NewRequest(method, path string, v interface{}) (*retryablehttp.Request, error) {
	req, err := c.Client.NewRequest(method, path, v)
	if err != nil {
		return nil, err
	}
	if c.agentID != nil {
		req.Header.Add(agentIDHeader, *c.agentID)
	}
	return req, err
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
	// add agent ID to future requests
	agentID := agent.ID
	c.agentID = &agentID
	return &agent, nil
}

func (c *rpcClient) getAgentJobs(ctx context.Context, agentID string) ([]*Job, error) {
	req, err := c.NewRequest("GET", "agents/jobs", nil)
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
	if err := c.Do(ctx, req, &jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (c *rpcClient) updateAgentStatus(ctx context.Context, agentID string, status AgentStatus) error {
	req, err := c.NewRequest("POST", "agents/status", &updateAgentStatusParams{
		Status: status,
	})
	if err != nil {
		return err
	}
	if err := c.Do(ctx, req, nil); err != nil {
		return err
	}
	return nil
}

// agent tokens

func (c *rpcClient) CreateAgentToken(ctx context.Context, poolID string, opts CreateAgentTokenOptions) (*agentToken, []byte, error) {
	u := fmt.Sprintf("agent-tokens/%s/create", poolID)
	req, err := c.NewRequest("POST", u, &opts)
	if err != nil {
		return nil, nil, err
	}
	var buf bytes.Buffer
	if err := c.Do(ctx, req, &buf); err != nil {
		return nil, nil, err
	}
	return nil, buf.Bytes(), nil
}

// jobs

func (c *rpcClient) startJob(ctx context.Context, spec JobSpec) ([]byte, error) {
	req, err := c.NewRequest("POST", "agents/start", &spec)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := c.Do(ctx, req, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *rpcClient) finishJob(ctx context.Context, spec JobSpec, opts finishJobOptions) error {
	req, err := c.NewRequest("POST", "agents/finish", &finishJobParams{
		JobSpec:          spec,
		finishJobOptions: opts,
	})
	if err != nil {
		return err
	}
	if err := c.Do(ctx, req, nil); err != nil {
		return err
	}
	return nil
}
