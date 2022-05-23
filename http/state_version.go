package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/dto"
)

func (s *Server) ListStateVersions(w http.ResponseWriter, r *http.Request) {
	var opts otf.StateVersionListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	obj, err := s.StateVersionService().List(opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, StateVersionListJSONAPIObject(obj))
}

func (s *Server) CurrentStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	obj, err := s.StateVersionService().Current(vars["workspace_id"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, StateVersionJSONAPIObject(obj))
}

func (s *Server) GetStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	obj, err := s.StateVersionService().Get(vars["id"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, StateVersionJSONAPIObject(obj))
}

func (s *Server) CreateStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := otf.StateVersionCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	obj, err := s.StateVersionService().Create(vars["workspace_id"], opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, StateVersionJSONAPIObject(obj))
}

func (s *Server) DownloadStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resp, err := s.StateVersionService().Download(vars["id"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.Write(resp)
}

// StateVersionJSONAPIObject converts a StateVersion to a struct that can be
// marshalled into a JSON-API object
func StateVersionJSONAPIObject(r *otf.StateVersion) *dto.StateVersion {
	return &dto.StateVersion{
		ID:          r.ID(),
		CreatedAt:   r.CreatedAt,
		DownloadURL: fmt.Sprintf("/state-versions/%s/download", r.ID()),
		Serial:      r.Serial,
		Outputs:     StateVersionOutputListJSONAPIObject(r.Outputs),
	}
}

// StateVersionListJSONAPIObject converts a StateVersionList to
// a struct that can be marshalled into a JSON-API object
func StateVersionListJSONAPIObject(l *otf.StateVersionList) *dto.StateVersionList {
	pagination := dto.Pagination(*l.Pagination)
	obj := &dto.StateVersionList{
		Pagination: &pagination,
	}
	for _, item := range l.Items {
		obj.Items = append(obj.Items, StateVersionJSONAPIObject(item))
	}
	return obj
}

// StateVersionOutputJSONAPIObject converts a StateVersionOutput to a struct that can be marshalled into a
// JSON-API object
func StateVersionOutputJSONAPIObject(svo *otf.StateVersionOutput) *dto.StateVersionOutput {
	return &dto.StateVersionOutput{
		ID:        svo.ID(),
		Name:      svo.Name,
		Sensitive: svo.Sensitive,
		Type:      svo.Type,
		Value:     svo.Value,
	}
}

// StateVersionOutputListJSONAPIObject converts a StateVersionOutputList to
// a struct that can be marshalled into a JSON-API object
func StateVersionOutputListJSONAPIObject(svol otf.StateVersionOutputList) []*dto.StateVersionOutput {
	var obj []*dto.StateVersionOutput
	for _, item := range svol {
		obj = append(obj, StateVersionOutputJSONAPIObject(item))
	}
	return obj
}
