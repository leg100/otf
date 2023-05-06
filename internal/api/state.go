package api

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/api/types"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/state"
)

// Implements TFC state versions API:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/state-versions#state-versions-api
func (a *api) addStateHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	r.HandleFunc("/workspaces/{workspace_id}/state-versions", a.createVersion).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/current-state-version", a.getCurrentVersion).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/state-versions", a.rollbackVersion).Methods("PATCH")
	r.HandleFunc("/state-versions/{id}", a.getVersion).Methods("GET")
	r.HandleFunc("/state-versions", a.listVersions).Methods("GET")
	r.HandleFunc("/state-versions/{id}/download", a.downloadState).Methods("GET")
	r.HandleFunc("/state-versions/{id}", a.deleteVersion).Methods("DELETE")

	r.HandleFunc("/workspaces/{workspace_id}/current-state-version-outputs", a.getCurrentVersionOutputs).Methods("GET")
	r.HandleFunc("/state-versions/{id}/outputs", a.listOutputs).Methods("GET")
	r.HandleFunc("/state-version-outputs/{id}", a.getOutput).Methods("GET")
}

func (a *api) createVersion(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		Error(w, err)
		return
	}

	opts := types.StateVersionCreateVersionOptions{}
	if err := unmarshal(r.Body, &opts); err != nil {
		Error(w, err)
		return
	}

	// required options
	if opts.Serial == nil {
		Error(w, &internal.MissingParameterError{Parameter: "serial"})
		return
	}
	if opts.MD5 == nil {
		Error(w, &internal.MissingParameterError{Parameter: "md5"})
		return
	}

	// base64-decode state to []byte
	decoded, err := base64.StdEncoding.DecodeString(*opts.State)
	if err != nil {
		Error(w, err)
		return
	}

	// validate md5 checksum
	if fmt.Sprintf("%x", md5.Sum(decoded)) != *opts.MD5 {
		Error(w, err)
		return
	}

	// TODO: validate lineage

	// The docs (linked above) state the serial in the create options must match the
	// serial in the state file. However, the go-tfe integration tests we use
	// send different values for each and expect the serial in the create
	// options to take precedence, without error. We've opted to support that
	// behaviour.
	sv, err := a.CreateStateVersion(r.Context(), state.CreateStateVersionOptions{
		WorkspaceID: internal.String(workspaceID),
		State:       decoded,
		Serial:      opts.Serial,
	})
	if err != nil {
		Error(w, err)
		return
	}
	a.writeResponse(w, r, sv, withCode(http.StatusCreated))
}

func (a *api) listVersions(w http.ResponseWriter, r *http.Request) {
	var opts state.StateVersionListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		Error(w, err)
		return
	}
	svl, err := a.ListStateVersions(r.Context(), opts)
	if err != nil {
		Error(w, err)
		return
	}
	a.writeResponse(w, r, svl)
}

func (a *api) getCurrentVersion(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		Error(w, err)
		return
	}

	sv, err := a.GetCurrentStateVersion(r.Context(), workspaceID)
	if err != nil {
		Error(w, err)
		return
	}
	a.writeResponse(w, r, sv)
}

func (a *api) getVersion(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.Param("id", r)
	if err != nil {
		Error(w, err)
		return
	}
	sv, err := a.GetStateVersion(r.Context(), versionID)
	if err != nil {
		Error(w, err)
		return
	}
	a.writeResponse(w, r, sv)
}

func (a *api) deleteVersion(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.Param("id", r)
	if err != nil {
		Error(w, err)
		return
	}
	if err := a.DeleteStateVersion(r.Context(), versionID); err != nil {
		Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) rollbackVersion(w http.ResponseWriter, r *http.Request) {
	opts := types.RollbackStateVersionOptions{}
	if err := unmarshal(r.Body, &opts); err != nil {
		Error(w, err)
		return
	}

	sv, err := a.RollbackStateVersion(r.Context(), opts.RollbackStateVersion.ID)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, sv)
}

func (a *api) downloadState(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.Param("id", r)
	if err != nil {
		Error(w, err)
		return
	}
	resp, err := a.DownloadState(r.Context(), versionID)
	if err != nil {
		Error(w, err)
		return
	}
	w.Write(resp)
}

func (a *api) getCurrentVersionOutputs(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		Error(w, err)
		return
	}

	sv, err := a.GetCurrentStateVersion(r.Context(), workspaceID)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, sv.Outputs)
}

func (a *api) listOutputs(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.Param("id", r)
	if err != nil {
		Error(w, err)
		return
	}
	sv, err := a.GetStateVersion(r.Context(), versionID)
	if err != nil {
		Error(w, err)
		return
	}
	a.writeResponse(w, r, sv.Outputs)
}

func (a *api) getOutput(w http.ResponseWriter, r *http.Request) {
	outputID, err := decode.Param("id", r)
	if err != nil {
		Error(w, err)
		return
	}
	out, err := a.GetStateVersionOutput(r.Context(), outputID)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, out)
}
