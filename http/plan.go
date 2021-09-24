package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

// Plan represents a Terraform Enterprise plan.
type Plan struct {
	ID                   string                    `jsonapi:"primary,plans"`
	HasChanges           bool                      `jsonapi:"attr,has-changes"`
	LogReadURL           string                    `jsonapi:"attr,log-read-url"`
	ResourceAdditions    int                       `jsonapi:"attr,resource-additions"`
	ResourceChanges      int                       `jsonapi:"attr,resource-changes"`
	ResourceDestructions int                       `jsonapi:"attr,resource-destructions"`
	Status               otf.PlanStatus            `jsonapi:"attr,status"`
	StatusTimestamps     *otf.PlanStatusTimestamps `jsonapi:"attr,status-timestamps"`
}

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

	var opts otf.GetChunkOptions

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
func (s *Server) PlanJSONAPIObject(p *otf.Plan) *Plan {
	obj := &Plan{
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
