package agent

import (
	"context"

	otfrun "github.com/leg100/otf/internal/run"
)

type service struct {
	*db
}

func (s *service) createPool(ctx context.Context, run *otfrun.Run, agentID string) (*Job, error) {
	return newJob(run, agentID), nil
}

func (s *service) createJob(ctx context.Context, run *otfrun.Run, agentID string) (*Job, error) {
	return newJob(run, agentID), nil
}

func (s *service) ping(ctx context.Context, agentID string) error {
	return nil
}
