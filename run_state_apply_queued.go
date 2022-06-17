package otf

type applyQueuedState struct {
	run *Run
	*runStateMixin
}

func newApplyQueuedState(r *Run) *applyQueuedState {
	return &applyQueuedState{
		run: r,
		runStateMixin: &runStateMixin{
			run: r,
		},
	}
}

func (s *applyQueuedState) String() string { return "apply_queued" }

func (s *applyQueuedState) Start() error {
	s.run.setState(s.run.applyingState)
	return nil
}

func (s *applyQueuedState) Cancel() error {
	s.run.setState(s.run.canceledState)
	return nil
}
