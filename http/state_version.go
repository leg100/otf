package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/go-tfe"
	"github.com/leg100/jsonapi"
	"github.com/leg100/ots"
)

func (s *Server) ListStateVersions(w http.ResponseWriter, r *http.Request) {
	var opts tfe.StateVersionListOptions
	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.StateVersionService.List(opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.StateVersionListJSONAPIObject(obj))
}

func (s *Server) CurrentStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := s.StateVersionService.Current(vars["workspace_id"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.StateVersionJSONAPIObject(obj))
}

func (s *Server) GetStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := s.StateVersionService.Get(vars["id"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.StateVersionJSONAPIObject(obj))
}

func (s *Server) CreateStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := tfe.StateVersionCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.StateVersionService.Create(vars["workspace_id"], opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.StateVersionJSONAPIObject(obj))
}

func (s *Server) DownloadStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	resp, err := s.StateVersionService.Download(vars["id"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.Write(resp)
}

// StateVersionJSONAPIObject converts a StateVersion to a struct that can be
// marshalled into a JSON-API object
func (s *Server) StateVersionJSONAPIObject(r *ots.StateVersion) *tfe.StateVersion {
	obj := &tfe.StateVersion{
		ID:          r.ExternalID,
		CreatedAt:   r.CreatedAt,
		DownloadURL: r.DownloadURL(),
		Serial:      r.Serial,
		Outputs:     s.StateVersionOutputListJSONAPIObject(r.Outputs),
	}

	return obj
}

// StateVersionListJSONAPIObject converts a StateVersionList to
// a struct that can be marshalled into a JSON-API object
func (s *Server) StateVersionListJSONAPIObject(cvl *ots.StateVersionList) *tfe.StateVersionList {
	obj := &tfe.StateVersionList{
		Pagination: cvl.Pagination,
	}
	for _, item := range cvl.Items {
		obj.Items = append(obj.Items, s.StateVersionJSONAPIObject(item))
	}

	return obj
}

// StateVersionOutputJSONAPIObject converts a StateVersionOutput to a struct that can be marshalled into a
// JSON-API object
func (s *Server) StateVersionOutputJSONAPIObject(svo *ots.StateVersionOutput) *tfe.StateVersionOutput {
	obj := &tfe.StateVersionOutput{
		ID:        svo.ExternalID,
		Name:      svo.Name,
		Sensitive: svo.Sensitive,
		Type:      svo.Type,
		Value:     svo.Value,
	}

	return obj
}

// StateVersionOutputListJSONAPIObject converts a StateVersionOutputList to
// a struct that can be marshalled into a JSON-API object
func (s *Server) StateVersionOutputListJSONAPIObject(svol ots.StateVersionOutputList) []*tfe.StateVersionOutput {
	var obj []*tfe.StateVersionOutput
	for _, item := range svol {
		obj = append(obj, s.StateVersionOutputJSONAPIObject(item))
	}

	return obj
}
