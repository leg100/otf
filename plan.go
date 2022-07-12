package otf

import (
	"fmt"
	"net/http"

	jsonapi "github.com/leg100/otf/http/dto"
	httputil "github.com/leg100/otf/http/util"
)

const (
	LocalStateFilename = "terraform.tfstate"
	PlanFilename       = "plan.out"
	JSONPlanFilename   = "plan.out.json"
)

// Plan is the plan phase of a run
type Plan struct {
	// report of planned resource changes
	*ResourceReport

	runID string
	*phaseStatus
}

func (p *Plan) ID() string       { return p.runID }
func (p *Plan) Phase() PhaseType { return PlanPhase }

// HasChanges determines whether plan has any changes (adds/changes/deletions).
func (p *Plan) HasChanges() bool {
	return p.ResourceReport != nil && p.ResourceReport.HasChanges()
}

// ToJSONAPI assembles a JSON-API DTO.
func (p *Plan) ToJSONAPI(req *http.Request) any {
	dto := &jsonapi.Plan{
		ID:               ConvertID(p.runID, "plan"),
		HasChanges:       p.HasChanges(),
		LogReadURL:       httputil.Absolute(req, fmt.Sprintf("runs/%s/logs/plan", p.runID)),
		Status:           string(p.Status()),
		StatusTimestamps: &jsonapi.PhaseStatusTimestamps{},
	}
	if p.ResourceReport != nil {
		dto.Additions = &p.Additions
		dto.Changes = &p.Changes
		dto.Destructions = &p.Destructions
	}
	for _, ts := range p.StatusTimestamps() {
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

func newPlan(run *Run) *Plan {
	return &Plan{
		runID:       run.id,
		phaseStatus: newPhaseStatus(),
	}
}
