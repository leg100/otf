package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/go-tfe"
	"github.com/leg100/jsonapi"
)

func (s *Server) ListStateVersions(w http.ResponseWriter, r *http.Request) {
	var opts tfe.StateVersionListOptions
	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.StateVersionService.ListStateVersions(*opts.Organization, *opts.Workspace, opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, obj)
}

func (s *Server) CurrentStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := s.StateVersionService.CurrentStateVersion(vars["workspace_id"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, obj)
}

func (s *Server) GetStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := s.StateVersionService.GetStateVersion(vars["id"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, obj)
}

func (s *Server) CreateStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := tfe.StateVersionCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.StateVersionService.CreateStateVersion(vars["workspace_id"], &opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, obj)
}

func (s *Server) DownloadStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	resp, err := s.StateVersionService.DownloadStateVersion(vars["id"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.Write(resp)
}
