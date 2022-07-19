package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

func (s *Server) GetApply(w http.ResponseWriter, r *http.Request) {
	applyID := mux.Vars(r)["apply_id"]
	runID := otf.ConvertID(applyID, "run")

	run, err := s.RunService().GetRun(r.Context(), runID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, run.Apply())
}
