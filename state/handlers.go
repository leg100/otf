package state

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/dto"
)

type Server struct {
	otf.Application
}

func (s *Server) CreateStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := versionJSONAPICreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		dto.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	// convert from json-api to domain
	sv, err := s.Application.CreateStateVersion(r.Context(), vars["workspace_id"], otf.StateVersionCreateOptions{
		Lineage: opts.Lineage,
		Serial:  opts.Serial,
		State:   opts.State,
	})
	if err != nil {
		dto.Error(w, http.StatusNotFound, err)
		return
	}
	dto.WriteResponse(w, r, sv)
}

func (s *Server) ListStateVersions(w http.ResponseWriter, r *http.Request) {
	var opts otf.StateVersionListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		dto.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	svl, err := s.Application.ListStateVersions(r.Context(), opts)
	if err != nil {
		dto.Error(w, http.StatusNotFound, err)
		return
	}
	dto.WriteResponse(w, r, &StateVersionList{svl})
}

func (s *Server) CurrentStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sv, err := s.Application.CurrentStateVersion(r.Context(), vars["workspace_id"])
	if err != nil {
		dto.Error(w, http.StatusNotFound, err)
		return
	}
	dto.WriteResponse(w, r, &StateVersion{sv})
}

func (s *Server) GetStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sv, err := s.Application.GetStateVersion(r.Context(), vars["id"])
	if err != nil {
		dto.Error(w, http.StatusNotFound, err)
		return
	}
	dto.WriteResponse(w, r, &StateVersion{sv})
}

func (s *Server) DownloadStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resp, err := s.DownloadState(r.Context(), vars["id"])
	if err != nil {
		dto.Error(w, http.StatusNotFound, err)
		return
	}
	w.Write(resp)
}
