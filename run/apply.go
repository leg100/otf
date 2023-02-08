package run

import (
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

// Apply is the apply phase of a run
type Apply struct {
	// ResourcesReport is a report of applied resource changes
	*ResourceReport

	runID string
	*phaseStatus
}

func (a *Apply) ID() string       { return a.runID }
func (a *Apply) Phase() PhaseType { return ApplyPhase }

// ToJSONAPI assembles a JSONAPI DTO.
func (a *apply) ToJSONAPI() any {
	dto := &jsonapi.Apply{
		ID:               otf.ConvertID(a.ID(), "apply"),
		LogReadURL:       a.signedLogURL(a.req, a.ID(), "apply"),
		Status:           string(a.Status()),
		StatusTimestamps: &jsonapi.PhaseStatusTimestamps{},
	}
	if a.ResourceReport != nil {
		dto.Additions = &a.Additions
		dto.Changes = &a.Changes
		dto.Destructions = &a.Destructions
	}
	for _, ts := range a.StatusTimestamps() {
		switch ts.Status {
		case otf.PhasePending:
			dto.StatusTimestamps.PendingAt = &ts.Timestamp
		case otf.PhaseCanceled:
			dto.StatusTimestamps.CanceledAt = &ts.Timestamp
		case otf.PhaseErrored:
			dto.StatusTimestamps.ErroredAt = &ts.Timestamp
		case otf.PhaseFinished:
			dto.StatusTimestamps.FinishedAt = &ts.Timestamp
		case otf.PhaseQueued:
			dto.StatusTimestamps.QueuedAt = &ts.Timestamp
		case otf.PhaseRunning:
			dto.StatusTimestamps.StartedAt = &ts.Timestamp
		case otf.PhaseUnreachable:
			dto.StatusTimestamps.UnreachableAt = &ts.Timestamp
		}
	}
	return dto
}

func newApply(run *Run) *Apply {
	return &Apply{
		runID:       run.id,
		phaseStatus: newPhaseStatus(),
	}
}
