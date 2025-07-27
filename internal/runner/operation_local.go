package runner

import (
	"context"

	"github.com/leg100/otf/internal/logr"
	"golang.org/x/sync/errgroup"
)

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

	n int
}

func (s *localOperationSpawner) SpawnOperation(ctx context.Context, g *errgroup.Group, job *Job, jobToken []byte) error {
	s.n++
	doOperation(ctx, g, operationOptions{
		logger:          s.logger,
		OperationConfig: s.config.OperationConfig,
		job:             job,
		jobToken:        jobToken,
		jobs:            s.jobs,
		runs:            s.runs,
		workspaces:      s.workspaces,
		variables:       s.variables,
		state:           s.state,
		configs:         s.configs,
		server:          s.server,
	})
	s.n--
	return nil
}

func (s *localOperationSpawner) currentJobs() int {
	return s.n
}
