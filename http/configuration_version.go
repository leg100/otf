package http

import (
	"bytes"
	"io"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/leg100/go-tfe"
	"github.com/leg100/jsonapi"
	"github.com/leg100/ots"
)

func (s *Server) CreateConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := tfe.ConfigurationVersionCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.ConfigurationVersionService.Create(vars["workspace_id"], &opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.ConfigurationVersionJSONAPIObject(obj), WithCode(http.StatusCreated))
}

func (s *Server) GetConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := s.ConfigurationVersionService.Get(vars["id"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.ConfigurationVersionJSONAPIObject(obj))
}

func (s *Server) ListConfigurationVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var opts tfe.ConfigurationVersionListOptions
	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.ConfigurationVersionService.List(vars["workspace_id"], opts)
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

	if err := s.ConfigurationVersionService.Upload(vars["id"], buf.Bytes()); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}
}

// ConfigurationVersionJSONAPIObject converts a ConfigurationVersion to a struct
// that can be marshalled into a JSON-API object
func (s *Server) ConfigurationVersionJSONAPIObject(cv *ots.ConfigurationVersion) *tfe.ConfigurationVersion {
	obj := &tfe.ConfigurationVersion{
		ID:            cv.ID,
		AutoQueueRuns: cv.AutoQueueRuns,
		Error:         cv.Error,
		ErrorMessage:  cv.ErrorMessage,
		Speculative:   cv.Speculative,
		Source:        cv.Source,
		Status:        cv.Status,
		UploadURL:     s.GetURL(UploadConfigurationVersionRoute, cv.ID),
	}

	if cv.StatusTimestamps != nil && !reflect.ValueOf(cv.StatusTimestamps).Elem().IsZero() {
		obj.StatusTimestamps = cv.StatusTimestamps
	}

	return obj
}

// ConfigurationVersionListJSONAPIObject converts a ConfigurationVersionList to
// a struct that can be marshalled into a JSON-API object
func (s *Server) ConfigurationVersionListJSONAPIObject(cvl *ots.ConfigurationVersionList) *tfe.ConfigurationVersionList {
	obj := &tfe.ConfigurationVersionList{
		Pagination: cvl.Pagination,
	}
	for _, item := range cvl.Items {
		obj.Items = append(obj.Items, s.ConfigurationVersionJSONAPIObject(item))
	}

	return obj
}
