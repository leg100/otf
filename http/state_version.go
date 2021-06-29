package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

func (h *Server) ListStateVersions(w http.ResponseWriter, r *http.Request) {
	var opts ots.StateVersionListOptions
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		ErrUnprocessable(w, fmt.Errorf("unable to decode query string: %w", err))
		return
	}

	SanitizeListOptions(&opts.ListOptions)

	ListObjects(w, r, func() (interface{}, error) {
		return h.StateVersionService.ListStateVersions(*opts.Organization, *opts.Workspace, opts)
	})
}

func (h *Server) CurrentStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	GetObject(w, r, func() (interface{}, error) {
		return h.StateVersionService.CurrentStateVersion(vars["workspace_id"])
	})
}

func (h *Server) GetStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	GetObject(w, r, func() (interface{}, error) {
		return h.StateVersionService.GetStateVersion(vars["id"])
	})
}

func (h *Server) CreateStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	CreateObject(w, r, &tfe.StateVersionCreateOptions{}, func(opts interface{}) (interface{}, error) {
		return h.StateVersionService.CreateStateVersion(vars["workspace_id"], opts.(*tfe.StateVersionCreateOptions))
	})
}

func (h *Server) DownloadStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	resp, err := h.StateVersionService.DownloadStateVersion(vars["id"])
	if err != nil {
		ErrNotFound(w)
		return
	}

	w.Write(resp)
}
