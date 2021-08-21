package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/go-tfe"
	"github.com/leg100/jsonapi"
	"github.com/leg100/ots"
)

func (s *Server) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
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

	obj, err := s.WorkspaceService.Create(vars["org"], &opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.WorkspaceJSONAPIObject(obj), WithCode(http.StatusCreated))
}

func (s *Server) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	spec := ots.WorkspaceSpecifier{
		Name:             ots.String(vars["name"]),
		OrganizationName: ots.String(vars["org"]),
	}

	obj, err := s.WorkspaceService.Get(spec)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.WorkspaceJSONAPIObject(obj))
}

func (s *Server) GetWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	spec := ots.WorkspaceSpecifier{
		ID: ots.String(vars["id"]),
	}

	obj, err := s.WorkspaceService.Get(spec)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.WorkspaceJSONAPIObject(obj))
}

func (s *Server) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// Unmarshal query into opts struct
	var opts ots.WorkspaceListOptions
	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	// Add org name from path to opts
	organizationName := vars["org"]
	opts.OrganizationName = &organizationName

	obj, err := s.WorkspaceService.List(opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.WorkspaceListJSONAPIObject(obj))
}

func (s *Server) UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
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

	spec := ots.WorkspaceSpecifier{
		Name:             ots.String(vars["name"]),
		OrganizationName: ots.String(vars["org"]),
	}

	obj, err := s.WorkspaceService.Update(spec, &opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.WorkspaceJSONAPIObject(obj))
}

func (s *Server) UpdateWorkspaceByID(w http.ResponseWriter, r *http.Request) {
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

	spec := ots.WorkspaceSpecifier{
		ID: ots.String(vars["id"]),
	}

	obj, err := s.WorkspaceService.Update(spec, &opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.WorkspaceJSONAPIObject(obj))
}

func (s *Server) LockWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := tfe.WorkspaceLockOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.WorkspaceService.Lock(vars["id"], opts)
	if err == ots.ErrWorkspaceAlreadyLocked {
		WriteError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.WorkspaceJSONAPIObject(obj))
}

func (s *Server) UnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := s.WorkspaceService.Unlock(vars["id"])
	if err == ots.ErrWorkspaceAlreadyUnlocked {
		WriteError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.WorkspaceJSONAPIObject(obj))
}

func (s *Server) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	spec := ots.WorkspaceSpecifier{
		Name:             ots.String(vars["name"]),
		OrganizationName: ots.String(vars["org"]),
	}

	if err := s.WorkspaceService.Delete(spec); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) DeleteWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	spec := ots.WorkspaceSpecifier{ID: ots.String(vars["id"])}

	if err := s.WorkspaceService.Delete(spec); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// WorkspaceJSONAPIObject converts a Workspace to a struct that can be marshalled into a
// JSON-API object
func (s *Server) WorkspaceJSONAPIObject(ws *ots.Workspace) *tfe.Workspace {
	obj := &tfe.Workspace{
		ID:                         ws.ID,
		Actions:                    ws.Actions(),
		AllowDestroyPlan:           ws.AllowDestroyPlan,
		AutoApply:                  ws.AutoApply,
		CanQueueDestroyPlan:        ws.CanQueueDestroyPlan,
		CreatedAt:                  ws.Model.CreatedAt,
		Description:                ws.Description,
		Environment:                ws.Environment,
		ExecutionMode:              ws.ExecutionMode,
		FileTriggersEnabled:        ws.FileTriggersEnabled,
		GlobalRemoteState:          ws.GlobalRemoteState,
		Locked:                     ws.Locked,
		MigrationEnvironment:       ws.MigrationEnvironment,
		Name:                       ws.Name,
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
		UpdatedAt:                  ws.Model.UpdatedAt,
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
