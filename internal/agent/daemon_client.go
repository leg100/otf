package agent

import (
	"context"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/leg100/otf/internal"
	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/logs"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/workspace"
)

type (
	// daemonClient allows the daemon to communicate with the server endpoints.
	daemonClient struct {
		runs       runClient
		workspaces workspaceClient
		variables  variablesClient
		agents     agentClient
		state      stateClient
		configs    configClient
		logs       logsClient
		server     hostnameClient

		// address of OTF server peer - only populated when daemonClient is using RPC
		address string
	}

	// InProcClient is a client for in-process communication with the server.

	runClient interface {
		GetRun(ctx context.Context, runID string) (*run.Run, error)
		GetPlanFile(ctx context.Context, id string, format run.PlanFormat) ([]byte, error)
		UploadPlanFile(ctx context.Context, id string, plan []byte, format run.PlanFormat) error
		GetLockFile(ctx context.Context, id string) ([]byte, error)
		UploadLockFile(ctx context.Context, id string, lockFile []byte) error
	}

	workspaceClient interface {
		GetWorkspace(ctx context.Context, workspaceID string) (*workspace.Workspace, error)
	}

	variablesClient interface {
		ListEffectiveVariables(ctx context.Context, runID string) ([]*variable.Variable, error)
	}

	agentClient interface {
		registerAgent(ctx context.Context, opts registerAgentOptions) (*Agent, error)
		getAgentJobs(ctx context.Context, agentID string) ([]*Job, error)
		updateAgentStatus(ctx context.Context, agentID string, status AgentStatus) error

		startJob(ctx context.Context, spec JobSpec) ([]byte, error)
		finishJob(ctx context.Context, spec JobSpec, opts finishJobOptions) error
	}

	configClient interface {
		DownloadConfig(ctx context.Context, id string) ([]byte, error)
	}

	stateClient interface {
		CreateStateVersion(ctx context.Context, opts state.CreateStateVersionOptions) (*state.Version, error)
		DownloadCurrentState(ctx context.Context, workspaceID string) ([]byte, error)
	}

	logsClient interface {
		PutChunk(ctx context.Context, opts internal.PutChunkOptions) error
	}

	hostnameClient interface {
		Hostname() string
	}
)

// newRPCDaemonClient constructs a daemon client that communicates with services
// via RPC
func newRPCDaemonClient(cfg otfapi.Config, agentID *string) (*daemonClient, error) {
	apiClient, err := otfapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &daemonClient{
		runs:       &run.Client{Client: apiClient},
		workspaces: &workspace.Client{Client: apiClient},
		variables:  &variable.Client{Client: apiClient},
		agents:     &client{Client: apiClient, agentID: agentID},
		state:      &state.Client{Client: apiClient},
		configs:    &configversion.Client{Client: apiClient},
		logs:       &logs.Client{Client: apiClient},
		server:     apiClient,
		address:    cfg.Address,
	}, nil
}

// newJobClient constructs a client for communicating with services via RPC on
// behalf of a job, authenticating as a job using the job token arg.
func (c *daemonClient) newJobClient(agentID string, token []byte, logger logr.Logger) (*daemonClient, error) {
	return newRPCDaemonClient(otfapi.Config{
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
	}, &agentID)
}
