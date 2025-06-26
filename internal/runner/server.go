package runner

import (
	"context"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
)

// ServerRunnerOptions are options for constructing a server runner.
type ServerRunnerOptions struct {
	*Config

	Logger     logr.Logger
	Runners    *Service
	Runs       runClient
	Workspaces workspaceClient
	Variables  variablesClient
	State      stateClient
	Configs    configClient
	Server     hostnameClient
	Jobs       operationJobsClient
}

// NewServerRunner constructs a server runner.
func NewServerRunner(opts ServerRunnerOptions) (*Runner, error) {
	daemon, err := newRunner(
		opts.Logger,
		opts.Runners,
		&localOperationSpawner{
			logger:     opts.Logger,
			config:     *opts.Config,
			runs:       opts.Runs,
			workspaces: opts.Workspaces,
			variables:  opts.Variables,
			state:      opts.State,
			configs:    opts.Configs,
			server:     opts.Server,
			jobs:       opts.Jobs,
		},
		false,
		*opts.Config,
	)
	if err != nil {
		return nil, err
	}
	return daemon, nil
}

type localOperationSpawner struct {
	config     Config
	logger     logr.Logger
	runs       runClient
	workspaces workspaceClient
	variables  variablesClient
	state      stateClient
	configs    configClient
	server     hostnameClient
	jobs       operationJobsClient
}

func (s *localOperationSpawner) NewOperation(ctx context.Context, jobID resource.TfeID, jobToken []byte) (*operation, error) {
	return newOperation(ctx, operationOptions{
		logger:          s.logger,
		OperationConfig: s.config.OperationConfig,
		jobID:           jobID,
		jobToken:        jobToken,
		jobs:            s.jobs,
		runs:            s.runs,
		workspaces:      s.workspaces,
		variables:       s.variables,
		state:           s.state,
		configs:         s.configs,
		server:          s.server,
	})
}
