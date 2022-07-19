package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

// PlanFileOptions represents the options for retrieving the plan file for a
// run.
type PlanFileOptions struct {
	// Format of plan file. Valid values are json and binary.
	Format otf.PlanFormat `schema:"format"`
}

func (s *Server) GetPlan(w http.ResponseWriter, r *http.Request) {
	planID := mux.Vars(r)["plan_id"]
	runID := otf.ConvertID(planID, "run")

	run, err := s.RunService().GetRun(r.Context(), runID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, run.Plan())
}

func (s *Server) GetPlanJSON(w http.ResponseWriter, r *http.Request) {
	planID := mux.Vars(r)["plan_id"]
	runID := otf.ConvertID(planID, "run")

	json, err := s.RunService().GetPlanFile(r.Context(), runID, otf.PlanFormatJSON)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if _, err := w.Write(json); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
