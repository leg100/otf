package runner

import (
	"context"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
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
	// client allows the runner to communicate with the server endpoints.
	client struct {
		runs       runClient
		workspaces workspaceClient
		variables  variablesClient
		runners    runnerClient
		state      stateClient
		configs    configClient
		logs       logsClient
		server     hostnameClient

		// URL of OTF server peer - only populated when daemonClient is using RPC
		url string
	}

	runnerClient interface {
		registerRunner(ctx context.Context, opts registerRunnerOptions) (*Runner, error)
		getRunnerJobs(ctx context.Context, agentID string) ([]*Job, error)
		updateRunnerStatus(ctx context.Context, agentID string, status RunnerStatus) error

		startJob(ctx context.Context, spec JobSpec) ([]byte, error)
		finishJob(ctx context.Context, spec JobSpec, opts finishJobOptions) error
	}
)

// newRPCClient constructs a daemon client that communicates with services
// via RPC
func newRPCClient(cfg otfapi.Config, agentID *string) (*client, error) {
	apiClient, err := otfapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &client{
		runs:       &run.Client{Client: apiClient},
		workspaces: &workspace.Client{Client: apiClient},
		variables:  &variable.Client{Client: apiClient},
		runners:    &client{Client: apiClient, agentID: agentID},
		state:      &state.Client{Client: apiClient},
		configs:    &configversion.Client{Client: apiClient},
		logs:       &logs.Client{Client: apiClient},
		server:     apiClient,
		url:        cfg.URL,
	}, nil
}

// newJobClient constructs a client for communicating with services via RPC on
// behalf of a job, authenticating as a job using the job token arg.
func (c *client) newJobClient(agentID string, token []byte, logger logr.Logger) (*client, error) {
	return newRPCClient(otfapi.Config{
		URL:           c.url,
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
