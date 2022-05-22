package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/dto"
)

func (s *Server) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := otf.WorkspaceCreateOptions{
		Organization: vars["org"],
	}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.WorkspaceService().Create(r.Context(), opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, WorkspaceDTO(obj), WithCode(http.StatusCreated))
}

func (s *Server) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	spec := otf.WorkspaceSpec{
		Name:             otf.String(vars["name"]),
		OrganizationName: otf.String(vars["org"]),
	}

	obj, err := s.WorkspaceService().Get(r.Context(), spec)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, WorkspaceDTO(obj))
}

func (s *Server) GetWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	spec := otf.WorkspaceSpec{
		ID: otf.String(vars["id"]),
	}

	obj, err := s.WorkspaceService().Get(r.Context(), spec)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, WorkspaceDTO(obj))
}

func (s *Server) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// Unmarshal query into opts struct
	var opts otf.WorkspaceListOptions
	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	// Add org name from path to opts
	opts.OrganizationName = vars["org"]

	obj, err := s.WorkspaceService().List(r.Context(), opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, WorkspaceListJSONAPIObject(obj))
}

func (s *Server) UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := otf.WorkspaceUpdateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	spec := otf.WorkspaceSpec{
		Name:             otf.String(vars["name"]),
		OrganizationName: otf.String(vars["org"]),
	}

	obj, err := s.WorkspaceService().Update(r.Context(), spec, opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, WorkspaceDTO(obj))
}

func (s *Server) UpdateWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := otf.WorkspaceUpdateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	spec := otf.WorkspaceSpec{
		ID: otf.String(vars["id"]),
	}

	obj, err := s.WorkspaceService().Update(r.Context(), spec, opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, WorkspaceDTO(obj))
}

func (s *Server) LockWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := otf.WorkspaceLockOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	id := vars["id"]
	spec := otf.WorkspaceSpec{
		ID: &id,
	}

	obj, err := s.WorkspaceService().Lock(r.Context(), spec, opts)
	if err == otf.ErrWorkspaceAlreadyLocked {
		WriteError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, WorkspaceDTO(obj))
}

func (s *Server) UnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id := vars["id"]
	spec := otf.WorkspaceSpec{
		ID: &id,
	}

	obj, err := s.WorkspaceService().Unlock(r.Context(), spec)
	if err == otf.ErrWorkspaceAlreadyUnlocked {
		WriteError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, WorkspaceDTO(obj))
}

func (s *Server) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	spec := otf.WorkspaceSpec{
		Name:             otf.String(vars["name"]),
		OrganizationName: otf.String(vars["org"]),
	}

	if err := s.WorkspaceService().Delete(r.Context(), spec); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) DeleteWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	spec := otf.WorkspaceSpec{ID: otf.String(vars["id"])}

	if err := s.WorkspaceService().Delete(r.Context(), spec); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// WorkspaceDTO converts a Workspace to a struct that can be marshalled into a
// JSON-API object
func WorkspaceDTO(ws *otf.Workspace) *dto.Workspace {
	obj := &dto.Workspace{
		ID: ws.ID,
		Actions: &dto.WorkspaceActions{
			IsDestroyable: false,
		},
		AllowDestroyPlan:     ws.AllowDestroyPlan,
		AutoApply:            ws.AutoApply,
		CanQueueDestroyPlan:  ws.CanQueueDestroyPlan,
		CreatedAt:            ws.CreatedAt,
		Description:          ws.Description,
		Environment:          ws.Environment,
		ExecutionMode:        ws.ExecutionMode,
		FileTriggersEnabled:  ws.FileTriggersEnabled,
		GlobalRemoteState:    ws.GlobalRemoteState,
		Locked:               ws.Locked,
		MigrationEnvironment: ws.MigrationEnvironment,
		Name:                 ws.Name,
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
		QueueAllRuns:               ws.QueueAllRuns,
		SpeculativeEnabled:         ws.SpeculativeEnabled,
		SourceName:                 ws.SourceName,
		SourceURL:                  ws.SourceURL,
		StructuredRunOutputEnabled: ws.StructuredRunOutputEnabled,
		TerraformVersion:           ws.TerraformVersion,
		TriggerPrefixes:            ws.TriggerPrefixes,
		VCSRepo:                    ws.VCSRepo,
		WorkingDirectory:           ws.WorkingDirectory,
		UpdatedAt:                  ws.UpdatedAt,
	}

	if ws.ExecutionMode == "remote" {
		// Operations is deprecated but clients and go-tfe tests still use it
		obj.Operations = true
	}

	if ws.Organization != nil {
		obj.Organization = OrganizationDTO(ws.Organization)
	}

	return obj
}

// WorkspaceListJSONAPIObject converts a WorkspaceList to
// a struct that can be marshalled into a JSON-API object
func WorkspaceListJSONAPIObject(l *otf.WorkspaceList) *dto.WorkspaceList {
	pagination := dto.Pagination(*l.Pagination)
	obj := &dto.WorkspaceList{
		Pagination: &pagination,
	}
	for _, item := range l.Items {
		obj.Items = append(obj.Items, WorkspaceDTO(item))
	}

	return obj
}
