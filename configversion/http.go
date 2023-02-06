package http

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/jsonapi"
)

func (s *Server) CreateConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := jsonapi.ConfigurationVersionCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	cv, err := s.Application.CreateConfigurationVersion(r.Context(), vars["workspace_id"], otf.ConfigurationVersionCreateOptions{
		AutoQueueRuns: opts.AutoQueueRuns,
		Speculative:   opts.Speculative,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &ConfigurationVersion{cv, s}, withCode(http.StatusCreated))
}

func (s *Server) GetConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cv, err := s.Application.GetConfigurationVersion(r.Context(), vars["id"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &ConfigurationVersion{cv, s})
}

func (s *Server) ListConfigurationVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var opts otf.ConfigurationVersionListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	cvl, err := s.Application.ListConfigurationVersions(r.Context(), vars["workspace_id"], opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &ConfigurationVersionList{cvl, s})
}

func (s *Server) UploadConfigurationVersion() http.HandlerFunc {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, r.Body); err != nil {
			writeError(w, http.StatusUnprocessableEntity, err)
			return
		}
		if err := s.UploadConfig(r.Context(), vars["id"], buf.Bytes()); err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
	})
	return http.MaxBytesHandler(h, s.MaxConfigSize).ServeHTTP
}

func (s *Server) DownloadConfigurationVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resp, err := s.DownloadConfig(r.Context(), vars["id"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.Write(resp)
}

type ConfigurationVersion struct {
	*otf.ConfigurationVersion
	*Server
}

// ToJSONAPI assembles a JSONAPI DTO.
func (cv *ConfigurationVersion) ToJSONAPI() any {
	obj := &jsonapi.ConfigurationVersion{
		ID:               cv.ID(),
		AutoQueueRuns:    cv.AutoQueueRuns(),
		Speculative:      cv.Speculative(),
		Source:           string(cv.Source()),
		Status:           string(cv.Status()),
		StatusTimestamps: &jsonapi.CVStatusTimestamps{},
		UploadURL:        cv.signedUploadURL(cv.ID()),
	}
	for _, ts := range cv.StatusTimestamps() {
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

type ConfigurationVersionList struct {
	*otf.ConfigurationVersionList
	*Server
}

// ToJSONAPI assembles a JSONAPI DTO
func (l *ConfigurationVersionList) ToJSONAPI() any {
	obj := &jsonapi.ConfigurationVersionList{
		Pagination: l.Pagination.ToJSONAPI(),
	}
	for _, item := range l.Items {
		obj.Items = append(obj.Items, (&ConfigurationVersion{item, l.Server}).ToJSONAPI().(*jsonapi.ConfigurationVersion))
	}
	return obj
}