package otf

type appliedState struct {
	run *Run
	*runStateMixin
}

func newAppliedState(r *Run) *appliedState {
	return &appliedState{
		run: r,
		runStateMixin: &runStateMixin{
			run: r,
		},
	}
}

func (s *appliedState) String() string { return "applied" }
func (s *appliedState) Done() bool     { return true }
