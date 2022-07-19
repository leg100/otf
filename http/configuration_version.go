package http

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/dto"
)

func (s *Server) CreateConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := dto.ConfigurationVersionCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	cv, err := s.ConfigurationVersionService().CreateConfigurationVersion(r.Context(), vars["workspace_id"], otf.ConfigurationVersionCreateOptions{
		AutoQueueRuns: opts.AutoQueueRuns,
		Speculative:   opts.Speculative,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, cv, withCode(http.StatusCreated))
}

func (s *Server) GetConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cv, err := s.ConfigurationVersionService().GetConfigurationVersion(r.Context(), vars["id"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, cv)
}

func (s *Server) ListConfigurationVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var opts otf.ConfigurationVersionListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	cvl, err := s.ConfigurationVersionService().ListConfigurationVersion(r.Context(), vars["workspace_id"], opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, cvl)
}

func (s *Server) UploadConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := s.ConfigurationVersionService().UploadConfig(r.Context(), vars["id"], buf.Bytes()); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
}
