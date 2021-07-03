package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) GetApply(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	GetObject(w, r, func() (interface{}, error) {
		return s.ApplyService.GetApply(vars["id"])
	})
}
