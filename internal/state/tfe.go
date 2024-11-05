package state

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tfeapi/types"
	"github.com/leg100/otf/internal/workspace"
	"github.com/leg100/surl"
	"golang.org/x/exp/maps"
)

type tfe struct {
	*tfeapi.Responder
	*surl.Signer

	state      *Service
	workspaces *workspace.Service
}

// Implements TFC state versions API:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/state-versions#state-versions-api
func (a *tfe) addHandlers(r *mux.Router) {
	api := r.PathPrefix(tfeapi.APIPrefixV2).Subrouter()

	api.HandleFunc("/workspaces/{workspace_id}/state-versions", a.createVersion).Methods("POST")
	api.HandleFunc("/workspaces/{workspace_id}/current-state-version", a.getCurrentVersion).Methods("GET")
	api.HandleFunc("/workspaces/{workspace_id}/state-versions", a.rollbackVersion).Methods("PATCH")
	api.HandleFunc("/state-versions/{id}", a.getVersion).Methods("GET")
	api.HandleFunc("/state-versions", a.listVersionsByName).Methods("GET")
	api.HandleFunc("/state-versions/{id}/download", a.downloadState).Methods("GET")
	api.HandleFunc("/state-versions/{id}", a.deleteVersion).Methods("DELETE")

	api.HandleFunc("/workspaces/{workspace_id}/current-state-version-outputs", a.getCurrentVersionOutputs).Methods("GET")
	api.HandleFunc("/state-versions/{id}/outputs", a.listOutputs).Methods("GET")
	api.HandleFunc("/state-version-outputs/{id}", a.getOutput).Methods("GET")

	// verify signed URLs
	signed := r.PathPrefix("/signed/{signature.expiry}").Subrouter()
	signed.Use(internal.VerifySignedURL(a.Signer))
	signed.HandleFunc("/state-versions/{id}/upload", a.uploadState).Methods("PUT")
	// terraform as of v1.6.0 uploads a 'JSON' version of the state (by
	// which they mean a state using a well documented, public, schema rather than their
	// internal schema). OTF doesn't do anything yet with the JSON version but
	// in order to avoid breaking terraform (see
	// https://github.com/leg100/otf/issues/626) OTF accepts the upload but does
	// nothing with it.
	signed.HandleFunc("/state-versions/{id}/upload/json", func(http.ResponseWriter, *http.Request) {})
}

func (a *tfe) createVersion(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	opts := types.StateVersionCreateVersionOptions{}
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	// required options
	if opts.Serial == nil {
		tfeapi.Error(w, &internal.ErrMissingParameter{Parameter: "serial"})
		return
	}
	// TFE docs say md5 is a required option yet the state itself is optional.
	// OTF follows this behavour, mandating the md5 parameter, but it is only
	// actually used if the state is also provided at creation-time. If the
	// state is only later uploaded, the md5 is not used.
	if opts.MD5 == nil {
		tfeapi.Error(w, &internal.ErrMissingParameter{Parameter: "md5"})
		return
	}

	// state is optional as of terraform v1.6.0
	var state []byte
	if opts.State != nil {
		// base64-decode state to []byte
		decoded, err := base64.StdEncoding.DecodeString(*opts.State)
		if err != nil {
			tfeapi.Error(w, err)
			return
		}
		state = decoded
		// validate md5 checksum
		if fmt.Sprintf("%x", md5.Sum(state)) != *opts.MD5 {
			tfeapi.Error(w, err)
			return
		}
	}

	// TODO: validate lineage

	sv, err := a.state.Create(r.Context(), CreateStateVersionOptions{
		WorkspaceID: internal.String(workspaceID),
		State:       state,
		Serial:      opts.Serial,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	to, err := a.toStateVersion(sv, r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, to, http.StatusCreated)
}

func (a *tfe) listVersionsByName(w http.ResponseWriter, r *http.Request) {
	var opts StateVersionListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		tfeapi.Error(w, err)
		return
	}
	ws, err := a.workspaces.GetByName(r.Context(), opts.Organization, opts.Workspace)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	page, err := a.state.List(r.Context(), ws.ID, opts.PageOptions)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert items
	items, err := a.toStateVersionList(page, r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.RespondWithPage(w, r, items, page.Pagination)
}

func (a *tfe) getCurrentVersion(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	sv, err := a.state.GetCurrent(r.Context(), workspaceID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	to, err := a.toStateVersion(sv, r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, to, http.StatusOK)
}

func (a *tfe) getVersion(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.Param("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	sv, err := a.state.Get(r.Context(), versionID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	to, err := a.toStateVersion(sv, r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, to, http.StatusOK)
}

func (a *tfe) deleteVersion(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.Param("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	if err := a.state.Delete(r.Context(), versionID); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) rollbackVersion(w http.ResponseWriter, r *http.Request) {
	opts := types.RollbackStateVersionOptions{}
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	sv, err := a.state.Rollback(r.Context(), opts.RollbackStateVersion.ID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	to, err := a.toStateVersion(sv, r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, to, http.StatusOK)
}

func (a *tfe) uploadState(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.Param("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		tfeapi.Error(w, err)
	}
	if err := a.state.Upload(r.Context(), versionID, buf.Bytes()); err != nil {
		tfeapi.Error(w, err)
		return
	}
}

func (a *tfe) downloadState(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.Param("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	resp, err := a.state.Download(r.Context(), versionID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.Write(resp)
}

func (a *tfe) getCurrentVersionOutputs(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	sv, err := a.state.GetCurrent(r.Context(), workspaceID)
	if err != nil {
		tfeapi.Error(w, err)
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
		StateVersionID resource.ID `schema:"id,required"`
		resource.PageOptions
	}
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	sv, err := a.state.Get(r.Context(), params.StateVersionID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// client expects a page of results, so convert outputs map to a page
	page := resource.NewPage(maps.Values(sv.Outputs), params.PageOptions, nil)

	// convert to list of tfe types
	items := make([]*types.StateVersionOutput, len(page.Items))
	for i, from := range page.Items {
		items[i] = a.toOutput(from, false)
	}
	a.RespondWithPage(w, r, items, page.Pagination)
}

func (a *tfe) getOutput(w http.ResponseWriter, r *http.Request) {
	outputID, err := decode.Param("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	out, err := a.state.GetOutput(r.Context(), outputID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.toOutput(out, false), http.StatusOK)
}

func (a *tfe) toStateVersion(from *Version, r *http.Request) (*types.StateVersion, error) {
	to := &types.StateVersion{
		ID:                 from.ID,
		CreatedAt:          from.CreatedAt,
		Serial:             from.Serial,
		Status:             types.StateVersionStatus(from.Status),
		DownloadURL:        fmt.Sprintf("/api/v2/state-versions/%s/download", from.ID),
		ResourcesProcessed: true,
		Outputs:            make([]*types.StateVersionOutput, len(from.Outputs)),
	}
	// generate signed url for upload state endpoint
	uploadURL, err := a.generateSignedURL(r, "/state-versions/%s/upload", from.ID)
	if err != nil {
		return nil, err
	}
	to.UploadURL = uploadURL
	// generate signed url for upload json state endpoint
	jsonUploadURL, err := a.generateSignedURL(r, "/state-versions/%s/upload/json", from.ID)
	if err != nil {
		return nil, err
	}
	to.JSONUploadURL = jsonUploadURL
	for i, out := range maps.Values(from.Outputs) {
		to.Outputs[i] = &types.StateVersionOutput{ID: out.ID}
	}
	if from.State != nil {
		var state File
		if err := json.Unmarshal(from.State, &state); err != nil {
			return nil, err
		}
		to.StateVersion = state.Version
		to.TerraformVersion = state.TerraformVersion
	}
	return to, nil
}

func (a *tfe) toStateVersionList(from *resource.Page[*Version], r *http.Request) ([]*types.StateVersion, error) {
	// convert items
	items := make([]*types.StateVersion, len(from.Items))
	for i, from := range from.Items {
		to, err := a.toStateVersion(from, r)
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
func (a *tfe) includeOutputs(ctx context.Context, v any) ([]any, error) {
	to, ok := v.(*types.StateVersion)
	if !ok {
		return nil, nil
	}
	// re-retrieve the state version, because the tfe state version only
	// possesses the IDs of the outputs, whereas we need the full output structs
	from, err := a.state.Get(ctx, to.ID)
	if err != nil {
		return nil, err
	}
	include := make([]any, len(from.Outputs))
	for i, out := range maps.Values(from.Outputs) {
		// do not scrub sensitive values for included outputs
		include[i] = a.toOutput(out, false)
	}
	return include, nil
}

// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/workspaces#outputs
func (a *tfe) includeWorkspaceCurrentOutputs(ctx context.Context, v any) ([]any, error) {
	ws, ok := v.(*types.Workspace)
	if !ok {
		return nil, nil
	}
	sv, err := a.state.GetCurrent(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	include := make([]any, len(sv.Outputs))
	// TODO: we both include the full output types and populate the list of IDs
	// in ws.Outputs, but really the latter should *always* be populated, and
	// that should be the responsibility of the workspace pkg. To avoid an
	// import cycle, perhaps the workspace SQL queries could return a list of
	// output IDs.
	ws.Outputs = make([]*types.WorkspaceOutput, len(sv.Outputs))
	var i int
	for _, from := range sv.Outputs {
		include[i] = &types.WorkspaceOutput{
			ID:        from.ID,
			Name:      from.Name,
			Sensitive: from.Sensitive,
			Type:      from.Type,
			Value:     from.Value,
		}
		ws.Outputs[i] = &types.WorkspaceOutput{
			ID: from.ID,
		}
		i++
	}
	return include, nil
}

func (a *tfe) generateSignedURL(r *http.Request, fmtpath string, args ...any) (string, error) {
	path := fmt.Sprintf(fmtpath, args...)
	signedPath, err := a.Sign(path, time.Hour)
	if err != nil {
		return "", err
	}
	return otfhttp.Absolute(r, signedPath), nil
}
