package otf

type planQueuedState struct {
	run *Run
	*runStateMixin
	Job
}

func newPlanQueuedState(r *Run) *planQueuedState {
	return &planQueuedState{
		run:           r,
		runStateMixin: &runStateMixin{},
	}
}

func (s *planQueuedState) String() string    { return "plan_queued" }
func (s *planQueuedState) Status() RunStatus { return RunPlanQueued }
func (s *planQueuedState) Canceleable() bool { return true }

func (s *planQueuedState) Start() error {
	s.run.setState(s.run.planningState)
	s.run.Plan.updateStatus(JobRunning)
	return nil
}

func (s *planQueuedState) Cancel() error {
	s.run.setState(s.run.canceledState)
	s.run.Plan.updateStatus(JobCanceled)
	s.run.Apply.updateStatus(JobUnreachable)
	return nil
}
