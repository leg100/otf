package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/dto"
)

func (s *Server) CreateConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := otf.ConfigurationVersionCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.ConfigurationVersionService().Create(vars["workspace_id"], opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, ConfigurationVersionDTO(obj), WithCode(http.StatusCreated))
}

func (s *Server) GetConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := s.ConfigurationVersionService().Get(vars["id"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, ConfigurationVersionDTO(obj))
}

func (s *Server) ListConfigurationVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var opts otf.ConfigurationVersionListOptions
	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.ConfigurationVersionService().List(vars["workspace_id"], opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.ConfigurationVersionListJSONAPIObject(obj))
}

func (s *Server) UploadConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := s.ConfigurationVersionService().Upload(vars["id"], buf.Bytes()); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}
}

// ConfigurationVersionDTO converts a cv into a DTO
func ConfigurationVersionDTO(cv *otf.ConfigurationVersion) *dto.ConfigurationVersion {
	obj := &dto.ConfigurationVersion{
		ID:            cv.ID,
		AutoQueueRuns: cv.AutoQueueRuns,
		Speculative:   cv.Speculative,
		Source:        string(cv.Source),
		Status:        string(cv.Status),
		UploadURL:     fmt.Sprintf(string(UploadConfigurationVersionRoute), cv.ID),
	}

	for _, ts := range cv.StatusTimestamps {
		if obj.StatusTimestamps == nil {
			obj.StatusTimestamps = &dto.CVStatusTimestamps{}
		}
		switch ts.Status {
		case otf.ConfigurationPending:
			obj.StatusTimestamps.QueuedAt = &ts.Timestamp
		case otf.ConfigurationErrored:
			obj.StatusTimestamps.FinishedAt = &ts.Timestamp
		case otf.ConfigurationUploaded:
			obj.StatusTimestamps.StartedAt = &ts.Timestamp
		}
	}

	return obj
}

// ConfigurationVersionListJSONAPIObject converts a ConfigurationVersionList to
// a struct that can be marshalled into a JSON-API object
func (s *Server) ConfigurationVersionListJSONAPIObject(cvl *otf.ConfigurationVersionList) *dto.ConfigurationVersionList {
	pagination := dto.Pagination(*cvl.Pagination)
	obj := &dto.ConfigurationVersionList{
		Pagination: &pagination,
	}
	for _, item := range cvl.Items {
		obj.Items = append(obj.Items, ConfigurationVersionDTO(item))
	}

	return obj
}
