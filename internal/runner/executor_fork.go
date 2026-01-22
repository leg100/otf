package runner

import (
	"context"

	"github.com/leg100/otf/internal/logr"
	"golang.org/x/sync/errgroup"
)

const ForkExecutorKind = "fork"

type forkExecutor struct {
	config          OperationConfig
	logger          logr.Logger
	operationClient OperationClient

	n int
}

func (s *forkExecutor) SpawnOperation(ctx context.Context, g *errgroup.Group, job *Job, jobToken []byte) error {
	if s.operationClient.OperationClientUseToken != nil {
		s.operationClient.UseToken(string(jobToken))
	}

	s.n++
	DoOperation(ctx, g, OperationOptions{
		Logger:          s.logger,
		OperationConfig: s.config,
		Job:             job,
		JobToken:        jobToken,
		Client:          s.operationClient,
	})
	s.n--
	return nil
}

func (s *forkExecutor) currentJobs() int {
	return s.n
}
