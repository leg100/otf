package state

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/api/types"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/workspace"
	"golang.org/x/exp/maps"
)

type tfe struct {
	Service
	workspace.WorkspaceService
	*api.Responder
}

// Implements TFC state versions API:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/state-versions#state-versions-api
func (a *tfe) addHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	r.HandleFunc("/workspaces/{workspace_id}/state-versions", a.createVersion).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/current-state-version", a.getCurrentVersion).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/state-versions", a.rollbackVersion).Methods("PATCH")
	r.HandleFunc("/state-versions/{id}", a.getVersion).Methods("GET")
	r.HandleFunc("/state-versions", a.listVersionsByName).Methods("GET")
	r.HandleFunc("/state-versions/{id}/download", a.downloadState).Methods("GET")
	r.HandleFunc("/state-versions/{id}", a.deleteVersion).Methods("DELETE")

	r.HandleFunc("/workspaces/{workspace_id}/current-state-version-outputs", a.getCurrentVersionOutputs).Methods("GET")
	r.HandleFunc("/state-versions/{id}/outputs", a.listOutputs).Methods("GET")
	r.HandleFunc("/state-version-outputs/{id}", a.getOutput).Methods("GET")

	// specific to OTF
	r.HandleFunc("/workspaces/{workspace_id}/state-versions", a.listVersions).Methods("GET")
}

func (a *tfe) createVersion(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		api.Error(w, err)
		return
	}

	opts := types.StateVersionCreateVersionOptions{}
	if err := api.Unmarshal(r.Body, &opts); err != nil {
		api.Error(w, err)
		return
	}

	// required options
	if opts.Serial == nil {
		api.Error(w, &internal.MissingParameterError{Parameter: "serial"})
		return
	}
	if opts.MD5 == nil {
		api.Error(w, &internal.MissingParameterError{Parameter: "md5"})
		return
	}

	// base64-decode state to []byte
	decoded, err := base64.StdEncoding.DecodeString(*opts.State)
	if err != nil {
		api.Error(w, err)
		return
	}

	// validate md5 checksum
	if fmt.Sprintf("%x", md5.Sum(decoded)) != *opts.MD5 {
		api.Error(w, err)
		return
	}

	// TODO: validate lineage

	// The docs (linked above) state the serial in the create options must match the
	// serial in the state file. However, the go-tfe integration tests we use
	// send different values for each and expect the serial in the create
	// options to take precedence, without error. We've opted to support that
	// behaviour.
	sv, err := a.CreateStateVersion(r.Context(), CreateStateVersionOptions{
		WorkspaceID: internal.String(workspaceID),
		State:       decoded,
		Serial:      opts.Serial,
	})
	if err != nil {
		api.Error(w, err)
		return
	}

	to, err := a.toStateVersion(sv)
	if err != nil {
		api.Error(w, err)
		return
	}
	a.Respond(w, r, to, http.StatusCreated)
}

func (a *tfe) listVersionsByName(w http.ResponseWriter, r *http.Request) {
	var opts StateVersionListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		api.Error(w, err)
		return
	}
	ws, err := a.GetWorkspaceByName(r.Context(), opts.Organization, opts.Workspace)
	if err != nil {
		api.Error(w, err)
		return
	}
	page, err := a.ListStateVersions(r.Context(), ws.ID, opts.PageOptions)
	if err != nil {
		api.Error(w, err)
		return
	}

	// convert items
	items, err := a.toStateVersionList(page)
	if err != nil {
		api.Error(w, err)
		return
	}

	a.RespondWithPage(w, r, items, page.Pagination)
}

func (a *tfe) listVersions(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID string `schema:"workspace_id,required"`
		resource.PageOptions
	}
	if err := decode.All(&params, r); err != nil {
		api.Error(w, err)
		return
	}
	page, err := a.ListStateVersions(r.Context(), params.WorkspaceID, params.PageOptions)
	if err != nil {
		api.Error(w, err)
		return
	}

	// convert items
	items, err := a.toStateVersionList(page)
	if err != nil {
		api.Error(w, err)
		return
	}

	a.RespondWithPage(w, r, items, page.Pagination)
}

func (a *tfe) getCurrentVersion(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		api.Error(w, err)
		return
	}

	sv, err := a.GetCurrentStateVersion(r.Context(), workspaceID)
	if err != nil {
		api.Error(w, err)
		return
	}

	to, err := a.toStateVersion(sv)
	if err != nil {
		api.Error(w, err)
		return
	}
	a.Respond(w, r, to, http.StatusOK)
}

func (a *tfe) getVersion(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.Param("id", r)
	if err != nil {
		api.Error(w, err)
		return
	}
	sv, err := a.GetStateVersion(r.Context(), versionID)
	if err != nil {
		api.Error(w, err)
		return
	}

	to, err := a.toStateVersion(sv)
	if err != nil {
		api.Error(w, err)
		return
	}
	a.Respond(w, r, to, http.StatusOK)
}

func (a *tfe) deleteVersion(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.Param("id", r)
	if err != nil {
		api.Error(w, err)
		return
	}
	if err := a.DeleteStateVersion(r.Context(), versionID); err != nil {
		api.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) rollbackVersion(w http.ResponseWriter, r *http.Request) {
	opts := types.RollbackStateVersionOptions{}
	if err := api.Unmarshal(r.Body, &opts); err != nil {
		api.Error(w, err)
		return
	}

	sv, err := a.RollbackStateVersion(r.Context(), opts.RollbackStateVersion.ID)
	if err != nil {
		api.Error(w, err)
		return
	}

	a.Respond(w, r, sv, http.StatusOK)
}

func (a *tfe) downloadState(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.Param("id", r)
	if err != nil {
		api.Error(w, err)
		return
	}
	resp, err := a.DownloadState(r.Context(), versionID)
	if err != nil {
		api.Error(w, err)
		return
	}
	w.Write(resp)
}

func (a *tfe) getCurrentVersionOutputs(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		api.Error(w, err)
		return
	}

	sv, err := a.GetCurrentStateVersion(r.Context(), workspaceID)
	if err != nil {
		api.Error(w, err)
		return
	}

	// this particular endpoint does not reveal sensitive values:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/state-version-outputs#show-current-state-version-outputs-for-a-workspace
	to := make([]*types.StateVersionOutput, len(sv.Outputs))
	for i, f := range maps.Values(sv.Outputs) {
		to[i] = a.toOutput(f, true)
	}

	a.Respond(w, r, to, http.StatusOK)
}

func (a *tfe) listOutputs(w http.ResponseWriter, r *http.Request) {
	var params struct {
		StateVersionID string `schema:"id,required"`
		resource.PageOptions
	}
	if err := decode.All(&params, r); err != nil {
		api.Error(w, err)
		return
	}

	sv, err := a.GetStateVersion(r.Context(), params.StateVersionID)
	if err != nil {
		api.Error(w, err)
		return
	}

	// client expects a page of results, so convert outputs map to a page
	page := resource.NewPage(maps.Values(sv.Outputs), params.PageOptions, nil)

	a.Respond(w, r, page, http.StatusOK)
}

func (a *tfe) getOutput(w http.ResponseWriter, r *http.Request) {
	outputID, err := decode.Param("id", r)
	if err != nil {
		api.Error(w, err)
		return
	}
	out, err := a.GetStateVersionOutput(r.Context(), outputID)
	if err != nil {
		api.Error(w, err)
		return
	}

	a.Respond(w, r, out, http.StatusOK)
}

func (a *tfe) toStateVersion(from *Version) (*types.StateVersion, error) {
	var state File
	if err := json.Unmarshal(from.State, &state); err != nil {
		return nil, err
	}
	to := &types.StateVersion{
		ID:                 from.ID,
		CreatedAt:          from.CreatedAt,
		DownloadURL:        fmt.Sprintf("/api/v2/state-versions/%s/download", from.ID),
		Serial:             from.Serial,
		ResourcesProcessed: true,
		StateVersion:       state.Version,
		TerraformVersion:   state.TerraformVersion,
	}
	for _, out := range from.Outputs {
		to.Outputs = append(to.Outputs, &types.StateVersionOutput{ID: out.ID})
	}
	return to, nil
}

func (a *tfe) toStateVersionList(from *resource.Page[*Version]) ([]*types.StateVersion, error) {
	// convert items
	items := make([]*types.StateVersion, len(from.Items))
	for i, from := range from.Items {
		to, err := a.toStateVersion(from)
		if err != nil {
			return nil, err
		}
		items[i] = to
	}
	return items, nil
}

func (*tfe) toOutput(from *Output, scrubSensitive bool) *types.StateVersionOutput {
	to := &types.StateVersionOutput{
		ID:        from.ID,
		Name:      from.Name,
		Sensitive: from.Sensitive,
		Type:      from.Type,
		Value:     from.Value,
	}
	if to.Sensitive && scrubSensitive {
		to.Value = nil
	}
	return to
}

// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/state-versions#outputs
func (a *tfe) includeOutputs(ctx context.Context, v any) (any, error) {
	sv, ok := v.(*types.StateVersion)
	if !ok {
		return nil, nil
	}
	var include []any
	for _, outWithOnlyID := range sv.Outputs {
		out, err := a.GetStateVersionOutput(ctx, outWithOnlyID.ID)
		if err != nil {
			return nil, err
		}
		// do not scrub sensitive values for included outputs
		include = append(include, a.toOutput(out, false))
	}
	return include, nil
}

// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/workspaces#outputs
func (a *tfe) includeWorkspaceCurrentOutputs(ctx context.Context, v any) (any, error) {
	ws, ok := v.(*types.Workspace)
	if !ok {
		return nil, nil
	}
	sv, err := a.GetCurrentStateVersion(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	var include []any
	for _, from := range sv.Outputs {
		// scrub sensitive values for included outputs
		to := a.toOutput(from, true)
		include = append(include, to)
		ws.Outputs = append(ws.Outputs, &types.StateVersionOutput{
			ID: to.ID,
		})
	}
	return include, nil
}
