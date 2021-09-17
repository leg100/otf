package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

func (s *Server) GetPlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := s.PlanService.Get(vars["id"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.PlanJSONAPIObject(obj))
}

func (s *Server) GetPlanJSON(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	json, err := s.PlanService.GetPlanJSON(vars["id"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	if _, err := w.Write(json); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) GetPlanLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var opts ots.GetChunkOptions

	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	logs, err := s.RunService.GetPlanLogs(vars["id"], opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	if _, err := w.Write(logs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// PlanJSONAPIObject converts a Plan to a struct that can be
// marshalled into a JSON-API object
func (s *Server) PlanJSONAPIObject(p *ots.Plan) *tfe.Plan {
	obj := &tfe.Plan{
		ID:                   p.ID,
		HasChanges:           p.HasChanges(),
		LogReadURL:           s.GetURL(GetPlanLogsRoute, p.ID),
		ResourceAdditions:    p.ResourceAdditions,
		ResourceChanges:      p.ResourceChanges,
		ResourceDestructions: p.ResourceDestructions,
		Status:               p.Status,
		StatusTimestamps:     p.StatusTimestamps,
	}

	return obj
}
