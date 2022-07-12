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
	ws, err := s.WorkspaceService().Create(r.Context(), otf.WorkspaceCreateOptions{
		AllowDestroyPlan:           opts.AllowDestroyPlan,
		AutoApply:                  opts.AutoApply,
		Description:                opts.Description,
		ExecutionMode:              opts.ExecutionMode,
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
	writeResponse(w, r, ws, withCode(http.StatusCreated))
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
	ws, err := s.WorkspaceService().Get(r.Context(), spec)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, ws)
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
	ws, err := s.WorkspaceService().List(r.Context(), opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, ws)
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
	ws, err := s.WorkspaceService().Update(r.Context(), spec, otf.WorkspaceUpdateOptions{
		AllowDestroyPlan:           opts.AllowDestroyPlan,
		AutoApply:                  opts.AutoApply,
		Description:                opts.Description,
		ExecutionMode:              opts.ExecutionMode,
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
	writeResponse(w, r, ws)
}

func (s *Server) LockWorkspace(w http.ResponseWriter, r *http.Request) {
	user, err := userFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	opts := otf.WorkspaceLockOptions{
		Requestor: user,
	}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	spec := otf.WorkspaceSpec{}
	if err := decode.Route(&spec, r); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	ws, err := s.WorkspaceService().Lock(r.Context(), spec, opts)
	if err == otf.ErrWorkspaceAlreadyLocked {
		writeError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, ws)
}

func (s *Server) UnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	user, err := userFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	opts := otf.WorkspaceUnlockOptions{
		Requestor: user,
	}
	spec := otf.WorkspaceSpec{}
	if err := decode.Route(&spec, r); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	ws, err := s.WorkspaceService().Unlock(r.Context(), spec, opts)
	if err == otf.ErrWorkspaceAlreadyUnlocked {
		writeError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, ws)
}

func (s *Server) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	var spec otf.WorkspaceSpec
	if err := decode.Route(&spec, r); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := s.WorkspaceService().Delete(r.Context(), spec); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
