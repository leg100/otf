package otf

type plannedAndFinishedState struct {
	run *Run
	*runStateMixin
	Job
}

func newPlannedAndFinishedState(r *Run) *plannedAndFinishedState {
	return &plannedAndFinishedState{
		run:           r,
		runStateMixin: &runStateMixin{},
	}
}

func (s *plannedAndFinishedState) String() string    { return "planned_and_finished" }
func (s *plannedAndFinishedState) Status() RunStatus { return RunPlannedAndFinished }
func (s *plannedAndFinishedState) Done() bool        { return true }
