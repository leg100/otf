package otf

type pendingState struct {
	run *Run
	*runStateMixin
	Job
}

func newPendingState(r *Run) *pendingState {
	return &pendingState{
		run: r,
		runStateMixin: &runStateMixin{
			jobStatus: JobPending,
		},
	}
}

func (s *pendingState) Status() RunStatus { return RunPending }

func (s *pendingState) Enqueue() error {
	s.run.setState(s.run.planQueuedState)
	s.run.Plan.updateStatus(JobQueued)
	return nil
}

func (s *pendingState) Cancelable() bool { return true }

func (s *pendingState) Cancel() error {
	s.run.setState(s.run.canceledState)
	s.run.Plan.updateStatus(JobUnreachable)
	s.run.Apply.updateStatus(JobUnreachable)
	return nil
}
