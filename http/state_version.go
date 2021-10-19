package http

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
)

// StateVersion represents a Terraform Enterprise state version.
type StateVersion struct {
	ID           string    `jsonapi:"primary,state-versions"`
	CreatedAt    time.Time `jsonapi:"attr,created-at,iso8601"`
	DownloadURL  string    `jsonapi:"attr,hosted-state-download-url"`
	Serial       int64     `jsonapi:"attr,serial"`
	VCSCommitSHA string    `jsonapi:"attr,vcs-commit-sha"`
	VCSCommitURL string    `jsonapi:"attr,vcs-commit-url"`

	// Relations
	Run     *Run                  `jsonapi:"relation,run"`
	Outputs []*StateVersionOutput `jsonapi:"relation,outputs"`
}

// StateVersionList represents a list of state versions.
type StateVersionList struct {
	*otf.Pagination
	Items []*StateVersion
}

func (s *Server) ListStateVersions(w http.ResponseWriter, r *http.Request) {
	var opts otf.StateVersionListOptions
	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.StateVersionService.List(opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, StateVersionListJSONAPIObject(obj))
}

func (s *Server) CurrentStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := s.StateVersionService.Current(vars["workspace_id"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, StateVersionJSONAPIObject(obj))
}

func (s *Server) GetStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := s.StateVersionService.Get(vars["id"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, StateVersionJSONAPIObject(obj))
}

func (s *Server) CreateStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := otf.StateVersionCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.StateVersionService.Create(vars["workspace_id"], opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, StateVersionJSONAPIObject(obj))
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
func StateVersionJSONAPIObject(r *otf.StateVersion) *StateVersion {
	obj := &StateVersion{
		ID:          r.ID,
		CreatedAt:   r.CreatedAt,
		DownloadURL: r.DownloadURL(),
		Serial:      r.Serial,
		Outputs:     StateVersionOutputListJSONAPIObject(r.Outputs),
	}

	return obj
}

// StateVersionListJSONAPIObject converts a StateVersionList to
// a struct that can be marshalled into a JSON-API object
func StateVersionListJSONAPIObject(cvl *otf.StateVersionList) *StateVersionList {
	obj := &StateVersionList{
		Pagination: cvl.Pagination,
	}
	for _, item := range cvl.Items {
		obj.Items = append(obj.Items, StateVersionJSONAPIObject(item))
	}

	return obj
}

// StateVersionOutputJSONAPIObject converts a StateVersionOutput to a struct that can be marshalled into a
// JSON-API object
func StateVersionOutputJSONAPIObject(svo *otf.StateVersionOutput) *StateVersionOutput {
	obj := &StateVersionOutput{
		ID:        svo.ID,
		Name:      svo.Name,
		Sensitive: svo.Sensitive,
		Type:      svo.Type,
		Value:     svo.Value,
	}

	return obj
}

// StateVersionOutputListJSONAPIObject converts a StateVersionOutputList to
// a struct that can be marshalled into a JSON-API object
func StateVersionOutputListJSONAPIObject(svol otf.StateVersionOutputList) []*StateVersionOutput {
	var obj []*StateVersionOutput
	for _, item := range svol {
		obj = append(obj, StateVersionOutputJSONAPIObject(item))
	}

	return obj
}
