package otf

import (
	"fmt"
	"net/http"

	jsonapi "github.com/leg100/otf/http/dto"
	httputil "github.com/leg100/otf/http/util"
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
func (a *Apply) ToJSONAPI(req *http.Request) any {
	dto := &jsonapi.Apply{
		ID:               ConvertID(a.runID, "apply"),
		LogReadURL:       httputil.Absolute(req, fmt.Sprintf("runs/%s/logs/apply", a.runID)),
		Status:           string(a.Status()),
		StatusTimestamps: &jsonapi.PhaseStatusTimestamps{},
	}
	if a.ResourceReport != nil {
		dto.ResourceAdditions = a.Additions
		dto.ResourceChanges = a.Changes
		dto.ResourceDestructions = a.Destructions
	}
	for _, ts := range a.StatusTimestamps() {
		switch ts.Status {
		case PhasePending:
			dto.StatusTimestamps.PendingAt = &ts.Timestamp
		case PhaseCanceled:
			dto.StatusTimestamps.CanceledAt = &ts.Timestamp
		case PhaseErrored:
			dto.StatusTimestamps.ErroredAt = &ts.Timestamp
		case PhaseFinished:
			dto.StatusTimestamps.FinishedAt = &ts.Timestamp
		case PhaseQueued:
			dto.StatusTimestamps.QueuedAt = &ts.Timestamp
		case PhaseRunning:
			dto.StatusTimestamps.StartedAt = &ts.Timestamp
		case PhaseUnreachable:
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
