package otf

type canceledState struct {
	*runStateMixin
	Job
}

func newCanceledState(r *Run) *canceledState {
	return &canceledState{
		runStateMixin: &runStateMixin{},
	}
}

func (s *canceledState) String() string    { return "canceled" }
func (s *canceledState) Status() RunStatus { return RunCanceled }
func (s *canceledState) Done() bool        { return true }
