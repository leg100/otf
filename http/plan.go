package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/ots"
)

func (s *Server) GetPlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	GetObject(w, r, func() (interface{}, error) {
		return s.PlanService.GetPlan(vars["id"])
	})
}

func (s *Server) GetPlanLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var opts ots.PlanLogOptions
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		ErrUnprocessable(w, fmt.Errorf("unable to decode query string: %w", err))
		return
	}

	logs, err := s.PlanService.GetPlanLogs(vars["id"], opts)
	if err != nil {
		ErrNotFound(w)
		return
	}

	if _, err := w.Write(logs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) UploadPlanLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		ErrUnprocessable(w, err)
		return
	}

	if err := s.PlanService.UploadPlanLogs(vars["id"], buf.Bytes()); err != nil {
		ErrNotFound(w)
		return
	}
}
