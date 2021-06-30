package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/go-tfe"
)

func (h *Server) ListConfigurationVersions(w http.ResponseWriter, r *http.Request) {
	var opts tfe.ConfigurationVersionListOptions
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		ErrUnprocessable(w, fmt.Errorf("unable to decode query string: %w", err))
		return
	}

	SanitizeListOptions(&opts.ListOptions)

	ListObjects(w, r, func() (interface{}, error) {
		return h.ConfigurationVersionService.ListConfigurationVersions(opts)
	})
}

func (h *Server) GetConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	GetObject(w, r, func() (interface{}, error) {
		return h.ConfigurationVersionService.GetConfigurationVersion(vars["name"])
	})
}

func (h *Server) CreateConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	CreateObject(w, r, &tfe.ConfigurationVersionCreateOptions{}, func(opts interface{}) (interface{}, error) {
		return h.ConfigurationVersionService.CreateConfigurationVersion(opts.(*tfe.ConfigurationVersionCreateOptions))
	})
}

func (h *Server) UploadConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		ErrUnprocessable(w, err)
		return
	}

	if err := h.ConfigurationVersionService.UploadConfigurationVersion(vars["id"], buf.Bytes()); err != nil {
		ErrNotFound(w)
		return
	}
}
