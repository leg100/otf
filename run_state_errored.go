package otf

type erroredState struct {
	run *Run
	*runStateMixin
	Job
}

func newErroredState(r *Run) *erroredState {
	return &erroredState{
		run:           r,
		runStateMixin: &runStateMixin{},
	}
}

func (s *erroredState) String() string    { return "errored" }
func (s *erroredState) Status() RunStatus { return RunErrored }
func (s *erroredState) Done() bool        { return true }
