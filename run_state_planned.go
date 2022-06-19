package otf

type plannedState struct {
	run *Run
	*runStateMixin
	Job
}

func newPlannedState(r *Run) *plannedState {
	return &plannedState{
		run:           r,
		runStateMixin: &runStateMixin{},
	}
}

func (s *plannedState) String() string    { return "planned" }
func (s *plannedState) Status() RunStatus { return RunPlanned }
func (s *plannedState) Discardable() bool { return true }
func (s *plannedState) Confirmable() bool { return true }

func (s *plannedState) ApplyRun() error {
	s.run.setState(s.run.applyQueuedState)
	s.run.Apply.updateStatus(JobQueued)
	return nil
}
