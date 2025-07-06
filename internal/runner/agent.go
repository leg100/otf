package runner

import (
	"context"

	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/workspace"
	"golang.org/x/sync/errgroup"
)

type AgentOptions struct {
	*Config

	URL   string
	Token string
}

func NewAgent(logger logr.Logger, opts AgentOptions) (*Runner, error) {
	apiClient, err := otfapi.NewClient(otfapi.Config{
		URL:           opts.URL,
		Token:         opts.Token,
		Logger:        logger,
		RetryRequests: true,
	})
	if err != nil {
		return nil, err
	}
	opts.OperationConfig.IsAgent = true
	return newRunner(
		logger,
		&remoteClient{Client: apiClient},
		&RemoteOperationSpawner{
			Logger: logger,
			Config: opts.OperationConfig,
			URL:    opts.URL,
		},
		true,
		*opts.Config,
	)
}

type RemoteOperationSpawner struct {
	Config OperationConfig
	Logger logr.Logger
	URL    string
}

func (s *RemoteOperationSpawner) SpawnOperation(ctx context.Context, g *errgroup.Group, job *Job, jobToken []byte) error {
	client, err := otfapi.NewClient(otfapi.Config{
		URL:           s.URL,
		Token:         string(jobToken),
		Logger:        s.Logger,
		RetryRequests: true,
	})
	if err != nil {
		return err
	}
	doOperation(ctx, g, operationOptions{
		logger:          s.Logger,
		OperationConfig: s.Config,
		job:             job,
		jobToken:        jobToken,
		runs:            &run.Client{Client: client},
		jobs:            &remoteClient{Client: client},
		workspaces:      &workspace.Client{Client: client},
		variables:       &variable.Client{Client: client},
		state:           &state.Client{Client: client},
		configs:         &configversion.Client{Client: client},
		server:          client,
	})
	return nil
}
