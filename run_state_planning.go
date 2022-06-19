package otf

import "context"

type planningState struct {
	run *Run
	*runStateMixin
	Job
}

func newPlanningState(r *Run) *planningState {
	return &planningState{
		run:           r,
		runStateMixin: &runStateMixin{},
	}
}

func (s *planningState) Status() RunStatus { return RunPlanning }
func (s *planningState) Canceleable() bool { return true }

// Do performs a terraform plan
func (s *planningState) Do(env Environment) error {
	return s.run.Plan.Do(env)
}

// Start returns a job already started error
func (s *plannedState) Start() error {
	return ErrJobAlreadyClaimed
}

// Finish finishes a plan job
func (s *planningState) Finish(svc ReportService, opts JobFinishOptions) error {
	if opts.Errored {
		s.run.setState(s.run.erroredState)
		s.run.Plan.updateStatus(JobErrored)
		s.run.Apply.updateStatus(JobUnreachable)
		return nil
	}
	report, err := svc.CreatePlanReport(context.Background(), s.run.Plan.id)
	if err != nil {
		return err
	}
	s.run.Plan.ResourceReport = &report

	// we set planned state even if we transition immediately to
	// planned-and-finished or apply-queued because doing so records a timestamp
	// and we depend on that timestamp to determine whether the plan phase has
	// completed.
	s.run.setState(s.run.plannedState)

	if !s.run.HasChanges() || s.run.Speculative() {
		s.run.setState(s.run.plannedAndFinishedState)
	} else if s.run.autoApply {
		s.run.setState(s.run.applyQueuedState)
	}
	return nil
}

func (s *planningState) Cancelable() bool { return true }

func (s *planningState) Cancel() error {
	s.run.setState(s.run.canceledState)
	s.run.Plan.updateStatus(JobCanceled)
	s.run.Apply.updateStatus(JobUnreachable)
	return nil
}
