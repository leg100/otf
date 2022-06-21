package otf

type erroredState struct {
	run *Run
	*runStateMixin
	Job
}

func newErroredState(r *Run) *erroredState {
	return &erroredState{
		run: r,
		runStateMixin: &runStateMixin{
			final: true,
		},
	}
}

func (s *erroredState) Status() RunStatus { return RunErrored }
func (s *erroredState) Done() bool        { return true }
