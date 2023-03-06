package workspace

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/jsonapi"
)

type api struct {
	*JSONAPIMarshaler

	svc             service
	tokenMiddleware mux.MiddlewareFunc
}

// byName are parameters used when looking up a workspace by
// name
type byName struct {
	Name         string `schema:"workspace_name,required"`
	Organization string `schema:"organization_name,required"`
}

// unlockOptions are POST options for unlocking a workspace via the API
type unlockOptions struct {
	Force bool `json:"force"`
}

func (a *api) addHandlers(r *mux.Router) {
	r.Use(a.tokenMiddleware) // require bearer token

	r.HandleFunc("/organizations/{organization_name}/workspaces", a.list)
	r.HandleFunc("/organizations/{organization_name}/workspaces", a.create)
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}", a.GetWorkspaceByName)
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}", a.UpdateWorkspaceByName)
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}", a.DeleteWorkspaceByName)

	r.HandleFunc("/workspaces/{workspace_id}", a.UpdateWorkspace)
	r.HandleFunc("/workspaces/{workspace_id}", a.GetWorkspace)
	r.HandleFunc("/workspaces/{workspace_id}", a.DeleteWorkspace)
	r.HandleFunc("/workspaces/{workspace_id}/actions/lock", a.LockWorkspace)
	r.HandleFunc("/workspaces/{workspace_id}/actions/unlock", a.UnlockWorkspace)
}

func (a *api) create(w http.ResponseWriter, r *http.Request) {
	var params jsonapi.WorkspaceCreateOptions
	if err := decode.Route(&params, r); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := jsonapi.UnmarshalPayload(r.Body, &params); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	opts := otf.CreateWorkspaceOptions{
		AllowDestroyPlan:           params.AllowDestroyPlan,
		AutoApply:                  params.AutoApply,
		Description:                params.Description,
		ExecutionMode:              (*otf.ExecutionMode)(params.ExecutionMode),
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
			jsonapi.Error(w, http.StatusUnprocessableEntity, err)
			return
		}
		if *params.Operations {
			opts.ExecutionMode = otf.ExecutionModePtr(otf.RemoteExecutionMode)
		} else {
			opts.ExecutionMode = otf.ExecutionModePtr(otf.LocalExecutionMode)
		}
	}
	if params.VCSRepo != nil {
		if params.VCSRepo.Identifier == nil || params.VCSRepo.OAuthTokenID == nil {
			err := errors.New("must specify both oauth-token-id and identifier attributes for vcs-repo")
			jsonapi.Error(w, http.StatusUnprocessableEntity, err)
			return
		}
		opts.Repo = &otf.Connection{
			Identifier:    *params.VCSRepo.Identifier,
			VCSProviderID: *params.VCSRepo.OAuthTokenID,
		}
		if params.VCSRepo.Branch != nil {
			opts.Branch = params.VCSRepo.Branch
		}
	}

	ws, err := a.svc.create(r.Context(), opts)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	jworkspace, err := a.toJSONAPI(ws, r)
	if err != nil {
		jsonapi.Error(w, http.StatusInternalServerError, err)
		return
	}
	jsonapi.WriteResponse(w, r, jworkspace, jsonapi.WithCode(http.StatusCreated))
}

func (a *api) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	ws, err := a.svc.get(r.Context(), id)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	jworkspace, err := a.toJSONAPI(ws, r)
	if err != nil {
		jsonapi.Error(w, http.StatusInternalServerError, err)
		return
	}
	jsonapi.WriteResponse(w, r, jworkspace)
}

func (a *api) GetWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params byName
	if err := decode.All(&params, r); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	ws, err := a.svc.getByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	jworkspace, err := a.toJSONAPI(ws, r)
	if err != nil {
		jsonapi.Error(w, http.StatusInternalServerError, err)
		return
	}
	jsonapi.WriteResponse(w, r, jworkspace)
}

func (s *api) list(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization    string `schema:"organization_name,required"`
		otf.ListOptions        // Pagination
	}
	if err := decode.All(&params, r); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	wsl, err := s.svc.list(r.Context(), otf.WorkspaceListOptions{
		Organization: &params.Organization,
		ListOptions:  params.ListOptions,
	})
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	jlist, err := s.toJSONAPIList(wsl, r)
	if err != nil {
		jsonapi.Error(w, http.StatusInternalServerError, err)
		return
	}
	jsonapi.WriteResponse(w, r, jlist)
}

// UpdateWorkspace updates a workspace using its ID.
//
// TODO: support updating workspace's vcs repo.
func (s *api) UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	s.updateWorkspace(w, r, workspaceID)
}

// UpdateWorkspaceByName updates a workspace using its name and organization.
//
// TODO: support updating workspace's vcs repo.
func (s *api) UpdateWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params byName
	if err := decode.Route(&params, r); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	ws, err := s.svc.getByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	s.updateWorkspace(w, r, ws.ID)
}

func (s *api) LockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	ws, err := s.svc.lock(r.Context(), id, nil)
	if err == otf.ErrWorkspaceAlreadyLocked {
		jsonapi.Error(w, http.StatusConflict, err)
		return
	} else if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	jworkspace, err := s.toJSONAPI(ws, r)
	if err != nil {
		jsonapi.Error(w, http.StatusInternalServerError, err)
		return
	}
	jsonapi.WriteResponse(w, r, jworkspace)
}

func (s *api) UnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	var opts unlockOptions
	if err := decode.Form(&opts, r); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	ws, err := s.svc.unlock(r.Context(), id, opts.Force)
	if err == otf.ErrWorkspaceAlreadyUnlocked {
		jsonapi.Error(w, http.StatusConflict, err)
		return
	} else if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	jworkspace, err := s.toJSONAPI(ws, r)
	if err != nil {
		jsonapi.Error(w, http.StatusInternalServerError, err)
		return
	}
	jsonapi.WriteResponse(w, r, jworkspace)
}

func (s *api) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	_, err = s.svc.delete(r.Context(), workspaceID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *api) DeleteWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params byName
	if err := decode.All(&params, r); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	ws, err := s.svc.getByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	_, err = s.svc.delete(r.Context(), ws.ID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *api) updateWorkspace(w http.ResponseWriter, r *http.Request, workspaceID string) {
	opts := jsonapi.WorkspaceUpdateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := opts.Validate(); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	ws, err := s.svc.update(r.Context(), workspaceID, otf.UpdateWorkspaceOptions{
		AllowDestroyPlan:           opts.AllowDestroyPlan,
		AutoApply:                  opts.AutoApply,
		Description:                opts.Description,
		ExecutionMode:              (*otf.ExecutionMode)(opts.ExecutionMode),
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
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	jworkspace, err := s.toJSONAPI(ws, r)
	if err != nil {
		jsonapi.Error(w, http.StatusInternalServerError, err)
		return
	}
	jsonapi.WriteResponse(w, r, jworkspace)
}
