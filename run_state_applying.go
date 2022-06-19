package otf

import (
	"context"
)

type applyingState struct {
	run *Run
	*runStateMixin
	Job
}

func newApplyingState(r *Run) *applyingState {
	return &applyingState{
		run:           r,
		runStateMixin: &runStateMixin{},
	}
}

func (s *applyingState) Status() RunStatus { return RunApplying }

func (s *applyingState) Start() error {
	return ErrJobAlreadyClaimed
}

func (s *applyingState) Finish(svc ReportService, opts JobFinishOptions) error {
	if opts.Errored {
		s.run.setState(s.run.erroredState)
		s.run.Apply.updateStatus(JobErrored)
		return nil
	}
	report, err := svc.CreateApplyReport(context.Background(), s.run.Apply.ID())
	if err != nil {
		return err
	}
	s.run.Apply.ResourceReport = &report
	s.run.setState(s.run.appliedState)
	s.run.Apply.status = JobFinished
	return nil
}

func (s *applyingState) Cancelable() bool { return true }

func (s *applyingState) Cancel() error {
	s.run.setState(s.run.canceledState)
	return nil
}
