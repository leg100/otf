package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) GetApply(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	apply, err := s.ApplyService().Get(r.Context(), vars["id"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, apply)
}

func (s *Server) GetApplyLogs(w http.ResponseWriter, r *http.Request) {
	getLogs(w, r, s.ApplyService(), mux.Vars(r)["apply_id"])
}

func (s *Server) UploadApplyLogs(w http.ResponseWriter, r *http.Request) {
	uploadLogs(w, r, s.ApplyService(), mux.Vars(r)["apply_id"])
}
