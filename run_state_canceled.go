package otf

type canceledState struct {
	run *Run
	*runStateMixin
}

func newCanceledState(r *Run) *canceledState {
	return &canceledState{
		run: r,
		runStateMixin: &runStateMixin{
			run: r,
		},
	}
}

func (s *canceledState) String() string { return "canceled" }
func (s *canceledState) Done() bool     { return true }
