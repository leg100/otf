package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/go-tfe"
	"github.com/leg100/jsonapi"
	"github.com/leg100/ots"
)

func (h *Server) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := tfe.WorkspaceCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := opts.Valid(); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := h.WorkspaceService.Create(vars["org"], &opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, h.WorkspaceJSONAPIObject(obj), WithCode(http.StatusCreated))
}

func (h *Server) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := h.WorkspaceService.Get(vars["name"], vars["org"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, h.WorkspaceJSONAPIObject(obj))
}

func (h *Server) GetWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := h.WorkspaceService.GetByID(vars["id"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, h.WorkspaceJSONAPIObject(obj))
}

func (h *Server) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var opts tfe.WorkspaceListOptions
	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := h.WorkspaceService.List(vars["org"], opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, h.WorkspaceListJSONAPIObject(obj))
}

func (h *Server) UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := tfe.WorkspaceUpdateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := opts.Valid(); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := h.WorkspaceService.Update(vars["name"], vars["org"], &opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, h.WorkspaceJSONAPIObject(obj))
}

func (h *Server) UpdateWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := tfe.WorkspaceUpdateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := opts.Valid(); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := h.WorkspaceService.UpdateByID(vars["id"], &opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, h.WorkspaceJSONAPIObject(obj))
}

func (h *Server) LockWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := tfe.WorkspaceLockOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := h.WorkspaceService.Lock(vars["id"], opts)
	if err == ots.ErrWorkspaceAlreadyLocked {
		WriteError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, h.WorkspaceJSONAPIObject(obj))
}

func (h *Server) UnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := h.WorkspaceService.Unlock(vars["id"])
	if err == ots.ErrWorkspaceAlreadyUnlocked {
		WriteError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, h.WorkspaceJSONAPIObject(obj))
}

func (h *Server) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if err := h.WorkspaceService.Delete(vars["name"], vars["org"]); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Server) DeleteWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if err := h.WorkspaceService.DeleteByID(vars["id"]); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// WorkspaceJSONAPIObject converts a Workspace to a struct that can be marshalled into a
// JSON-API object
func (s *Server) WorkspaceJSONAPIObject(ws *ots.Workspace) *tfe.Workspace {
	obj := &tfe.Workspace{
		ID:                         ws.ExternalID,
		Actions:                    ws.Actions(),
		AllowDestroyPlan:           ws.AllowDestroyPlan,
		AutoApply:                  ws.AutoApply,
		CanQueueDestroyPlan:        ws.CanQueueDestroyPlan,
		CreatedAt:                  ws.CreatedAt,
		Description:                ws.Description,
		Environment:                ws.Environment,
		ExecutionMode:              ws.ExecutionMode,
		FileTriggersEnabled:        ws.FileTriggersEnabled,
		GlobalRemoteState:          ws.GlobalRemoteState,
		Locked:                     ws.Locked,
		MigrationEnvironment:       ws.MigrationEnvironment,
		Name:                       ws.Name,
		Operations:                 ws.Operations,
		Permissions:                ws.Permissions,
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
		ResourceCount:              ws.ResourceCount,
		ApplyDurationAverage:       ws.ApplyDurationAverage,
		PlanDurationAverage:        ws.PlanDurationAverage,
		PolicyCheckFailures:        ws.PolicyCheckFailures,
		RunFailures:                ws.RunFailures,
		RunsCount:                  ws.RunsCount,
	}

	if ws.Organization != nil {
		obj.Organization = s.OrganizationJSONAPIObject(ws.Organization)
	}

	return obj
}

// WorkspaceListJSONAPIObject converts a WorkspaceList to
// a struct that can be marshalled into a JSON-API object
func (s *Server) WorkspaceListJSONAPIObject(cvl *ots.WorkspaceList) *tfe.WorkspaceList {
	obj := &tfe.WorkspaceList{
		Pagination: cvl.Pagination,
	}
	for _, item := range cvl.Items {
		obj.Items = append(obj.Items, s.WorkspaceJSONAPIObject(item))
	}

	return obj
}
