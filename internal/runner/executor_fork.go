package runner

import (
	"context"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"golang.org/x/sync/errgroup"
)

const ForkExecutorKind = "fork"

type forkExecutor struct {
	config                 OperationConfig
	logger                 logr.Logger
	operationClientCreator OperationClientCreator

	n int
}

func (s *forkExecutor) SpawnOperation(ctx context.Context, g *errgroup.Group, job *Job, jobToken []byte) error {
	s.n++
	DoOperation(ctx, g, OperationOptions{
		Logger:          s.logger,
		OperationConfig: s.config,
		Job:             job,
		JobToken:        jobToken,
		Client:          s.operationClientCreator(string(jobToken)),
	})
	s.n--
	return nil
}

func (s *forkExecutor) currentJobs(_ context.Context, _ resource.TfeID) int {
	return s.n
}
