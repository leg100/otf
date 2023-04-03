package state

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/jsonapi"
)

type api struct {
	svc Service

	*jsonapiMarshaler
}

// Implements TFC state versions API:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/state-versions#state-versions-api
func (h *api) addHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

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
		jsonapi.Error(w, err)
		return
	}

	opts := jsonapi.StateVersionCreateVersionOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, err)
		return
	}

	// required options
	if opts.Serial == nil {
		jsonapi.Error(w, fmt.Errorf("%w: %s", otf.ErrMissingParameter, "serial number"))
		return
	}
	if opts.MD5 == nil {
		jsonapi.Error(w, fmt.Errorf("%w: %s", otf.ErrMissingParameter, "md5"))
		return
	}

	// base64-decode state to []byte
	decoded, err := base64.StdEncoding.DecodeString(*opts.State)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	// TODO: validate md5, lineage

	// The docs (linked above) state the serial in the create options must match the
	// serial in the state file. However, the go-tfe integration tests we use
	// send different values for each and expect the serial in the create
	// options to take precedence, without error. We've opted to support that
	// behaviour.
	sv, err := h.svc.CreateStateVersion(r.Context(), CreateStateVersionOptions{
		WorkspaceID: otf.String(workspaceID),
		State:       decoded,
		Serial:      opts.Serial,
	})
	if err != nil {
		jsonapi.Error(w, err)
		return
	}
	h.writeResponse(w, r, sv, jsonapi.WithCode(http.StatusCreated))
}

func (h *api) listVersions(w http.ResponseWriter, r *http.Request) {
	var opts StateVersionListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		jsonapi.Error(w, err)
		return
	}
	svl, err := h.svc.ListStateVersions(r.Context(), opts)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}
	h.writeResponse(w, r, svl)
}

func (h *api) getCurrentVersion(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	sv, err := h.svc.GetCurrentStateVersion(r.Context(), workspaceID)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}
	h.writeResponse(w, r, sv)
}

func (h *api) getVersion(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.Param("id", r)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}
	sv, err := h.svc.GetStateVersion(r.Context(), versionID)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}
	h.writeResponse(w, r, sv)
}

func (h *api) downloadState(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.Param("id", r)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}
	resp, err := h.svc.DownloadState(r.Context(), versionID)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}
	w.Write(resp)
}

func (h *api) listOutputs(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.Param("id", r)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}
	sv, err := h.svc.GetStateVersion(r.Context(), versionID)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}
	h.writeResponse(w, r, sv.Outputs)
}

func (h *api) getOutput(w http.ResponseWriter, r *http.Request) {
	outputID, err := decode.Param("id", r)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}
	out, err := h.svc.GetStateVersionOutput(r.Context(), outputID)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}
	h.writeResponse(w, r, out)
}

// writeResponse encodes v as json:api and writes it to the body of the http response.
func (h *api) writeResponse(w http.ResponseWriter, r *http.Request, v any, opts ...func(http.ResponseWriter)) {
	var payload any

	switch v := v.(type) {
	case *VersionList:
		payload = h.toList(v)
	case *Version:
		payload = h.toVersion(v)
	case outputList:
		payload = h.toOutputList(v)
	case *Output:
		payload = h.toOutput(v)
	}
	jsonapi.WriteResponse(w, r, payload, opts...)
}
