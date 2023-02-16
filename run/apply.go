package run

import (
	"github.com/leg100/otf"
)

// Apply is the apply phase of a run
type Apply struct {
	// ResourcesReport is a report of applied resource changes
	*ResourceReport

	runID string
	*phaseStatus
}

func (a *Apply) ID() string           { return a.runID }
func (a *Apply) Phase() otf.PhaseType { return otf.ApplyPhase }

func newApply(run *Run) *Apply {
	return &Apply{
		runID:       run.id,
		phaseStatus: newPhaseStatus(),
	}
}
