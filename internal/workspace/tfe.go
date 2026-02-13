package workspace

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
)

type (
	// byWorkspaceName are parameters used when looking up a workspace by
	// name
	byWorkspaceName struct {
		Name         string            `schema:"workspace_name,required"`
		Organization organization.Name `schema:"organization_name,required"`
	}

	// tfe implements the TFC/E workspaces API:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/workspaces
	tfe struct {
		*Service
		*tfeapi.Responder
		*authz.Authorizer
	}
)

func (a *tfe) addHandlers(r *mux.Router) {
	r = r.PathPrefix(tfeapi.APIPrefixV2).Subrouter()

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
}

func (a *tfe) createWorkspace(w http.ResponseWriter, r *http.Request) {
	var params TFEWorkspaceCreateOptions
	if err := decode.Route(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}

	opts := CreateOptions{
		AgentPoolID:                params.AgentPoolID,
		AllowDestroyPlan:           params.AllowDestroyPlan,
		AutoApply:                  params.AutoApply,
		Description:                params.Description,
		ExecutionMode:              (*ExecutionMode)(params.ExecutionMode),
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
	opts.Tags = make([]TagSpec, len(params.Tags))
	for i, tag := range params.Tags {
		opts.Tags[i] = TagSpec{ID: tag.ID, Name: tag.Name}
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
			opts.ExecutionMode = new(RemoteExecutionMode)
		} else {
			opts.ExecutionMode = new(LocalExecutionMode)
		}
	}
	if params.VCSRepo != nil {
		if params.VCSRepo.Identifier == nil || params.VCSRepo.OAuthTokenID == nil {
			tfeapi.Error(w, errors.New("must specify both oauth-token-id and identifier attributes for vcs-repo"))
			return
		}
		opts.ConnectOptions = &ConnectOptions{
			RepoPath:      params.VCSRepo.Identifier,
			VCSProviderID: params.VCSRepo.OAuthTokenID,
			Branch:        params.VCSRepo.Branch,
			TagsRegex:     params.VCSRepo.TagsRegex,
		}
	}

	ws, err := a.Create(r.Context(), opts)
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

	ws, err := a.Get(r.Context(), id)
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
	var params byWorkspaceName
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.GetByName(r.Context(), params.Organization, params.Name)
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
	var params TFEWorkspaceListOptions
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	page, err := a.List(r.Context(), ListOptions{
		Search:       params.Search,
		Organization: &pathParams.Organization,
		PageOptions:  resource.PageOptions(params.ListOptions),
		Tags:         internal.SplitCSV(params.Tags),
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert items
	items := make([]*TFEWorkspace, len(page.Items))
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
	var params byWorkspaceName
	if err := decode.Route(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.GetByName(r.Context(), params.Organization, params.Name)
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

	ws, err := a.Lock(r.Context(), id, nil)
	if err != nil {
		if errors.Is(err, ErrWorkspaceAlreadyLocked) {
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

	ws, err := a.Unlock(r.Context(), id, nil, force)
	if err != nil {
		if internal.ErrorIs(err, ErrWorkspaceAlreadyUnlocked, ErrWorkspaceLockedByRun) {
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

	_, err = a.Delete(r.Context(), workspaceID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) deleteWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params byWorkspaceName
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.GetByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	_, err = a.Delete(r.Context(), ws.ID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) updateWorkspace(w http.ResponseWriter, r *http.Request, workspaceID resource.TfeID) {
	params := TFEWorkspaceUpdateOptions{}
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

	opts := UpdateOptions{
		AgentPoolID:                params.AgentPoolID,
		AllowDestroyPlan:           params.AllowDestroyPlan,
		AutoApply:                  params.AutoApply,
		Description:                params.Description,
		ExecutionMode:              (*ExecutionMode)(params.ExecutionMode),
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
			opts.ConnectOptions = &ConnectOptions{
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

	ws, err := a.Update(r.Context(), workspaceID, opts)
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

func (a *tfe) convert(from *Workspace, r *http.Request) (*TFEWorkspace, error) {
	return ToTFE(a.Authorizer, from, r)
}

// ToTFE converts an OTF workspace to a TFE workspace.
func ToTFE(a *authz.Authorizer, from *Workspace, r *http.Request) (*TFEWorkspace, error) {
	ctx := r.Context()
	perms := &TFEWorkspacePermissions{
		CanLock:           a.CanAccess(ctx, authz.LockWorkspaceAction, &from.ID),
		CanUnlock:         a.CanAccess(ctx, authz.UnlockWorkspaceAction, &from.ID),
		CanForceUnlock:    a.CanAccess(ctx, authz.UnlockWorkspaceAction, &from.ID),
		CanQueueApply:     a.CanAccess(ctx, authz.ApplyRunAction, &from.ID),
		CanQueueDestroy:   a.CanAccess(ctx, authz.ApplyRunAction, &from.ID),
		CanQueueRun:       a.CanAccess(ctx, authz.CreateRunAction, &from.ID),
		CanDestroy:        a.CanAccess(ctx, authz.DeleteWorkspaceAction, &from.ID),
		CanReadSettings:   a.CanAccess(ctx, authz.GetWorkspaceAction, &from.ID),
		CanUpdate:         a.CanAccess(ctx, authz.UpdateWorkspaceAction, &from.ID),
		CanUpdateVariable: a.CanAccess(ctx, authz.UpdateWorkspaceAction, &from.ID),
	}

	to := &TFEWorkspace{
		ID: from.ID,
		Actions: &TFEWorkspaceActions{
			IsDestroyable: true,
		},
		AllowDestroyPlan:     from.AllowDestroyPlan,
		AgentPoolID:          from.AgentPoolID,
		AutoApply:            from.AutoApply,
		CanQueueDestroyPlan:  from.CanQueueDestroyPlan,
		CreatedAt:            from.CreatedAt,
		Description:          from.Description,
		Environment:          from.Environment,
		ExecutionMode:        string(from.ExecutionMode),
		GlobalRemoteState:    from.GlobalRemoteState,
		Locked:               from.Locked(),
		MigrationEnvironment: from.MigrationEnvironment,
		Name:                 from.Name,
		// Operations is deprecated but clients and go-tfe tests still use it
		Operations:                 from.ExecutionMode == "remote",
		Permissions:                perms,
		QueueAllRuns:               from.QueueAllRuns,
		SpeculativeEnabled:         from.SpeculativeEnabled,
		SourceName:                 from.SourceName,
		SourceURL:                  from.SourceURL,
		StructuredRunOutputEnabled: from.StructuredRunOutputEnabled,
		TerraformVersion:           from.EngineVersion,
		TriggerPrefixes:            from.TriggerPrefixes,
		TriggerPatterns:            from.TriggerPatterns,
		WorkingDirectory:           from.WorkingDirectory,
		TagNames:                   from.Tags,
		UpdatedAt:                  from.UpdatedAt,
		Organization:               &organization.TFEOrganization{Name: from.Organization},
	}
	if len(from.TriggerPrefixes) > 0 || len(from.TriggerPatterns) > 0 {
		to.FileTriggersEnabled = true
	}
	if from.LatestRun != nil {
		to.CurrentRun = &TFERun{ID: from.LatestRun.ID}
	}

	// Add VCS repo to json:api struct if connected. NOTE: the terraform CLI
	// uses the presence of VCS repo to determine whether to allow a terraform
	// apply or not, displaying the following message if not:
	//
	//	Apply not allowed for workspaces with a VCS connection
	//
	//	A workspace that is connected to a VCS requires the VCS-driven workflow to ensure that the VCS remains the single source of truth.
	//
	// OTF permits the user to disable this behaviour by ommiting this info and
	// fool the terraform CLI into thinking its not a workspace with a VCS
	// connection.
	if from.Connection != nil {
		if !from.Connection.AllowCLIApply || !tfeapi.IsTerraformCLI(r) {
			to.VCSRepo = &TFEVCSRepo{
				OAuthTokenID: from.Connection.VCSProviderID,
				Branch:       from.Connection.Branch,
				Identifier:   from.Connection.Repo,
				TagsRegex:    from.Connection.TagsRegex,
			}
		}
	}
	return to, nil
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
	onlyID, ok := field.Interface().(*TFEWorkspace)
	if !ok {
		return nil, nil
	}
	// onlyID only contains the ID field, e.g. TFEWorkspace{ID: "ws-123"}; so
	// now retrieve the fully populated workspace, convert to a tfe workspace
	// and return.
	ws, err := a.Get(ctx, onlyID.ID)
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
	onlyIDs, ok := field.Interface().([]*TFEWorkspace)
	if !ok {
		return nil, nil
	}
	// onlyIDs only contains the ID field, e.g. []*TFEWorkspace{{ID: "ws-123"}};
	// so now retrieve the fully populated workspaces, convert and return them.
	include := make([]any, len(onlyIDs))
	for i, onlyID := range onlyIDs {
		ws, err := a.Get(ctx, onlyID.ID)
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
