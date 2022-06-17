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
	logs, err := svc.GetApplyLogs(context.Background(), s.run.Apply.JobID())
	if err != nil {
		return nil, err
	}
	report, err := ParseApplyOutput(string(logs))
	if err != nil {
		return nil, fmt.Errorf("compiling report of applied changes: %w", err)
	}

	s.run.setState(s.run.applyingState)
	return &report, nil
}

func (s *applyingState) Cancel() error {
	s.run.setState(s.run.canceledState)
	return nil
}
