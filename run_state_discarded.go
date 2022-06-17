package otf

type discardedState struct {
	run *Run
	*runStateMixin
}

func newDiscardedState(r *Run) *discardedState {
	return &discardedState{
		run: r,
		runStateMixin: &runStateMixin{
			run: r,
		},
	}
}

func (s *discardedState) String() string { return "discarded" }

func (s *discardedState) Start() error {
	s.run.setState(s.run.discardedState)
	return nil
}

func (s *discardedState) Cancel() error {
	s.run.setState(s.run.canceledState)
	return nil
}
