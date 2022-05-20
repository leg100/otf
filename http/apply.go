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

// Apply represents a Terraform Enterprise apply.
type Apply struct {
	ID                   string                 `jsonapi:"primary,applies"`
	LogReadURL           string                 `jsonapi:"attr,log-read-url"`
	ResourceAdditions    int                    `jsonapi:"attr,resource-additions"`
	ResourceChanges      int                    `jsonapi:"attr,resource-changes"`
	ResourceDestructions int                    `jsonapi:"attr,resource-destructions"`
	Status               otf.ApplyStatus        `jsonapi:"attr,status"`
	StatusTimestamps     *ApplyStatusTimestamps `jsonapi:"attr,status-timestamps"`
}

// ApplyStatusTimestamps holds the timestamps for individual apply statuses.
type ApplyStatusTimestamps struct {
	CanceledAt      *time.Time `json:"canceled-at,omitempty"`
	ErroredAt       *time.Time `json:"errored-at,omitempty"`
	FinishedAt      *time.Time `json:"finished-at,omitempty"`
	ForceCanceledAt *time.Time `json:"force-canceled-at,omitempty"`
	QueuedAt        *time.Time `json:"queued-at,omitempty"`
	StartedAt       *time.Time `json:"started-at,omitempty"`
}

func (s *Server) GetApply(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := s.ApplyService().Get(vars["id"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, ApplyJSONAPIObject(r, obj))
}

func (s *Server) GetApplyLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var opts otf.GetChunkOptions

	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	chunk, err := s.ApplyService().GetChunk(r.Context(), vars["id"], opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	if _, err := w.Write(chunk.Marshal()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) UploadApplyLogs(w http.ResponseWriter, r *http.Request) {
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

	if err := s.ApplyService().PutChunk(r.Context(), vars["id"], chunk); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}
}

// ApplyJSONAPIObject converts a Apply to a struct that can be marshalled into a
// JSON-API object
func ApplyJSONAPIObject(req *http.Request, a *otf.Apply) *Apply {
	obj := &Apply{
		ID:         a.ID,
		LogReadURL: httputil.Absolute(req, fmt.Sprintf(string(GetApplyLogsRoute), a.ID)),
		Status:     a.Status(),
	}
	if a.ResourceReport != nil {
		obj.ResourceAdditions = a.Additions
		obj.ResourceChanges = a.Changes
		obj.ResourceDestructions = a.Destructions
	}

	for _, ts := range a.StatusTimestamps() {
		if obj.StatusTimestamps == nil {
			obj.StatusTimestamps = &ApplyStatusTimestamps{}
		}
		switch ts.Status {
		case otf.ApplyCanceled:
			obj.StatusTimestamps.CanceledAt = &ts.Timestamp
		case otf.ApplyErrored:
			obj.StatusTimestamps.ErroredAt = &ts.Timestamp
		case otf.ApplyFinished:
			obj.StatusTimestamps.FinishedAt = &ts.Timestamp
		case otf.ApplyQueued:
			obj.StatusTimestamps.QueuedAt = &ts.Timestamp
		case otf.ApplyRunning:
			obj.StatusTimestamps.StartedAt = &ts.Timestamp
		}
	}

	return obj
}
