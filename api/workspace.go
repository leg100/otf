package api

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/workspace"
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
	var params jsonapi.WorkspaceCreateOptions
	if err := decode.Route(&params, r); err != nil {
		jsonapi.Error(w, err)
		return
	}
	if err := jsonapi.UnmarshalPayload(r.Body, &params); err != nil {
		jsonapi.Error(w, err)
		return
	}
	opts := workspace.CreateOptions{
		AllowDestroyPlan:           params.AllowDestroyPlan,
		AutoApply:                  params.AutoApply,
		Description:                params.Description,
		ExecutionMode:              (*workspace.ExecutionMode)(params.ExecutionMode),
		FileTriggersEnabled:        params.FileTriggersEnabled,
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
		WorkingDirectory:           params.WorkingDirectory,
	}
	if params.Operations != nil {
		if params.ExecutionMode != nil {
			err := errors.New("operations is deprecated and cannot be specified when execution mode is used")
			jsonapi.Error(w, err)
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
			err := errors.New("must specify both oauth-token-id and identifier attributes for vcs-repo")
			jsonapi.Error(w, err)
			return
		}
		opts.ConnectOptions = &workspace.ConnectOptions{
			RepoPath:      *params.VCSRepo.Identifier,
			VCSProviderID: *params.VCSRepo.OAuthTokenID,
		}
		if params.VCSRepo.Branch != nil {
			opts.Branch = params.VCSRepo.Branch
		}
	}

	ws, err := a.CreateWorkspace(r.Context(), opts)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	a.writeResponse(w, r, ws, jsonapi.WithCode(http.StatusCreated))
}

func (a *api) getWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	ws, err := a.GetWorkspace(r.Context(), id)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	a.writeResponse(w, r, ws)
}

func (a *api) getWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params byWorkspaceName
	if err := decode.All(&params, r); err != nil {
		jsonapi.Error(w, err)
		return
	}

	ws, err := a.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	a.writeResponse(w, r, ws)
}

func (a *api) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	var params workspace.ListOptions
	if err := decode.All(&params, r); err != nil {
		jsonapi.Error(w, err)
		return
	}

	wsl, err := a.ListWorkspaces(r.Context(), params)
	if err != nil {
		jsonapi.Error(w, err)
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
		jsonapi.Error(w, err)
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
		jsonapi.Error(w, err)
		return
	}

	ws, err := a.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	a.updateWorkspace(w, r, ws.ID)
}

func (a *api) lockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	ws, err := a.LockWorkspace(r.Context(), id, nil)
	if err != nil {
		jsonapi.Error(w, err)
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
		jsonapi.Error(w, err)
		return
	}

	_, err = a.DeleteWorkspace(r.Context(), workspaceID)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) deleteWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params byWorkspaceName
	if err := decode.All(&params, r); err != nil {
		jsonapi.Error(w, err)
		return
	}

	ws, err := a.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}
	_, err = a.DeleteWorkspace(r.Context(), ws.ID)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) updateWorkspace(w http.ResponseWriter, r *http.Request, workspaceID string) {
	opts := jsonapi.WorkspaceUpdateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, err)
		return
	}
	if err := opts.Validate(); err != nil {
		jsonapi.Error(w, err)
		return
	}

	ws, err := a.UpdateWorkspace(r.Context(), workspaceID, workspace.UpdateOptions{
		AllowDestroyPlan:           opts.AllowDestroyPlan,
		AutoApply:                  opts.AutoApply,
		Description:                opts.Description,
		ExecutionMode:              (*workspace.ExecutionMode)(opts.ExecutionMode),
		FileTriggersEnabled:        opts.FileTriggersEnabled,
		GlobalRemoteState:          opts.GlobalRemoteState,
		Name:                       opts.Name,
		QueueAllRuns:               opts.QueueAllRuns,
		SpeculativeEnabled:         opts.SpeculativeEnabled,
		StructuredRunOutputEnabled: opts.StructuredRunOutputEnabled,
		TerraformVersion:           opts.TerraformVersion,
		TriggerPrefixes:            opts.TriggerPrefixes,
		WorkingDirectory:           opts.WorkingDirectory,
	})
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	a.writeResponse(w, r, ws)
}

func (a *api) unlock(w http.ResponseWriter, r *http.Request, force bool) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	ws, err := a.UnlockWorkspace(r.Context(), id, nil, force)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	a.writeResponse(w, r, ws)
}
