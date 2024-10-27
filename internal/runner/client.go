package runner

import (
	"context"
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
		register(ctx context.Context, opts registerRunnerOptions) (*Runner, error)
		getRunnerJobs(ctx context.Context, agentID string) ([]*Job, error)
		updateRunnerStatus(ctx context.Context, agentID string, status RunnerStatus) error

		startJob(ctx context.Context, spec JobSpec) ([]byte, error)
		finishJob(ctx context.Context, spec JobSpec, opts finishJobOptions) error
	}
)
