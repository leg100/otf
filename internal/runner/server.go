package runner

import (
	"github.com/leg100/otf/internal/logr"
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
	daemon, err := New(
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
