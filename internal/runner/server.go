package runner

import (
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/logs"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/workspace"
)

type ServerOptions struct {
	Logger                      logr.Logger
	Config                      Config
	RunService                  *run.Service
	WorkspaceService            *workspace.Service
	VariableService             *variable.Service
	ConfigurationVersionService *configversion.Service
	StateService                *state.Service
	LogsService                 *logs.Service
	runnerService               *Service
	HostnameService             *internal.HostnameService
}

// NewServerRunner constructs a runner that is part of the otfd server.
func NewServerRunner(logger logr.Logger, cfg Config, opts ServerOptions) (*runner, error) {
	return newRunner(Options{
		Logger: logger,
		Config: cfg,
		client: &client{
			runs:       opts.RunService,
			workspaces: opts.WorkspaceService,
			state:      opts.StateService,
			variables:  opts.VariableService,
			configs:    opts.ConfigurationVersionService,
			logs:       opts.LogsService,
			runners:    opts.runnerService,
			server:     opts.HostnameService,
		},
	})
}
