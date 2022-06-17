package otf

type planQueuedState struct {
	run *Run
	*runStateMixin
}

func newPlanEnqueuedState(r *Run) *planQueuedState {
	return &planQueuedState{
		run: r,
		runStateMixin: &runStateMixin{
			run: r,
		},
	}
}

func (s *planQueuedState) String() string { return "plan_queued" }

func (s *planQueuedState) Start() error {
	s.run.setState(s.run.planningState)
	return nil
}

func (s *planQueuedState) Cancel() error {
	s.run.setState(s.run.canceledState)
	return nil
}
