package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/dto"
)

func (s *Server) CreateStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := dto.StateVersionCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	sv, err := s.StateVersionService().CreateStateVersion(r.Context(), vars["workspace_id"], otf.StateVersionCreateOptions{
		Lineage: opts.Lineage,
		Serial:  opts.Serial,
		State:   opts.State,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, sv)
}

func (s *Server) ListStateVersions(w http.ResponseWriter, r *http.Request) {
	var opts otf.StateVersionListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	svl, err := s.StateVersionService().ListStateVersion(r.Context(), opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, svl)
}

func (s *Server) CurrentStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sv, err := s.StateVersionService().CurrentStateVersion(r.Context(), vars["workspace_id"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, sv)
}

func (s *Server) GetStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sv, err := s.StateVersionService().GetStateVersion(r.Context(), vars["id"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, sv)
}

func (s *Server) DownloadStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resp, err := s.StateVersionService().DownloadState(r.Context(), vars["id"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.Write(resp)
}
