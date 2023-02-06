package workspace

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/rbac"
)

// workspaceNameParams are those parameters used when looking up a workspace by
// name
type workspaceNameParams struct {
	Name         string `schema:"workspace_name,required"`
	Organization string `schema:"organization_name,required"`
}

// WorkspaceList assembles a workspace list JSONAPI DTO
type WorkspaceList struct {
	r *http.Request
	otf.Application
	*otf.WorkspaceList
}

func (l *WorkspaceList) ToJSONAPI() any {
	obj := &jsonapi.WorkspaceList{
		Pagination: l.Pagination.ToJSONAPI(),
	}
	for _, item := range l.Items {
		obj.Items = append(obj.Items, (&Workspace{l.r, l.Application, item}).ToJSONAPI().(*jsonapi.Workspace))
	}
	return obj
}

func (h *handlers) AddHandlers(r *mux.Router) {
	// Workspace routes
	r.HandleFunc("/organizations/{organization_name}/workspaces", h.ListWorkspaces)
	r.HandleFunc("/organizations/{organization_name}/workspaces", h.CreateWorkspace)
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}", h.GetWorkspaceByName)
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}", h.UpdateWorkspaceByName)
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}", h.DeleteWorkspaceByName)

	r.HandleFunc("/workspaces/{workspace_id}", h.UpdateWorkspace)
	r.HandleFunc("/workspaces/{workspace_id}", h.GetWorkspace)
	r.HandleFunc("/workspaces/{workspace_id}", h.DeleteWorkspace)
	r.HandleFunc("/workspaces/{workspace_id}/actions/lock", h.LockWorkspace)
	r.HandleFunc("/workspaces/{workspace_id}/actions/unlock", h.UnlockWorkspace)
}

func (s *Server) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	var opts jsonapiCreateOptions
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

	ws, err := s.Application.CreateWorkspace(r.Context(), CreateWorkspaceOptions{
		AllowDestroyPlan:           opts.AllowDestroyPlan,
		AutoApply:                  opts.AutoApply,
		Description:                opts.Description,
		ExecutionMode:              (*ExecutionMode)(opts.ExecutionMode),
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
	jsonapi.WriteResponse(w, r, &Workspace{r, s.Application, ws}, jsonapi.WithCode(http.StatusCreated))
}

func (s *Server) GetWorkspace(w http.ResponseWriter, r *http.Request) {
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
	jsonapi.WriteResponse(w, r, &Workspace{r, s.Application, ws})
}

func (s *Server) GetWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params workspaceNameParams
	if err := decode.All(&params, r); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	ws, err := s.Application.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &Workspace{r, s.Application, ws})
}

func (s *Server) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	var opts WorkspaceListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := decode.Route(&opts, r); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	wsl, err := s.Application.ListWorkspaces(r.Context(), opts)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &WorkspaceList{r, s.Application, wsl})
}

// UpdateWorkspace updates a workspace using its ID.
//
// TODO: support updating workspace's vcs repo.
func (s *Server) UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
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
func (s *Server) UpdateWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params workspaceNameParams
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

func (s *Server) LockWorkspace(w http.ResponseWriter, r *http.Request) {
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
	jsonapi.WriteResponse(w, r, &Workspace{r, s.Application, ws})
}

func (s *Server) UnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	opts := otf.WorkspaceUnlockOptions{}
	ws, err := s.Application.UnlockWorkspace(r.Context(), id, opts)
	if err == ErrWorkspaceAlreadyUnlocked {
		jsonapi.Error(w, http.StatusConflict, err)
		return
	} else if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &Workspace{r, s.Application, ws})
}

func (s *Server) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) DeleteWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params workspaceNameParams
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

func (s *Server) updateWorkspace(w http.ResponseWriter, r *http.Request, workspaceID string) {
	opts := jsonapiUpdateOptions{}
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
		ExecutionMode:              (*ExecutionMode)(opts.ExecutionMode),
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
	jsonapi.WriteResponse(w, r, &Workspace{r, s.Application, ws})
}
