package otf

type appliedState struct {
	run *Run
	*runStateMixin
	Job
}

func newAppliedState(r *Run) *appliedState {
	return &appliedState{
		run:           r,
		runStateMixin: &runStateMixin{},
	}
}

func (s *appliedState) Status() RunStatus { return RunApplied }
func (s *appliedState) Done() bool        { return true }
