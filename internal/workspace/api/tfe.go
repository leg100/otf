package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/workspace"
	"github.com/leg100/otf/internal/workspace/execution"
)

type (
	// tfe implements the TFC/E workspaces API:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/workspaces
	tfe struct {
		*tfeapi.Responder
		client     tfeClient
		authorizer *authz.Authorizer
	}

	tfeClient interface {
		GetWorkspace(context.Context, resource.TfeID) (*workspace.Workspace, error)
		ListWorkspaces(ctx context.Context, opts workspace.ListOptions) (*resource.Page[*workspace.Workspace], error)
		CreateWorkspace(ctx context.Context, opts workspace.CreateOptions) (*workspace.Workspace, error)
		GetWorkspaceByName(ctx context.Context, organization organization.Name, name string) (*workspace.Workspace, error)
		GetWorkspacePolicy(ctx context.Context, workspaceID resource.TfeID) (workspace.Policy, error)
		UpdateWorkspace(ctx context.Context, workspaceID resource.TfeID, opts workspace.UpdateOptions) (*workspace.Workspace, error)
		DeleteWorkspace(ctx context.Context, workspaceID resource.TfeID) (*workspace.Workspace, error)
		Lock(ctx context.Context, workspaceID resource.TfeID, runID *resource.TfeID) (*workspace.Workspace, error)
		Unlock(ctx context.Context, workspaceID resource.TfeID, runID *resource.TfeID, force bool) (*workspace.Workspace, error)

		DeleteTags(ctx context.Context, organization organization.Name, tagIDs []resource.TfeID) error
		AddTags(ctx context.Context, workspaceID resource.TfeID, tags []workspace.TagSpec) error
		ListTags(ctx context.Context, organization organization.Name, opts workspace.ListTagsOptions) (*resource.Page[*workspace.Tag], error)
		ListWorkspaceTags(ctx context.Context, workspaceID resource.TfeID, opts workspace.ListWorkspaceTagsOptions) (*resource.Page[*workspace.Tag], error)
		RemoveTags(ctx context.Context, workspaceID resource.TfeID, tags []workspace.TagSpec) error
		TagWorkspaces(ctx context.Context, tagID resource.TfeID, workspaceIDs []resource.TfeID) error
	}
)

func NewTFEAPI(
	client tfeClient,
	responder *tfeapi.Responder,
	authorizer *authz.Authorizer,
) *tfe {
	api := &tfe{
		Responder:  responder,
		client:     client,
		authorizer: authorizer,
	}

	// Fetch workspace when API calls request workspace be included in the
	// response
	responder.Register(tfeapi.IncludeWorkspace, api.include)
	responder.Register(tfeapi.IncludeWorkspaces, api.includeMany)

	return api
}

func (a *tfe) AddHandlers(r *mux.Router) {
	r.HandleFunc("/organizations/{organization_name}/workspaces", a.listWorkspaces).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/workspaces", a.createWorkspace).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}", a.getWorkspaceByName).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}", a.updateWorkspaceByName).Methods("PATCH")
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}", a.deleteWorkspaceByName).Methods("DELETE")

	r.HandleFunc("/workspaces/{workspace_id}", a.updateWorkspaceByID).Methods("PATCH")
	r.HandleFunc("/workspaces/{workspace_id}", a.getWorkspace).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}", a.deleteWorkspace).Methods("DELETE")
	r.HandleFunc("/workspaces/{workspace_id}/actions/lock", a.lockWorkspace).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/actions/unlock", a.unlockWorkspace).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/actions/force-unlock", a.forceUnlockWorkspace).Methods("POST")

	r.HandleFunc("/workspaces/{workspace_id}/relationships/ssh-key", a.assignSSHKey).Methods("PATCH")
	r.HandleFunc("/workspaces/{workspace_id}/relationships/ssh-key", a.unassignSSHKey).Methods("PATCH")

	a.addTagHandlers(r)
}

func (a *tfe) createWorkspace(w http.ResponseWriter, r *http.Request) {
	var params workspace.TFEWorkspaceCreateOptions
	if err := decode.Route(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}

	opts := workspace.CreateOptions{
		AgentPoolID:                params.AgentPoolID,
		AllowDestroyPlan:           params.AllowDestroyPlan,
		AutoApply:                  params.AutoApply,
		AutoApplyRunTrigger:        params.AutoApplyRunTrigger,
		Description:                params.Description,
		ExecutionKind:              params.ExecutionMode,
		GlobalRemoteState:          params.GlobalRemoteState,
		MigrationEnvironment:       params.MigrationEnvironment,
		Name:                       params.Name,
		Organization:               params.Organization,
		QueueAllRuns:               params.QueueAllRuns,
		SpeculativeEnabled:         params.SpeculativeEnabled,
		SourceName:                 params.SourceName,
		SourceURL:                  params.SourceURL,
		StructuredRunOutputEnabled: params.StructuredRunOutputEnabled,
		EngineVersion:              params.TerraformVersion,
		TriggerPrefixes:            params.TriggerPrefixes,
		TriggerPatterns:            params.TriggerPatterns,
		WorkingDirectory:           params.WorkingDirectory,
	}
	// convert from json:api structs to tag specs
	opts.Tags = make([]workspace.TagSpec, len(params.Tags))
	for i, tag := range params.Tags {
		opts.Tags[i] = workspace.TagSpec{ID: tag.ID, Name: tag.Name}
	}
	// Always trigger runs if neither trigger patterns nor tags regex are set
	if len(params.TriggerPatterns) == 0 && (params.VCSRepo == nil || params.VCSRepo.TagsRegex == nil) {
		opts.AlwaysTrigger = new(true)
	}
	if params.Operations != nil {
		if params.ExecutionMode != nil {
			err := errors.New("operations is deprecated and cannot be specified when execution mode is used")
			tfeapi.Error(w, err)
			return
		}
		if *params.Operations {
			opts.ExecutionKind = new(execution.RemoteKind)
		} else {
			opts.ExecutionKind = new(execution.LocalKind)
		}
	}
	if params.VCSRepo != nil {
		if params.VCSRepo.Identifier == nil || params.VCSRepo.OAuthTokenID == nil {
			tfeapi.Error(w, errors.New("must specify both oauth-token-id and identifier attributes for vcs-repo"))
			return
		}
		opts.ConnectOptions = &workspace.ConnectOptions{
			RepoPath:      params.VCSRepo.Identifier,
			VCSProviderID: params.VCSRepo.OAuthTokenID,
			Branch:        params.VCSRepo.Branch,
			TagsRegex:     params.VCSRepo.TagsRegex,
		}
	}

	ws, err := a.client.CreateWorkspace(r.Context(), opts)
	if err != nil {
		var opts []tfeapi.ErrorOption
		if errors.Is(err, engine.ErrInvalidVersion) {
			opts = append(opts, tfeapi.WithStatus(http.StatusUnprocessableEntity))
		}
		tfeapi.Error(w, err, opts...)
		return
	}

	converted, err := a.convert(ws, r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, converted, http.StatusCreated)
}

func (a *tfe) getWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.client.GetWorkspace(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	converted, err := a.convert(ws, r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, converted, http.StatusOK)
}

func (a *tfe) getWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params workspace.ByName
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.client.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	converted, err := a.convert(ws, r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, converted, http.StatusOK)
}

func (a *tfe) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	var pathParams struct {
		Organization organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&pathParams, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params workspace.TFEWorkspaceListOptions
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	page, err := a.client.ListWorkspaces(r.Context(), workspace.ListOptions{
		Search:       params.Search,
		Organization: &pathParams.Organization,
		PageOptions:  resource.PageOptions(params.PageOptions),
		Tags:         internal.SplitCSV(params.Tags),
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert items
	items := make([]*workspace.TFEWorkspace, len(page.Items))
	for i, from := range page.Items {
		to, err := a.convert(from, r)
		if err != nil {
			tfeapi.Error(w, err)
			return
		}
		items[i] = to
	}

	a.RespondWithPage(w, r, items, page.Pagination)
}

// updateWorkspaceByID updates a workspace using its ID.
//
// TODO: support updating workspace's vcs repo.
func (a *tfe) updateWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.updateWorkspace(w, r, workspaceID)
}

// updateWorkspaceByName updates a workspace using its name and organization.
//
// TODO: support updating workspace's vcs repo.
func (a *tfe) updateWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params workspace.ByName
	if err := decode.Route(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.client.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.updateWorkspace(w, r, ws.ID)
}

func (a *tfe) lockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.client.Lock(r.Context(), id, nil)
	if err != nil {
		if errors.Is(err, workspace.ErrWorkspaceAlreadyLocked) {
			http.Error(w, "", http.StatusConflict)
		} else {
			tfeapi.Error(w, err)
		}
		return
	}

	converted, err := a.convert(ws, r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, converted, http.StatusOK)
}

func (a *tfe) unlockWorkspace(w http.ResponseWriter, r *http.Request) {
	a.unlock(w, r, false)
}

func (a *tfe) forceUnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	a.unlock(w, r, true)
}

func (a *tfe) unlock(w http.ResponseWriter, r *http.Request, force bool) {
	id, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.client.Unlock(r.Context(), id, nil, force)
	if err != nil {
		if internal.ErrorIs(err, workspace.ErrWorkspaceAlreadyUnlocked, workspace.ErrWorkspaceLockedByRun) {
			tfeapi.Error(w, err, tfeapi.WithStatus(http.StatusConflict))
		} else {
			tfeapi.Error(w, err)
		}
		return
	}

	converted, err := a.convert(ws, r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, converted, http.StatusOK)
}

func (a *tfe) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	_, err = a.client.DeleteWorkspace(r.Context(), workspaceID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) deleteWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params workspace.ByName
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.client.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	_, err = a.client.DeleteWorkspace(r.Context(), ws.ID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) updateWorkspace(w http.ResponseWriter, r *http.Request, workspaceID resource.TfeID) {
	params := workspace.TFEWorkspaceUpdateOptions{}
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	if err := params.Validate(); err != nil {
		tfeapi.Error(w, err)
		return
	}

	if params.AgentPoolID != nil && params.AgentPoolID.IsZero() {
		// The tfe terraform provider sends an empty string for the agent pool
		// ID when it should be sending nil.
		//
		// https://github.com/leg100/otf/issues/827
		params.AgentPoolID = nil
	}

	opts := workspace.UpdateOptions{
		AgentPoolID:                params.AgentPoolID,
		AllowDestroyPlan:           params.AllowDestroyPlan,
		AutoApply:                  params.AutoApply,
		AutoApplyRunTrigger:        params.AutoApplyRunTrigger,
		Description:                params.Description,
		ExecutionKind:              params.ExecutionMode,
		GlobalRemoteState:          params.GlobalRemoteState,
		Name:                       params.Name,
		QueueAllRuns:               params.QueueAllRuns,
		SpeculativeEnabled:         params.SpeculativeEnabled,
		StructuredRunOutputEnabled: params.StructuredRunOutputEnabled,
		EngineVersion:              params.TerraformVersion,
		TriggerPrefixes:            params.TriggerPrefixes,
		TriggerPatterns:            params.TriggerPatterns,
		WorkingDirectory:           params.WorkingDirectory,
	}

	// If file-triggers-enabled is set to false and tags regex is unspecified
	// then enable always trigger runs for this workspace.
	//
	// TODO: return error when client has sent incompatible combinations of
	// options:
	// (a) file-triggers-enabled=true and tags-regex=non-nil
	// (b) file-triggers-enabled=true and trigger-prefixes=empty
	// (b) trigger-prefixes=non-empty and tags-regex=non-nil
	if (params.FileTriggersEnabled != nil && !*params.FileTriggersEnabled) && (!params.VCSRepo.Set || !params.VCSRepo.Valid || params.VCSRepo.TagsRegex == nil) {
		opts.AlwaysTrigger = new(true)
	}

	if params.VCSRepo.Set {
		if params.VCSRepo.Valid {
			// client has provided non-null vcs options, which means they either
			// want to connect the workspace or modify the connection.
			opts.ConnectOptions = &workspace.ConnectOptions{
				RepoPath:      params.VCSRepo.Identifier,
				VCSProviderID: params.VCSRepo.OAuthTokenID,
				Branch:        params.VCSRepo.Branch,
				TagsRegex:     params.VCSRepo.TagsRegex,
			}
		} else {
			// client has explicitly set VCS options to null, which means they
			// want the workspace to be disconnected.
			opts.Disconnect = true
		}
	}

	ws, err := a.client.UpdateWorkspace(r.Context(), workspaceID, opts)
	if err != nil {
		var opts []tfeapi.ErrorOption
		if errors.Is(err, engine.ErrInvalidVersion) {
			opts = append(opts, tfeapi.WithStatus(http.StatusUnprocessableEntity))
		}
		tfeapi.Error(w, err, opts...)
		return
	}

	converted, err := a.convert(ws, r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, converted, http.StatusOK)
}

func (a *tfe) assignSSHKey(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	opts := workspace.TFEAssignSSHKeyOptions{}
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.client.UpdateWorkspace(r.Context(), workspaceID, workspace.UpdateOptions{
		UpdateSSHKeyOptions: &workspace.UpdateSSHKeyOptions{
			SSHKeyID: opts.SSHKeyID,
		},
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	converted, err := a.convert(ws, r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, converted, http.StatusOK)
}

func (a *tfe) unassignSSHKey(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.client.UpdateWorkspace(r.Context(), workspaceID, workspace.UpdateOptions{
		UpdateSSHKeyOptions: &workspace.UpdateSSHKeyOptions{
			SSHKeyID: nil,
		},
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	converted, err := a.convert(ws, r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, converted, http.StatusOK)
}

func (a *tfe) convert(from *workspace.Workspace, r *http.Request) (*workspace.TFEWorkspace, error) {
	return workspace.ToTFE(a.authorizer, from, r)
}

func (a *tfe) include(ctx context.Context, v any) ([]any, error) {
	dst := reflect.Indirect(reflect.ValueOf(v))

	// v must be a struct with a field named Workspace of type *types.Workspace
	if dst.Kind() != reflect.Struct {
		return nil, nil
	}
	field := dst.FieldByName("Workspace")
	if !field.IsValid() {
		return nil, nil
	}
	onlyID, ok := field.Interface().(*workspace.TFEWorkspace)
	if !ok {
		return nil, nil
	}
	// onlyID only contains the ID field, e.g. TFEWorkspace{ID: "ws-123"}; so
	// now retrieve the fully populated workspace, convert to a tfe workspace
	// and return.
	ws, err := a.client.GetWorkspace(ctx, onlyID.ID)
	if err != nil {
		return nil, fmt.Errorf("retrieving workspace: %w", err)
	}
	include, err := a.convert(ws, (&http.Request{}).WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return []any{include}, nil
}

func (a *tfe) includeMany(ctx context.Context, v any) ([]any, error) {
	dst := reflect.Indirect(reflect.ValueOf(v))

	// v must be a struct with a field named Workspaces of type []*types.Workspace
	if dst.Kind() != reflect.Struct {
		return nil, nil
	}
	field := dst.FieldByName("Workspaces")
	if !field.IsValid() {
		return nil, nil
	}
	onlyIDs, ok := field.Interface().([]*workspace.TFEWorkspace)
	if !ok {
		return nil, nil
	}
	// onlyIDs only contains the ID field, e.g. []*TFEWorkspace{{ID: "ws-123"}};
	// so now retrieve the fully populated workspaces, convert and return them.
	include := make([]any, len(onlyIDs))
	for i, onlyID := range onlyIDs {
		ws, err := a.client.GetWorkspace(ctx, onlyID.ID)
		if err != nil {
			return nil, fmt.Errorf("retrieving workspace: %w", err)
		}
		include[i], err = a.convert(ws, (&http.Request{}).WithContext(ctx))
		if err != nil {
			return nil, err
		}
	}
	return include, nil
}
