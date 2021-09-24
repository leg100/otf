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

// StateVersionCreateOptions represents the options for creating a state
// version.
type StateVersionCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,state-versions"`

	// The lineage of the state.
	Lineage *string `jsonapi:"attr,lineage,omitempty"`

	// The MD5 hash of the state version.
	MD5 *string `jsonapi:"attr,md5"`

	// The serial of the state.
	Serial *int64 `jsonapi:"attr,serial"`

	// The base64 encoded state.
	State *string `jsonapi:"attr,state"`

	// Force can be set to skip certain validations. Wrong use
	// of this flag can cause data loss, so USE WITH CAUTION!
	Force *bool `jsonapi:"attr,force"`

	// Specifies the run to associate the state with.
	Run *Run `jsonapi:"relation,run,omitempty"`
}

func (o *StateVersionCreateOptions) ToDomain() otf.StateVersionCreateOptions {
	domain := otf.StateVersionCreateOptions{
		Lineage: o.Lineage,
		MD5:     o.MD5,
		Serial:  o.Serial,
		State:   o.State,
		Force:   o.Force,
	}

	if o.Run != nil {
		domain.Run = o.Run.ToDomain()
	}

	return domain
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

	opts := StateVersionCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.StateVersionService.Create(vars["workspace_id"], opts.ToDomain())
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
func (s *Server) StateVersionJSONAPIObject(r *otf.StateVersion) *StateVersion {
	obj := &StateVersion{
		ID:          r.ID,
		CreatedAt:   r.Model.CreatedAt,
		DownloadURL: r.DownloadURL(),
		Serial:      r.Serial,
		Outputs:     s.StateVersionOutputListJSONAPIObject(r.Outputs),
	}

	return obj
}

// StateVersionListJSONAPIObject converts a StateVersionList to
// a struct that can be marshalled into a JSON-API object
func (s *Server) StateVersionListJSONAPIObject(cvl *otf.StateVersionList) *StateVersionList {
	obj := &StateVersionList{
		Pagination: cvl.Pagination,
	}
	for _, item := range cvl.Items {
		obj.Items = append(obj.Items, s.StateVersionJSONAPIObject(item))
	}

	return obj
}

// StateVersionOutputJSONAPIObject converts a StateVersionOutput to a struct that can be marshalled into a
// JSON-API object
func (s *Server) StateVersionOutputJSONAPIObject(svo *otf.StateVersionOutput) *StateVersionOutput {
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
func (s *Server) StateVersionOutputListJSONAPIObject(svol otf.StateVersionOutputList) []*StateVersionOutput {
	var obj []*StateVersionOutput
	for _, item := range svol {
		obj = append(obj, s.StateVersionOutputJSONAPIObject(item))
	}

	return obj
}
