package workspace

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/jsonapi"
)

type handlers struct {
	Application service
	*JSONAPIMarshaler
}

// byName are parameters used when looking up a workspace by
// name
type byName struct {
	Name         string `schema:"workspace_name,required"`
	Organization string `schema:"organization_name,required"`
}

func (h *handlers) AddHandlers(r *mux.Router) {
	// Workspace routes
	r.HandleFunc("/organizations/{organization_name}/workspaces", h.ListWorkspaces)
	r.HandleFunc("/organizations/{organization_name}/workspaces", h.create)
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}", h.GetWorkspaceByName)
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}", h.UpdateWorkspaceByName)
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}", h.DeleteWorkspaceByName)

	r.HandleFunc("/workspaces/{workspace_id}", h.UpdateWorkspace)
	r.HandleFunc("/workspaces/{workspace_id}", h.GetWorkspace)
	r.HandleFunc("/workspaces/{workspace_id}", h.DeleteWorkspace)
	r.HandleFunc("/workspaces/{workspace_id}/actions/lock", h.LockWorkspace)
	r.HandleFunc("/workspaces/{workspace_id}/actions/unlock", h.UnlockWorkspace)
}

func (s *handlers) create(w http.ResponseWriter, r *http.Request) {
	var opts jsonapi.WorkspaceCreateOptions
	if err := decode.Route(&opts, r); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := opts.Validate(); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	ws, err := s.Application.create(r.Context(), CreateWorkspaceOptions{
		AllowDestroyPlan:           opts.AllowDestroyPlan,
		AutoApply:                  opts.AutoApply,
		Description:                opts.Description,
		ExecutionMode:              (*otf.ExecutionMode)(opts.ExecutionMode),
		FileTriggersEnabled:        opts.FileTriggersEnabled,
		GlobalRemoteState:          opts.GlobalRemoteState,
		MigrationEnvironment:       opts.MigrationEnvironment,
		Name:                       opts.Name,
		Organization:               opts.Organization,
		QueueAllRuns:               opts.QueueAllRuns,
		SpeculativeEnabled:         opts.SpeculativeEnabled,
		SourceName:                 opts.SourceName,
		SourceURL:                  opts.SourceURL,
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
	jsonapi.WriteResponse(w, r, jworkspace, jsonapi.WithCode(http.StatusCreated))
}

func (s *handlers) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	ws, err := s.Application.GetWorkspace(r.Context(), id)
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

func (s *handlers) GetWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params byName
	if err := decode.All(&params, r); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	ws, err := s.Application.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
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

func (s *handlers) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	params := struct {
		Organization    string `schema:"organization_name,required"`
		otf.ListOptions        // Pagination
	}{}
	if err := decode.All(&params, r); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	wsl, err := s.Application.ListWorkspaces(r.Context(), WorkspaceListOptions{
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
func (s *handlers) UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
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
func (s *handlers) UpdateWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params byName
	if err := decode.Route(&params, r); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	ws, err := s.Application.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	s.updateWorkspace(w, r, ws.ID())
}

func (s *handlers) LockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	opts := otf.WorkspaceLockOptions{}
	ws, err := s.Application.LockWorkspace(r.Context(), id, opts)
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

func (s *handlers) UnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	opts := otf.WorkspaceUnlockOptions{}
	ws, err := s.Application.UnlockWorkspace(r.Context(), id, opts)
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

func (s *handlers) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	_, err = s.Application.DeleteWorkspace(r.Context(), workspaceID)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *handlers) DeleteWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params byName
	if err := decode.All(&params, r); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	ws, err := s.Application.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	_, err = s.Application.DeleteWorkspace(r.Context(), ws.ID())
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *handlers) updateWorkspace(w http.ResponseWriter, r *http.Request, workspaceID string) {
	opts := jsonapi.WorkspaceUpdateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := opts.Validate(); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	ws, err := s.Application.UpdateWorkspace(r.Context(), workspaceID, UpdateWorkspaceOptions{
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
