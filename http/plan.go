package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/dto"
	httputil "github.com/leg100/otf/http/util"
)

// PlanFileOptions represents the options for retrieving the plan file for a
// run.
type PlanFileOptions struct {
	// Format of plan file. Valid values are json and binary.
	Format otf.PlanFormat `schema:"format"`
}

func (s *Server) GetPlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	obj, err := s.PlanService().Get(vars["id"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, PlanDTO(r, obj))
}

func (s *Server) GetPlanJSON(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	json, err := s.RunService().GetPlanFile(r.Context(), otf.RunGetOptions{PlanID: &id}, otf.PlanFormatJSON)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
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
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	chunk, err := s.PlanService().GetChunk(r.Context(), vars["id"], opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if _, err := w.Write(chunk.Marshal()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) UploadPlanLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	var opts otf.PutChunkOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	chunk := otf.Chunk{
		Data:  buf.Bytes(),
		Start: opts.Start,
		End:   opts.End,
	}
	if err := s.PlanService().PutChunk(r.Context(), vars["id"], chunk); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
}

// PlanDTO converts a Plan to a struct that can be
// marshalled into a JSON-API object
func PlanDTO(r *http.Request, p *otf.Plan) *dto.Plan {
	result := &dto.Plan{
		ID:         p.ID(),
		HasChanges: p.HasChanges(),
		LogReadURL: httputil.Absolute(r, fmt.Sprintf(string(GetPlanLogsRoute), p.ID())),
		Status:     string(p.Status()),
	}
	if p.ResourceReport != nil {
		result.ResourceAdditions = p.Additions
		result.ResourceChanges = p.Changes
		result.ResourceDestructions = p.Destructions
	}
	for _, ts := range p.StatusTimestamps() {
		if result.StatusTimestamps == nil {
			result.StatusTimestamps = &dto.PlanStatusTimestamps{}
		}
		switch ts.Status {
		case otf.PlanCanceled:
			result.StatusTimestamps.CanceledAt = &ts.Timestamp
		case otf.PlanErrored:
			result.StatusTimestamps.ErroredAt = &ts.Timestamp
		case otf.PlanFinished:
			result.StatusTimestamps.FinishedAt = &ts.Timestamp
		case otf.PlanQueued:
			result.StatusTimestamps.QueuedAt = &ts.Timestamp
		case otf.PlanRunning:
			result.StatusTimestamps.StartedAt = &ts.Timestamp
		}
	}
	return result
}
