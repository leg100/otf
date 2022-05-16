package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	httputil "github.com/leg100/otf/http/util"
)

// Plan represents a Terraform Enterprise plan.
type Plan struct {
	ID                   string                `jsonapi:"primary,plans"`
	HasChanges           bool                  `jsonapi:"attr,has-changes"`
	LogReadURL           string                `jsonapi:"attr,log-read-url"`
	ResourceAdditions    int                   `jsonapi:"attr,resource-additions"`
	ResourceChanges      int                   `jsonapi:"attr,resource-changes"`
	ResourceDestructions int                   `jsonapi:"attr,resource-destructions"`
	Status               otf.PlanStatus        `jsonapi:"attr,status"`
	StatusTimestamps     *PlanStatusTimestamps `jsonapi:"attr,status-timestamps"`
}

// PlanStatusTimestamps holds the timestamps for individual plan statuses.
type PlanStatusTimestamps struct {
	CanceledAt      *time.Time `json:"canceled-at,omitempty"`
	ErroredAt       *time.Time `json:"errored-at,omitempty"`
	FinishedAt      *time.Time `json:"finished-at,omitempty"`
	ForceCanceledAt *time.Time `json:"force-canceled-at,omitempty"`
	QueuedAt        *time.Time `json:"queued-at,omitempty"`
	StartedAt       *time.Time `json:"started-at,omitempty"`
}

// PlanFileOptions represents the options for retrieving the plan file for a
// run.
type PlanFileOptions struct {
	// Format of plan file. Valid values are json and binary.
	Format otf.PlanFormat `schema:"format"`
}

// ToDomain converts http organization obj to a domain organization obj.
func (p *Plan) ToDomain() *otf.Plan {
	return &otf.Plan{
		ID: p.ID,
		ResourceReport: &otf.ResourceReport{
			Additions:    p.ResourceAdditions,
			Changes:      p.ResourceChanges,
			Destructions: p.ResourceDestructions,
		},
		Status: p.Status,
	}
}

func (s *Server) GetPlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := s.PlanService().Get(vars["id"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, PlanJSONAPIObject(r, obj))
}

func (s *Server) GetPlanJSON(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	json, err := s.RunService().GetPlanFile(r.Context(), otf.RunGetOptions{PlanID: &id}, otf.PlanFormatJSON)
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

	chunk, err := s.PlanService().GetChunk(r.Context(), vars["id"], opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
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
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	var opts otf.PutChunkOptions

	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	chunk := otf.Chunk{
		Data:  buf.Bytes(),
		Start: opts.Start,
		End:   opts.End,
	}

	if err := s.PlanService().PutChunk(r.Context(), vars["id"], chunk); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}
}

// PlanJSONAPIObject converts a Plan to a struct that can be
// marshalled into a JSON-API object
func PlanJSONAPIObject(r *http.Request, p *otf.Plan) *Plan {
	result := &Plan{
		ID:         p.ID,
		HasChanges: p.HasChanges(),
		LogReadURL: httputil.Absolute(r, fmt.Sprintf(string(GetPlanLogsRoute), p.ID)),
		Status:     p.Status,
	}
	if p.ResourceReport != nil {
		result.ResourceAdditions = p.Additions
		result.ResourceChanges = p.Changes
		result.ResourceDestructions = p.Destructions
	}

	for _, ts := range p.StatusTimestamps {
		if result.StatusTimestamps == nil {
			result.StatusTimestamps = &PlanStatusTimestamps{}
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
