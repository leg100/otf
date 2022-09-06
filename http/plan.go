package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
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
	writeResponse(w, r, run.Plan())
}

// getPlanJSON retrieves a plan object's plan file in JSON format.
//
// https://www.terraform.io/cloud-docs/api-docs/plans#retrieve-the-json-execution-plan
func (s *Server) getPlanJSON(w http.ResponseWriter, r *http.Request) {
	// otf's plan IDs are simply the corresponding run ID
	planID := mux.Vars(r)["plan_id"]
	runID := otf.ConvertID(planID, "run")

	json, err := s.Application.GetPlanFile(r.Context(), runID, otf.PlanFormatJSON)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if _, err := w.Write(json); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
