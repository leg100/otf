package http

import (
	"net/http"

	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/dto"
)

func (s *Server) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	var opts dto.WorkspaceCreateOptions
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := opts.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	ws, err := s.Application.CreateWorkspace(r.Context(), otf.WorkspaceCreateOptions{
		AllowDestroyPlan:           opts.AllowDestroyPlan,
		AutoApply:                  opts.AutoApply,
		Description:                opts.Description,
		ExecutionMode:              (*otf.ExecutionMode)(opts.ExecutionMode),
		FileTriggersEnabled:        opts.FileTriggersEnabled,
		GlobalRemoteState:          opts.GlobalRemoteState,
		MigrationEnvironment:       opts.MigrationEnvironment,
		Name:                       *opts.Name,
		OrganizationName:           *opts.Organization,
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
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &Workspace{ws}, withCode(http.StatusCreated))
}

func (s *Server) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	var spec otf.WorkspaceSpec
	if err := decode.Query(&spec, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := decode.Route(&spec, r); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	ws, err := s.Application.GetWorkspace(r.Context(), spec)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &Workspace{ws})
}

func (s *Server) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	var opts otf.WorkspaceListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	wsl, err := s.Application.ListWorkspaces(r.Context(), opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &WorkspaceList{wsl})
}

func (s *Server) UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
	opts := dto.WorkspaceUpdateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := opts.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	var spec otf.WorkspaceSpec
	if err := decode.Route(&spec, r); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	ws, err := s.Application.UpdateWorkspace(r.Context(), spec, otf.WorkspaceUpdateOptions{
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
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &Workspace{ws})
}

func (s *Server) LockWorkspace(w http.ResponseWriter, r *http.Request) {
	opts := otf.WorkspaceLockOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	spec := otf.WorkspaceSpec{}
	if err := decode.Route(&spec, r); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	ws, err := s.Application.LockWorkspace(r.Context(), spec, opts)
	if err == otf.ErrWorkspaceAlreadyLocked {
		writeError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &Workspace{ws})
}

func (s *Server) UnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	opts := otf.WorkspaceUnlockOptions{}
	spec := otf.WorkspaceSpec{}
	if err := decode.Route(&spec, r); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	ws, err := s.Application.UnlockWorkspace(r.Context(), spec, opts)
	if err == otf.ErrWorkspaceAlreadyUnlocked {
		writeError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &Workspace{ws})
}

func (s *Server) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	var spec otf.WorkspaceSpec
	if err := decode.Route(&spec, r); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := s.Application.DeleteWorkspace(r.Context(), spec); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type Workspace struct {
	*otf.Workspace
}

// ToJSONAPI assembles a JSONAPI DTO
func (ws *Workspace) ToJSONAPI() any {
	obj := &dto.Workspace{
		ID: ws.ID(),
		Actions: &dto.WorkspaceActions{
			IsDestroyable: false,
		},
		AllowDestroyPlan:     ws.AllowDestroyPlan(),
		AutoApply:            ws.AutoApply(),
		CanQueueDestroyPlan:  ws.CanQueueDestroyPlan(),
		CreatedAt:            ws.CreatedAt(),
		Description:          ws.Description(),
		Environment:          ws.Environment(),
		ExecutionMode:        string(ws.ExecutionMode()),
		FileTriggersEnabled:  ws.FileTriggersEnabled(),
		GlobalRemoteState:    ws.GlobalRemoteState(),
		Locked:               ws.Locked(),
		MigrationEnvironment: ws.MigrationEnvironment(),
		Name:                 ws.Name(),
		// Operations is deprecated but clients and go-tfe tests still use it
		Operations: ws.ExecutionMode() == "remote",
		Permissions: &dto.WorkspacePermissions{
			CanDestroy:        true,
			CanForceUnlock:    true,
			CanLock:           true,
			CanUnlock:         true,
			CanQueueApply:     true,
			CanQueueDestroy:   true,
			CanQueueRun:       true,
			CanReadSettings:   true,
			CanUpdate:         true,
			CanUpdateVariable: true,
		},
		QueueAllRuns:               ws.QueueAllRuns(),
		SpeculativeEnabled:         ws.SpeculativeEnabled(),
		SourceName:                 ws.SourceName(),
		SourceURL:                  ws.SourceURL(),
		StructuredRunOutputEnabled: ws.StructuredRunOutputEnabled(),
		TerraformVersion:           ws.TerraformVersion(),
		TriggerPrefixes:            ws.TriggerPrefixes(),
		WorkingDirectory:           ws.WorkingDirectory(),
		UpdatedAt:                  ws.UpdatedAt(),
	}
	if ws.Organization() != nil {
		obj.Organization = (&Organization{ws.Organization()}).ToJSONAPI().(*dto.Organization)
	} else {
		obj.Organization = &dto.Organization{ExternalID: ws.OrganizationID()}
	}
	return obj
}

type WorkspaceList struct {
	*otf.WorkspaceList
}

// ToJSONAPI assembles a JSON-API DTO.
func (l *WorkspaceList) ToJSONAPI() any {
	obj := &dto.WorkspaceList{
		Pagination: l.Pagination.ToJSONAPI(),
	}
	for _, item := range l.Items {
		obj.Items = append(obj.Items, (&Workspace{item}).ToJSONAPI().(*dto.Workspace))
	}
	return obj
}
