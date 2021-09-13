package http

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

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

	var opts ots.GetLogOptions

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

func (s *Server) UploadApplyLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	var opts ots.AppendLogOptions

	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := s.RunService.UploadApplyLogs(vars["id"], buf.Bytes(), opts); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}
}

// ApplyJSONAPIObject converts a Apply to a struct that can be marshalled into a
// JSON-API object
func (s *Server) ApplyJSONAPIObject(a *ots.Apply) *tfe.Apply {
	obj := &tfe.Apply{
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
