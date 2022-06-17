package otf

type plannedState struct {
	run *Run
	*runStateMixin
}

func newPlannedState(r *Run) *plannedState {
	return &plannedState{
		run: r,
		runStateMixin: &runStateMixin{
			run: r,
		},
	}
}

func (s *plannedState) String() string { return "planned" }
func (s *plannedState) Done() bool     { return true }
