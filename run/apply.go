package run

// Apply is the apply phase of a run
type Apply struct {
	// ResourcesReport is a report of applied resource changes
	*ResourceReport

	runID string
	*phaseStatus
}

func (a *Apply) ID() string       { return a.runID }
func (a *Apply) Phase() PhaseType { return ApplyPhase }

func newApply(run *Run) *Apply {
	return &Apply{
		runID:       run.id,
		phaseStatus: newPhaseStatus(),
	}
}
