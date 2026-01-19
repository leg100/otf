package runner

import (
	"context"

	"github.com/leg100/otf/internal/logr"
	"golang.org/x/sync/errgroup"
)

const ForkExecutorKind = "fork"

type forkExecutor struct {
	config                     OperationConfig
	logger                     logr.Logger
	operationClientConstructor operationClientConstructor

	n int
}

func (s *forkExecutor) SpawnOperation(ctx context.Context, g *errgroup.Group, job *Job, jobToken []byte) error {
	client, err := s.operationClientConstructor(jobToken)
	if err != nil {
		return err
	}
	s.n++
	DoOperation(ctx, g, OperationOptions{
		Logger:          s.logger,
		OperationConfig: s.config,
		Job:             job,
		JobToken:        jobToken,
		Client:          client,
	})
	s.n--
	return nil
}

func (s *forkExecutor) currentJobs() int {
	return s.n
}
