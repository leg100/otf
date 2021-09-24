package http

import (
	"bytes"
	"io"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
)

// ConfigurationVersion is a representation of an uploaded or ingressed
// Terraform configuration in TFE. A workspace must have at least one
// configuration version before any runs may be queued on it.
type ConfigurationVersion struct {
	ID               string                  `jsonapi:"primary,configuration-versions"`
	AutoQueueRuns    bool                    `jsonapi:"attr,auto-queue-runs"`
	Error            string                  `jsonapi:"attr,error"`
	ErrorMessage     string                  `jsonapi:"attr,error-message"`
	Source           otf.ConfigurationSource `jsonapi:"attr,source"`
	Speculative      bool                    `jsonapi:"attr,speculative "`
	Status           otf.ConfigurationStatus `jsonapi:"attr,status"`
	StatusTimestamps *otf.CVStatusTimestamps `jsonapi:"attr,status-timestamps"`
	UploadURL        string                  `jsonapi:"attr,upload-url"`
}

// ConfigurationVersionList represents a list of configuration versions.
type ConfigurationVersionList struct {
	*otf.Pagination
	Items []*ConfigurationVersion
}

func (s *Server) CreateConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := otf.ConfigurationVersionCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.ConfigurationVersionService.Create(vars["workspace_id"], opts)
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

	var opts otf.ConfigurationVersionListOptions
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
func (s *Server) ConfigurationVersionJSONAPIObject(cv *otf.ConfigurationVersion) *ConfigurationVersion {
	obj := &ConfigurationVersion{
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
func (s *Server) ConfigurationVersionListJSONAPIObject(cvl *otf.ConfigurationVersionList) *ConfigurationVersionList {
	obj := &ConfigurationVersionList{
		Pagination: cvl.Pagination,
	}
	for _, item := range cvl.Items {
		obj.Items = append(obj.Items, s.ConfigurationVersionJSONAPIObject(item))
	}

	return obj
}
