package workspace

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tfeapi/types"
)

type (
	// byWorkspaceName are parameters used when looking up a workspace by
	// name
	byWorkspaceName struct {
		Name         string `schema:"workspace_name,required"`
		Organization string `schema:"organization_name,required"`
	}

	// tfe implements the TFC/E workspaces API:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/workspaces
	tfe struct {
		Service
		*tfeapi.Responder
	}
)

func (a *tfe) addHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

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
	var params types.WorkspaceCreateOptions
	if err := decode.Route(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}

	opts := CreateOptions{
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
		TerraformVersion:           params.TerraformVersion,
		TriggerPrefixes:            params.TriggerPrefixes,
		TriggerPatterns:            params.TriggerPatterns,
		WorkingDirectory:           params.WorkingDirectory,
		// convert from json:api structs to tag specs
		Tags: toTagSpecs(params.Tags),
	}
	// Always trigger runs if neither trigger patterns nor tags regex are set
	if len(params.TriggerPatterns) == 0 && (params.VCSRepo == nil || params.VCSRepo.TagsRegex == nil) {
		opts.AlwaysTrigger = internal.Bool(true)
	}
	if params.Operations != nil {
		if params.ExecutionMode != nil {
			err := errors.New("operations is deprecated and cannot be specified when execution mode is used")
			tfeapi.Error(w, err)
			return
		}
		if *params.Operations {
			opts.ExecutionMode = ExecutionModePtr(RemoteExecutionMode)
		} else {
			opts.ExecutionMode = ExecutionModePtr(LocalExecutionMode)
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

	ws, err := a.CreateWorkspace(r.Context(), opts)
	if err != nil {
		tfeapi.Error(w, err)
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
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.GetWorkspace(r.Context(), id)
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

	ws, err := a.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
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
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params types.WorkspaceListOptions
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	page, err := a.ListWorkspaces(r.Context(), ListOptions{
		Search:       params.Search,
		Organization: &organization,
		PageOptions:  resource.PageOptions(params.ListOptions),
		Tags:         internal.SplitCSV(params.Tags),
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert items
	items := make([]*types.Workspace, len(page.Items))
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
	workspaceID, err := decode.Param("workspace_id", r)
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

	ws, err := a.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.updateWorkspace(w, r, ws.ID)
}

func (a *tfe) lockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.LockWorkspace(r.Context(), id, nil)
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

func (a *tfe) unlockWorkspace(w http.ResponseWriter, r *http.Request) {
	a.unlock(w, r, false)
}

func (a *tfe) forceUnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	a.unlock(w, r, true)
}

func (a *tfe) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	_, err = a.DeleteWorkspace(r.Context(), workspaceID)
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

	ws, err := a.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	_, err = a.DeleteWorkspace(r.Context(), ws.ID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) updateWorkspace(w http.ResponseWriter, r *http.Request, workspaceID string) {
	params := types.WorkspaceUpdateOptions{}
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	if err := params.Validate(); err != nil {
		tfeapi.Error(w, err)
		return
	}

	opts := UpdateOptions{
		AllowDestroyPlan:           params.AllowDestroyPlan,
		AutoApply:                  params.AutoApply,
		Description:                params.Description,
		ExecutionMode:              (*ExecutionMode)(params.ExecutionMode),
		GlobalRemoteState:          params.GlobalRemoteState,
		Name:                       params.Name,
		QueueAllRuns:               params.QueueAllRuns,
		SpeculativeEnabled:         params.SpeculativeEnabled,
		StructuredRunOutputEnabled: params.StructuredRunOutputEnabled,
		TerraformVersion:           params.TerraformVersion,
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
		opts.AlwaysTrigger = internal.Bool(true)
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

	ws, err := a.UpdateWorkspace(r.Context(), workspaceID, opts)
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

func (a *tfe) unlock(w http.ResponseWriter, r *http.Request, force bool) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.UnlockWorkspace(r.Context(), id, nil, force)
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

func (a *tfe) convert(from *Workspace, r *http.Request) (*types.Workspace, error) {
	subject, err := internal.SubjectFromContext(r.Context())
	if err != nil {
		return nil, err
	}
	policy, err := a.GetPolicy(r.Context(), from.ID)
	if err != nil {
		return nil, err
	}
	perms := &types.WorkspacePermissions{
		CanLock:           subject.CanAccessWorkspace(rbac.LockWorkspaceAction, policy),
		CanUnlock:         subject.CanAccessWorkspace(rbac.UnlockWorkspaceAction, policy),
		CanForceUnlock:    subject.CanAccessWorkspace(rbac.UnlockWorkspaceAction, policy),
		CanQueueApply:     subject.CanAccessWorkspace(rbac.ApplyRunAction, policy),
		CanQueueDestroy:   subject.CanAccessWorkspace(rbac.ApplyRunAction, policy),
		CanQueueRun:       subject.CanAccessWorkspace(rbac.CreateRunAction, policy),
		CanDestroy:        subject.CanAccessWorkspace(rbac.DeleteWorkspaceAction, policy),
		CanReadSettings:   subject.CanAccessWorkspace(rbac.GetWorkspaceAction, policy),
		CanUpdate:         subject.CanAccessWorkspace(rbac.UpdateWorkspaceAction, policy),
		CanUpdateVariable: subject.CanAccessWorkspace(rbac.UpdateWorkspaceAction, policy),
	}

	to := &types.Workspace{
		ID: from.ID,
		Actions: &types.WorkspaceActions{
			IsDestroyable: true,
		},
		AllowDestroyPlan:     from.AllowDestroyPlan,
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
		TerraformVersion:           from.TerraformVersion,
		TriggerPrefixes:            from.TriggerPrefixes,
		TriggerPatterns:            from.TriggerPatterns,
		WorkingDirectory:           from.WorkingDirectory,
		TagNames:                   from.Tags,
		UpdatedAt:                  from.UpdatedAt,
		Organization:               &types.Organization{Name: from.Organization},
	}
	if len(from.TriggerPrefixes) > 0 || len(from.TriggerPatterns) > 0 {
		to.FileTriggersEnabled = true
	}
	if from.LatestRun != nil {
		to.CurrentRun = &types.Run{ID: from.LatestRun.ID}
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
			to.VCSRepo = &types.VCSRepo{
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
	tfeWorkspace, ok := field.Interface().(*types.Workspace)
	if !ok {
		return nil, nil
	}
	ws, err := a.GetWorkspace(ctx, tfeWorkspace.ID)
	if err != nil {
		return nil, fmt.Errorf("retrieving workspace: %w", err)
	}
	converted, err := a.convert(ws, (&http.Request{}).WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return []any{converted}, nil
}
