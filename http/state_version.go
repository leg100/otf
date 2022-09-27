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

func (s *Server) CreateStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	opts := dto.StateVersionCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	sv, err := s.Application.CreateStateVersion(r.Context(), vars["workspace_id"], otf.StateVersionCreateOptions{
		Lineage: opts.Lineage,
		Serial:  opts.Serial,
		State:   opts.State,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &StateVersion{sv})
}

func (s *Server) ListStateVersions(w http.ResponseWriter, r *http.Request) {
	var opts otf.StateVersionListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	svl, err := s.Application.ListStateVersions(r.Context(), opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &StateVersionList{svl})
}

func (s *Server) CurrentStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sv, err := s.Application.CurrentStateVersion(r.Context(), vars["workspace_id"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &StateVersion{sv})
}

func (s *Server) GetStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sv, err := s.Application.GetStateVersion(r.Context(), vars["id"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &StateVersion{sv})
}

func (s *Server) DownloadStateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resp, err := s.DownloadState(r.Context(), vars["id"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.Write(resp)
}

type StateVersion struct {
	*otf.StateVersion
}

// ToJSONAPI assembles a JSON-API DTO.
func (sv *StateVersion) ToJSONAPI() any {
	obj := &dto.StateVersion{
		ID:          sv.ID(),
		CreatedAt:   sv.CreatedAt(),
		DownloadURL: fmt.Sprintf("/api/v2/state-versions/%s/download", sv.ID()),
		Serial:      sv.Serial(),
	}
	for _, out := range sv.Outputs() {
		obj.Outputs = append(obj.Outputs, &dto.StateVersionOutput{
			ID:        out.ID(),
			Name:      out.Name,
			Sensitive: out.Sensitive,
			Type:      out.Type,
			Value:     out.Value,
		})
	}
	return obj
}

type StateVersionList struct {
	*otf.StateVersionList
}

// ToJSONAPI assembles a JSON-API DTO.
func (l *StateVersionList) ToJSONAPI() any {
	obj := &dto.StateVersionList{
		Pagination: l.Pagination.ToJSONAPI(),
	}
	for _, item := range l.Items {
		obj.Items = append(obj.Items, (&StateVersion{item}).ToJSONAPI().(*dto.StateVersion))
	}
	return obj
}
