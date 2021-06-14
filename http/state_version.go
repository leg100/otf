package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/ots"
)

func (h *Server) ListStateVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var opts ots.StateVersionListOptions
	if err := DecodeAndSanitize(&opts, r.URL.Query()); err != nil {
		ErrUnprocessable(w, err)
		return
	}

	ListObjects(w, r, func() (interface{}, error) {
		return h.StateVersionService.ListStateVersions(vars["org"], opts)
	})
}

func (h *Server) CurrentStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	GetObject(w, r, func() (interface{}, error) {
		return h.StateVersionService.GetStateVersion(vars["name"], vars["org"])
	})
}

func (h *Server) GetStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	GetObject(w, r, func() (interface{}, error) {
		return h.StateVersionService.GetStateVersion(vars["name"], vars["org"])
	})
}

func (h *Server) CreateStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	CreateObject(w, r, &ots.StateVersionCreateOptions{}, func(opts interface{}) (interface{}, error) {
		return h.StateVersionService.CreateStateVersion(vars["org"], opts.(*ots.StateVersionCreateOptions))
	})
}
