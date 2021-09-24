package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

// Apply represents a Terraform Enterprise apply.
type Apply struct {
	ID                   string                     `jsonapi:"primary,applies"`
	LogReadURL           string                     `jsonapi:"attr,log-read-url"`
	ResourceAdditions    int                        `jsonapi:"attr,resource-additions"`
	ResourceChanges      int                        `jsonapi:"attr,resource-changes"`
	ResourceDestructions int                        `jsonapi:"attr,resource-destructions"`
	Status               otf.ApplyStatus            `jsonapi:"attr,status"`
	StatusTimestamps     *otf.ApplyStatusTimestamps `jsonapi:"attr,status-timestamps"`
}

func (s *Server) GetApply(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := s.ApplyService.Get(vars["id"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.ApplyJSONAPIObject(obj))
}

func (s *Server) GetApplyLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var opts otf.GetChunkOptions

	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	logs, err := s.RunService.GetApplyLogs(vars["id"], opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	if _, err := w.Write(logs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ApplyJSONAPIObject converts a Apply to a struct that can be marshalled into a
// JSON-API object
func (s *Server) ApplyJSONAPIObject(a *otf.Apply) *Apply {
	obj := &Apply{
		ID:                   a.ID,
		LogReadURL:           s.GetURL(GetApplyLogsRoute, a.ID),
		ResourceAdditions:    a.ResourceAdditions,
		ResourceChanges:      a.ResourceChanges,
		ResourceDestructions: a.ResourceDestructions,
		Status:               a.Status,
		StatusTimestamps:     a.StatusTimestamps,
	}

	return obj
}
