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
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tfeapi/types"
	"github.com/leg100/otf/internal/workspace"
	"github.com/leg100/surl/v2"
	"golang.org/x/exp/maps"
)

type tfe struct {
	*tfeapi.Responder
	client tfeClient
	signer tfeapi.Signer
}

type tfeClient interface {
	CreateStateVersion(ctx context.Context, opts CreateStateVersionOptions) (*Version, error)
	GetCurrentStateVersion(ctx context.Context, workspaceID resource.ID) (*Version, error)
	ListStateVersions(ctx context.Context, workspaceID resource.ID, opts resource.PageOptions) (*resource.Page[*Version], error)
	GetStateVersion(ctx context.Context, id resource.ID) (*Version, error)
	RollbackStateVersion(ctx context.Context, id resource.ID) (*Version, error)
	DeleteStateVersion(ctx context.Context, id resource.ID) error
	GetPreviousStateVersion(ctx context.Context, sv *Version) (*Version, error)
	GetWorkspaceByName(ctx context.Context, organization resource.ID, workspace string) (*workspace.Workspace, error)
	UploadState(ctx context.Context, svID resource.ID, state []byte) error
	DownloadState(ctx context.Context, svID resource.ID) ([]byte, error)
	GetStateOutput(ctx context.Context, outputID resource.ID) (*Output, error)
}

func NewTFEAPI(
	client tfeClient,
	responder *tfeapi.Responder,
	signer *surl.Signer,
) *tfe {
	api := &tfe{
		Responder: responder,
		client:    client,
		signer:    signer,
	}

	// include state version outputs in api responses when requested.
	responder.Register(tfeapi.IncludeOutputs, api.includeOutputs)
	responder.Register(tfeapi.IncludeOutputs, api.includeWorkspaceCurrentOutputs)

	return api
}

// Implements TFC state versions API:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/state-versions#state-versions-api
func (a *tfe) AddHandlers(r *mux.Router) {
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

	//
	// Signed URLs
	//
	r.HandleFunc("/state-versions/{id}/upload", a.uploadState).Methods("PUT")
	// terraform as of v1.6.0 uploads a 'JSON' version of the state (by
	// which they mean a state using a well documented, public, schema rather than their
	// internal schema). OTF doesn't do anything yet with the JSON version but
	// in order to avoid breaking terraform (see
	// https://github.com/leg100/otf/issues/626) OTF accepts the upload but does
	// nothing with it.
	r.HandleFunc("/state-versions/{id}/upload/json", func(http.ResponseWriter, *http.Request) {})
}

func (a *tfe) createVersion(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	opts := TFEStateVersionCreateVersionOptions{}
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

	sv, err := a.client.CreateStateVersion(r.Context(), CreateStateVersionOptions{
		WorkspaceID: workspaceID,
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
	var opts struct {
		types.ListOptions
		Organization organization.Name `schema:"filter[organization][name],required"`
		Workspace    string            `schema:"filter[workspace][name],required"`
	}
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		tfeapi.Error(w, err)
		return
	}
	ws, err := a.client.GetWorkspaceByName(r.Context(), opts.Organization, opts.Workspace)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	page, err := a.client.ListStateVersions(r.Context(), ws.ID, resource.PageOptions(opts.ListOptions))
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

	sv, err := a.client.GetCurrentStateVersion(r.Context(), workspaceID)
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
	versionID, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	sv, err := a.client.GetStateVersion(r.Context(), versionID)
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
	versionID, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	if err := a.client.DeleteStateVersion(r.Context(), versionID); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) rollbackVersion(w http.ResponseWriter, r *http.Request) {
	opts := TFERollbackStateVersionOptions{}
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	sv, err := a.client.RollbackStateVersion(r.Context(), opts.RollbackStateVersion.ID)
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
	versionID, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		tfeapi.Error(w, err)
	}
	if err := a.client.UploadState(r.Context(), versionID, buf.Bytes()); err != nil {
		tfeapi.Error(w, err)
		return
	}
}

func (a *tfe) downloadState(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	resp, err := a.client.DownloadState(r.Context(), versionID)
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

	sv, err := a.client.GetCurrentStateVersion(r.Context(), workspaceID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// this particular endpoint does not reveal sensitive values:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/state-version-outputs#show-current-state-version-outputs-for-a-workspace
	to := make([]*TFEStateVersionOutput, len(sv.Outputs))
	for i, f := range maps.Values(sv.Outputs) {
		to[i] = a.toOutput(f, true)
	}

	a.Respond(w, r, to, http.StatusOK)
}

func (a *tfe) listOutputs(w http.ResponseWriter, r *http.Request) {
	var params struct {
		StateVersionID resource.TfeID `schema:"id,required"`
		types.ListOptions
	}
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	sv, err := a.client.GetStateVersion(r.Context(), params.StateVersionID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// client expects a page of results, so convert outputs map to a page
	page := resource.NewPage(maps.Values(sv.Outputs), resource.PageOptions(params.ListOptions), nil)

	// convert to list of tfe types
	items := make([]*TFEStateVersionOutput, len(page.Items))
	for i, from := range page.Items {
		items[i] = a.toOutput(from, false)
	}
	a.RespondWithPage(w, r, items, page.Pagination)
}

func (a *tfe) getOutput(w http.ResponseWriter, r *http.Request) {
	outputID, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	out, err := a.client.GetStateOutput(r.Context(), outputID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.toOutput(out, false), http.StatusOK)
}

func (a *tfe) toStateVersion(from *Version, r *http.Request) (*TFEStateVersion, error) {
	to := &TFEStateVersion{
		ID:                 from.ID,
		CreatedAt:          from.CreatedAt,
		Serial:             from.Serial,
		Status:             TFEStateVersionStatus(from.Status),
		DownloadURL:        fmt.Sprintf("/api/v2/state-versions/%s/download", from.ID),
		ResourcesProcessed: true,
		Outputs:            make([]*TFEStateVersionOutput, len(from.Outputs)),
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
		to.Outputs[i] = &TFEStateVersionOutput{ID: out.ID}
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

func (a *tfe) toStateVersionList(from *resource.Page[*Version], r *http.Request) ([]*TFEStateVersion, error) {
	// convert items
	items := make([]*TFEStateVersion, len(from.Items))
	for i, from := range from.Items {
		to, err := a.toStateVersion(from, r)
		if err != nil {
			return nil, err
		}
		items[i] = to
	}
	return items, nil
}

func (*tfe) toOutput(from *Output, scrubSensitive bool) *TFEStateVersionOutput {
	to := &TFEStateVersionOutput{
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
	to, ok := v.(*TFEStateVersion)
	if !ok {
		return nil, nil
	}
	// re-retrieve the state version, because the tfe state version only
	// possesses the IDs of the outputs, whereas we need the full output structs
	from, err := a.client.GetStateVersion(ctx, to.ID)
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
	ws, ok := v.(*workspace.TFEWorkspace)
	if !ok {
		return nil, nil
	}
	sv, err := a.client.GetCurrentStateVersion(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	include := make([]any, len(sv.Outputs))
	// TODO: we both include the full output types and populate the list of IDs
	// in ws.Outputs, but really the latter should *always* be populated, and
	// that should be the responsibility of the workspace pkg. To avoid an
	// import cycle, perhaps the workspace SQL queries could return a list of
	// output IDs.
	ws.Outputs = make([]*workspace.TFEWorkspaceOutput, len(sv.Outputs))
	var i int
	for _, from := range sv.Outputs {
		include[i] = &workspace.TFEWorkspaceOutput{
			ID:        from.ID,
			Name:      from.Name,
			Sensitive: from.Sensitive,
			Type:      from.Type,
			Value:     from.Value,
		}
		ws.Outputs[i] = &workspace.TFEWorkspaceOutput{
			ID: from.ID,
		}
		i++
	}
	return include, nil
}

func (a *tfe) generateSignedURL(r *http.Request, fmtpath string, args ...any) (string, error) {
	path := fmt.Sprintf(fmtpath, args...)
	signedPath, err := a.signer.Sign(path, time.Now().Add(time.Hour))
	if err != nil {
		return "", err
	}
	return otfhttp.Absolute(r, signedPath), nil
}
