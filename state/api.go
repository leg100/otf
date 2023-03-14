package state

import (
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/jsonapi"
)

type api struct {
	svc Service
}

// Implements TFC state versions API:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/state-versions#state-versions-api
func (h *api) AddHandlers(r *mux.Router) {
	r.HandleFunc("/workspaces/{workspace_id}/state-versions", h.createVersion).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/current-state-version", h.getCurrentVersion).Methods("GET")
	r.HandleFunc("/state-versions/{id}", h.getVersion).Methods("GET")
	r.HandleFunc("/state-versions", h.listVersions).Methods("GET")
	r.HandleFunc("/state-versions/{id}/download", h.downloadState).Methods("GET")

	r.HandleFunc("/state-versions/{id}/outputs", h.listOutputs).Methods("GET")
	r.HandleFunc("/state-version-outputs/{id}", h.getOutput).Methods("GET")
}

func (h *api) createVersion(w http.ResponseWriter, r *http.Request) {
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

	// TODO: validate md5, lineage

	// The docs (linked above) state the serial in the create options must match the
	// serial in the state file. However, the go-tfe integration tests we use
	// send different values for each and expect the serial in the create
	// options to take precedence, without error. We've opted to support that
	// behaviour.
	sv, err := h.svc.createVersion(r.Context(), otf.CreateStateVersionOptions{
		WorkspaceID: otf.String(workspaceID),
		State:       decoded,
		Serial:      opts.Serial,
	})
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, sv)
}

func (h *api) listVersions(w http.ResponseWriter, r *http.Request) {
	var opts stateVersionListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	svl, err := h.svc.listVersions(r.Context(), opts)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, svl)
}

func (h *api) getCurrentVersion(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	sv, err := h.svc.currentVersion(r.Context(), workspaceID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, sv)
}

func (h *api) getVersion(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.Param("id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	sv, err := h.svc.getVersion(r.Context(), versionID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, sv)
}

func (h *api) downloadState(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.Param("id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	resp, err := h.svc.downloadState(r.Context(), versionID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	w.Write(resp)
}

func (h *api) listOutputs(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.Param("id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	sv, err := h.svc.getVersion(r.Context(), versionID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, sv.Outputs)
}

func (h *api) getOutput(w http.ResponseWriter, r *http.Request) {
	outputID, err := decode.Param("id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	out, err := h.svc.getOutput(r.Context(), outputID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, out)
}
