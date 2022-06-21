package otf

type discardedState struct {
	run *Run
	*runStateMixin
	Job
}

func newDiscardedState(r *Run) *discardedState {
	return &discardedState{
		run: r,
		runStateMixin: &runStateMixin{
			final:     true,
			jobStatus: JobFinished,
		},
	}
}

func (s *discardedState) Status() RunStatus { return RunDiscarded }
func (s *discardedState) Done() bool        { return true }
