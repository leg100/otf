package otf

type applyQueuedState struct {
	run *Run
	*runStateMixin
	Job
}

func newApplyQueuedState(r *Run) *applyQueuedState {
	return &applyQueuedState{
		run:           r,
		runStateMixin: &runStateMixin{},
	}
}

func (s *applyQueuedState) Status() RunStatus { return RunApplyQueued }
func (s *applyQueuedState) Canceleable() bool { return true }

func (s *applyQueuedState) Start() error {
	s.run.setState(s.run.applyingState)
	return nil
}

func (s *applyQueuedState) Cancel() error {
	s.run.setState(s.run.canceledState)
	return nil
}
