package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	jsonapi "github.com/leg100/otf/http/dto"
)

// These endpoints implement the documented plan API:
//
// https://www.terraform.io/cloud-docs/api-docs/plans#retrieve-the-json-execution-plan
//

// getPlan retrieves a plan object in JSON-API format.
//
// https://www.terraform.io/cloud-docs/api-docs/plans#show-a-plan
//
func (s *Server) getPlan(w http.ResponseWriter, r *http.Request) {
	// otf's plan IDs are simply the corresponding run ID
	planID := mux.Vars(r)["plan_id"]
	runID := otf.ConvertID(planID, "run")

	run, err := s.Application.GetRun(r.Context(), runID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &plan{run.Plan(), r})
}

// getPlanJSON retrieves a plan object's plan file in JSON format.
//
// https://www.terraform.io/cloud-docs/api-docs/plans#retrieve-the-json-execution-plan
func (s *Server) getPlanJSON(w http.ResponseWriter, r *http.Request) {
	// otf's plan IDs are simply the corresponding run ID
	planID := mux.Vars(r)["plan_id"]
	runID := otf.ConvertID(planID, "run")

	json, err := s.GetPlanFile(r.Context(), runID, otf.PlanFormatJSON)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if _, err := w.Write(json); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type plan struct {
	*otf.Plan
	req *http.Request
}

// ToJSONAPI assembles a JSON-API DTO.
func (p *plan) ToJSONAPI() any {
	dto := &jsonapi.Plan{
		ID:               otf.ConvertID(p.ID(), "plan"),
		HasChanges:       p.HasChanges(),
		LogReadURL:       otf.Absolute(p.req, fmt.Sprintf("api/v2/runs/%s/logs/plan", p.ID())),
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
