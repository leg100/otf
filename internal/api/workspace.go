package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/api/types"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/workspace"
)

type (
	// byWorkspaceName are parameters used when looking up a workspace by
	// name
	byWorkspaceName struct {
		Name         string `schema:"workspace_name,required"`
		Organization string `schema:"organization_name,required"`
	}
)

func (a *api) addWorkspaceHandlers(r *mux.Router) {
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

func (a *api) createWorkspace(w http.ResponseWriter, r *http.Request) {
	var params types.WorkspaceCreateOptions
	if err := decode.Route(&params, r); err != nil {
		Error(w, err)
		return
	}
	if err := unmarshal(r.Body, &params); err != nil {
		Error(w, err)
		return
	}

	opts := workspace.CreateOptions{
		AllowDestroyPlan:           params.AllowDestroyPlan,
		AutoApply:                  params.AutoApply,
		Description:                params.Description,
		ExecutionMode:              (*workspace.ExecutionMode)(params.ExecutionMode),
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
	if params.Operations != nil {
		if params.ExecutionMode != nil {
			err := errors.New("operations is deprecated and cannot be specified when execution mode is used")
			Error(w, err)
			return
		}
		if *params.Operations {
			opts.ExecutionMode = workspace.ExecutionModePtr(workspace.RemoteExecutionMode)
		} else {
			opts.ExecutionMode = workspace.ExecutionModePtr(workspace.LocalExecutionMode)
		}
	}
	if params.VCSRepo != nil {
		if params.VCSRepo.Identifier == nil || params.VCSRepo.OAuthTokenID == nil {
			Error(w, errors.New("must specify both oauth-token-id and identifier attributes for vcs-repo"))
			return
		}
		opts.ConnectOptions = &workspace.ConnectOptions{
			RepoPath:      params.VCSRepo.Identifier,
			VCSProviderID: params.VCSRepo.OAuthTokenID,
			Branch:        params.VCSRepo.Branch,
			TagsRegex:     params.VCSRepo.TagsRegex,
		}
	}

	ws, err := a.CreateWorkspace(r.Context(), opts)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, ws, withCode(http.StatusCreated))
}

func (a *api) getWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		Error(w, err)
		return
	}

	ws, err := a.GetWorkspace(r.Context(), id)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, ws)
}

func (a *api) getWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params byWorkspaceName
	if err := decode.All(&params, r); err != nil {
		Error(w, err)
		return
	}

	ws, err := a.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, ws)
}

func (a *api) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		Error(w, err)
		return
	}
	var params types.WorkspaceListOptions
	if err := decode.All(&params, r); err != nil {
		Error(w, err)
		return
	}

	wsl, err := a.ListWorkspaces(r.Context(), workspace.ListOptions{
		Search:       params.Search,
		Organization: &organization,
		PageOptions:  resource.PageOptions(params.ListOptions),
		Tags:         strings.FieldsFunc(params.Tags, func(r rune) bool { return r == ',' }),
	})
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, wsl)
}

// updateWorkspaceByID updates a workspace using its ID.
//
// TODO: support updating workspace's vcs repo.
func (a *api) updateWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		Error(w, err)
		return
	}

	a.updateWorkspace(w, r, workspaceID)
}

// updateWorkspaceByName updates a workspace using its name and organization.
//
// TODO: support updating workspace's vcs repo.
func (a *api) updateWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params byWorkspaceName
	if err := decode.Route(&params, r); err != nil {
		Error(w, err)
		return
	}

	ws, err := a.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		Error(w, err)
		return
	}

	a.updateWorkspace(w, r, ws.ID)
}

func (a *api) lockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		Error(w, err)
		return
	}

	ws, err := a.LockWorkspace(r.Context(), id, nil)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, ws)
}

func (a *api) unlockWorkspace(w http.ResponseWriter, r *http.Request) {
	a.unlock(w, r, false)
}

func (a *api) forceUnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	a.unlock(w, r, true)
}

func (a *api) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		Error(w, err)
		return
	}

	_, err = a.DeleteWorkspace(r.Context(), workspaceID)
	if err != nil {
		Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) deleteWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params byWorkspaceName
	if err := decode.All(&params, r); err != nil {
		Error(w, err)
		return
	}

	ws, err := a.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		Error(w, err)
		return
	}
	_, err = a.DeleteWorkspace(r.Context(), ws.ID)
	if err != nil {
		Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) updateWorkspace(w http.ResponseWriter, r *http.Request, workspaceID string) {
	params := types.WorkspaceUpdateOptions{}
	if err := unmarshal(r.Body, &params); err != nil {
		Error(w, err)
		return
	}
	if err := params.Validate(); err != nil {
		Error(w, err)
		return
	}

	opts := workspace.UpdateOptions{
		AllowDestroyPlan:           params.AllowDestroyPlan,
		AutoApply:                  params.AutoApply,
		Description:                params.Description,
		ExecutionMode:              (*workspace.ExecutionMode)(params.ExecutionMode),
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
	ws, err := a.UpdateWorkspace(r.Context(), workspaceID, opts)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, ws)
}

func (a *api) unlock(w http.ResponseWriter, r *http.Request, force bool) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		Error(w, err)
		return
	}

	ws, err := a.UnlockWorkspace(r.Context(), id, nil, force)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, ws)
}
