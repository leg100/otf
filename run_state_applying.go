package otf

import (
	"context"
	"fmt"
)

type applyingState struct {
	run *Run
	*runStateMixin
}

func newApplyingState(r *Run) *applyingState {
	return &applyingState{
		run: r,
		runStateMixin: &runStateMixin{
			run: r,
		},
	}
}

func (s *applyingState) String() string { return "applying" }

func (s *applyingState) Finish(svc RunService) (*ResourceReport, error) {
	chunk, err := svc.GetChunk(ctx, run.Apply.JobID(), otf.GetChunkOptions{})
	if err != nil {
		return err
	}
	report, err := otf.ParseApplyOutput(string(chunk.Data))
	if err != nil {
		return fmt.Errorf("compiling report of applied changes: %w", err)
	}
	if err := s.db.CreateApplyReport(ctx, run.Apply.JobID(), report); err != nil {
		return fmt.Errorf("saving applied changes report: %w", err)
	}
	if err := svc.CreateApplyReport(context.Background(), s.run.id); err != nil {
		return err
	}
	s.run.setState(s.run.applyingState)
	return nil
}

func (s *applyingState) Cancel() error {
	s.run.setState(s.run.canceledState)
	return nil
}
