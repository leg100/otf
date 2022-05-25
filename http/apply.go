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

func (s *Server) GetApply(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	obj, err := s.ApplyService().Get(vars["id"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, ApplyDTO(r, obj))
}

func (s *Server) GetApplyLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var opts otf.GetChunkOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	chunk, err := s.ApplyService().GetChunk(r.Context(), vars["id"], opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
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
	if err := s.ApplyService().PutChunk(r.Context(), vars["id"], chunk); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
}

// ApplyDTO converts an apply into a DTO
func ApplyDTO(req *http.Request, a *otf.Apply) *dto.Apply {
	o := &dto.Apply{
		ID:         a.ID(),
		LogReadURL: httputil.Absolute(req, fmt.Sprintf(string(GetApplyLogsRoute), a.ID())),
		Status:     string(a.Status()),
	}
	if a.ResourceReport != nil {
		o.ResourceAdditions = a.Additions
		o.ResourceChanges = a.Changes
		o.ResourceDestructions = a.Destructions
	}
	for _, ts := range a.StatusTimestamps() {
		if o.StatusTimestamps == nil {
			o.StatusTimestamps = &dto.ApplyStatusTimestamps{}
		}
		switch ts.Status {
		case otf.ApplyCanceled:
			o.StatusTimestamps.CanceledAt = &ts.Timestamp
		case otf.ApplyErrored:
			o.StatusTimestamps.ErroredAt = &ts.Timestamp
		case otf.ApplyFinished:
			o.StatusTimestamps.FinishedAt = &ts.Timestamp
		case otf.ApplyQueued:
			o.StatusTimestamps.QueuedAt = &ts.Timestamp
		case otf.ApplyRunning:
			o.StatusTimestamps.StartedAt = &ts.Timestamp
		}
	}
	return o
}
