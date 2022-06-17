package otf

type plannedAndFinishedState struct {
	run *Run
	*runStateMixin
}

func newPlannedAndFinishedState(r *Run) *plannedAndFinishedState {
	return &plannedAndFinishedState{
		run: r,
		runStateMixin: &runStateMixin{
			run: r,
		},
	}
}

func (s *plannedAndFinishedState) String() string { return "planned_and_finished" }
