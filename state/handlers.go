package state

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/jsonapi"
)

type handlers struct {
	app appService
}

// Implements TFC state versions API:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/state-versions#state-versions-api
//
func (h *handlers) AddHandlers(r *mux.Router) {
	r.HandleFunc("/workspaces/{workspace_id}/state-versions", h.createVersion)
	r.HandleFunc("/workspaces/{workspace_id}/current-state-version", h.getCurrentVersion)
	r.HandleFunc("/state-versions/{id}", h.getVersion)
	r.HandleFunc("/state-versions", h.listVersions)
	r.HandleFunc("/state-versions/{id}/download", h.downloadState)
}

func (h *handlers) createVersion(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	opts := jsonapiCreateVersionOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	// required options
	if opts.Serial == nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, errors.New("missing serial number"))
		return
	}
	if opts.MD5 == nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, errors.New("missing md5"))
		return
	}

	// base64-decode state to []byte
	decoded, err := base64.StdEncoding.DecodeString(*opts.State)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	// The docs (linked above) state the serial in the create options must match the
	// serial in the state file. However, the go-tfe integration tests we use
	// send different values for each and expect the serial in the create
	// options to take precedence, without error. We've opted to support that
	// behaviour and therefore we need to update the state file with whatever serial
	// is sent before forwarding it onto the app.
	var state State
	if err := json.Unmarshal(decoded, &state); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	if *opts.Serial != state.Serial {
		state.Serial = *opts.Serial
		decoded, err = json.Marshal(state)
		if err != nil {
			jsonapi.Error(w, http.StatusUnprocessableEntity, err)
			return
		}
	}

	// TODO: validate md5, lineage

	sv, err := h.app.createVersion(r.Context(), otf.CreateStateVersionOptions{
		WorkspaceID: otf.String(workspaceID),
		State:       decoded,
	})
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, sv)
}

func (s *handlers) listVersions(w http.ResponseWriter, r *http.Request) {
	var opts StateVersionListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	svl, err := s.app.listVersions(r.Context(), opts)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, svl)
}

func (s *handlers) getCurrentVersion(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	sv, err := s.app.currentVersion(r.Context(), workspaceID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, sv)
}

func (s *handlers) getVersion(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.Param("id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	sv, err := s.app.getVersion(r.Context(), versionID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, sv)
}

func (s *handlers) downloadState(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.Param("id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	resp, err := s.app.downloadState(r.Context(), versionID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	w.Write(resp)
}
