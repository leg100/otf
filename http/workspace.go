package http

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
)

// Workspace represents a Terraform Enterprise workspace.
type Workspace struct {
	ID                         string                    `jsonapi:"primary,workspaces"`
	Actions                    *otf.WorkspaceActions     `jsonapi:"attr,actions"`
	AgentPoolID                string                    `jsonapi:"attr,agent-pool-id"`
	AllowDestroyPlan           bool                      `jsonapi:"attr,allow-destroy-plan"`
	AutoApply                  bool                      `jsonapi:"attr,auto-apply"`
	CanQueueDestroyPlan        bool                      `jsonapi:"attr,can-queue-destroy-plan"`
	CreatedAt                  time.Time                 `jsonapi:"attr,created-at,iso8601"`
	Description                string                    `jsonapi:"attr,description"`
	Environment                string                    `jsonapi:"attr,environment"`
	ExecutionMode              string                    `jsonapi:"attr,execution-mode"`
	FileTriggersEnabled        bool                      `jsonapi:"attr,file-triggers-enabled"`
	GlobalRemoteState          bool                      `jsonapi:"attr,global-remote-state"`
	Locked                     bool                      `jsonapi:"attr,locked"`
	MigrationEnvironment       string                    `jsonapi:"attr,migration-environment"`
	Name                       string                    `jsonapi:"attr,name"`
	Operations                 bool                      `jsonapi:"attr,operations"`
	Permissions                *otf.WorkspacePermissions `jsonapi:"attr,permissions"`
	QueueAllRuns               bool                      `jsonapi:"attr,queue-all-runs"`
	SpeculativeEnabled         bool                      `jsonapi:"attr,speculative-enabled"`
	SourceName                 string                    `jsonapi:"attr,source-name"`
	SourceURL                  string                    `jsonapi:"attr,source-url"`
	StructuredRunOutputEnabled bool                      `jsonapi:"attr,structured-run-output-enabled"`
	TerraformVersion           string                    `jsonapi:"attr,terraform-version"`
	TriggerPrefixes            []string                  `jsonapi:"attr,trigger-prefixes"`
	VCSRepo                    *otf.VCSRepo              `jsonapi:"attr,vcs-repo"`
	WorkingDirectory           string                    `jsonapi:"attr,working-directory"`
	UpdatedAt                  time.Time                 `jsonapi:"attr,updated-at,iso8601"`
	ResourceCount              int                       `jsonapi:"attr,resource-count"`
	ApplyDurationAverage       time.Duration             `jsonapi:"attr,apply-duration-average"`
	PlanDurationAverage        time.Duration             `jsonapi:"attr,plan-duration-average"`
	PolicyCheckFailures        int                       `jsonapi:"attr,policy-check-failures"`
	RunFailures                int                       `jsonapi:"attr,run-failures"`
	RunsCount                  int                       `jsonapi:"attr,workspace-kpis-runs-count"`

	// Relations
	CurrentRun   *Run          `jsonapi:"relation,current-run"`
	Organization *Organization `jsonapi:"relation,organization"`
}

// WorkspaceList represents a list of workspaces.
type WorkspaceList struct {
	*otf.Pagination
	Items []*Workspace
}

// ToDomain converts an http obj to its domain equivalent
func (w *Workspace) ToDomain() *otf.Workspace {
	domain := otf.Workspace{
		ID:                         w.ID,
		AllowDestroyPlan:           w.AllowDestroyPlan,
		AutoApply:                  w.AutoApply,
		CanQueueDestroyPlan:        w.CanQueueDestroyPlan,
		Description:                w.Description,
		Environment:                w.Environment,
		ExecutionMode:              w.ExecutionMode,
		FileTriggersEnabled:        w.FileTriggersEnabled,
		GlobalRemoteState:          w.GlobalRemoteState,
		Locked:                     w.Locked,
		MigrationEnvironment:       w.MigrationEnvironment,
		Name:                       w.Name,
		QueueAllRuns:               w.QueueAllRuns,
		SpeculativeEnabled:         w.SpeculativeEnabled,
		SourceName:                 w.SourceName,
		SourceURL:                  w.SourceURL,
		StructuredRunOutputEnabled: w.StructuredRunOutputEnabled,
		TerraformVersion:           w.TerraformVersion,
		VCSRepo:                    w.VCSRepo,
		WorkingDirectory:           w.WorkingDirectory,
		TriggerPrefixes:            w.TriggerPrefixes,
	}

	if w.Organization != nil {
		domain.Organization = w.Organization.ToDomain()
	}

	return &domain
}

// ToDomain converts http workspace list obj to a domain workspace list obj.
func (wl *WorkspaceList) ToDomain() *otf.WorkspaceList {
	domain := otf.WorkspaceList{
		Pagination: wl.Pagination,
	}
	for _, i := range wl.Items {
		domain.Items = append(domain.Items, i.ToDomain())
	}

	return &domain
}

func (s *Server) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := otf.WorkspaceCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.WorkspaceService.Create(r.Context(), vars["org"], opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, WorkspaceJSONAPIObject(obj), WithCode(http.StatusCreated))
}

func (s *Server) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	spec := otf.WorkspaceSpecifier{
		Name:             otf.String(vars["name"]),
		OrganizationName: otf.String(vars["org"]),
	}

	obj, err := s.WorkspaceService.Get(r.Context(), spec)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, WorkspaceJSONAPIObject(obj))
}

func (s *Server) GetWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	spec := otf.WorkspaceSpecifier{
		ID: otf.String(vars["id"]),
	}

	obj, err := s.WorkspaceService.Get(r.Context(), spec)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, WorkspaceJSONAPIObject(obj))
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
	organizationName := vars["org"]
	opts.OrganizationName = &organizationName

	obj, err := s.WorkspaceService.List(r.Context(), opts)
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

	spec := otf.WorkspaceSpecifier{
		Name:             otf.String(vars["name"]),
		OrganizationName: otf.String(vars["org"]),
	}

	obj, err := s.WorkspaceService.Update(r.Context(), spec, opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, WorkspaceJSONAPIObject(obj))
}

func (s *Server) UpdateWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := otf.WorkspaceUpdateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	spec := otf.WorkspaceSpecifier{
		ID: otf.String(vars["id"]),
	}

	obj, err := s.WorkspaceService.Update(r.Context(), spec, opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, WorkspaceJSONAPIObject(obj))
}

func (s *Server) LockWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := otf.WorkspaceLockOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.WorkspaceService.Lock(r.Context(), vars["id"], opts)
	if err == otf.ErrWorkspaceAlreadyLocked {
		WriteError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, WorkspaceJSONAPIObject(obj))
}

func (s *Server) UnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := s.WorkspaceService.Unlock(r.Context(), vars["id"])
	if err == otf.ErrWorkspaceAlreadyUnlocked {
		WriteError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, WorkspaceJSONAPIObject(obj))
}

func (s *Server) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	spec := otf.WorkspaceSpecifier{
		Name:             otf.String(vars["name"]),
		OrganizationName: otf.String(vars["org"]),
	}

	if err := s.WorkspaceService.Delete(r.Context(), spec); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) DeleteWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	spec := otf.WorkspaceSpecifier{ID: otf.String(vars["id"])}

	if err := s.WorkspaceService.Delete(r.Context(), spec); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// WorkspaceJSONAPIObject converts a Workspace to a struct that can be marshalled into a
// JSON-API object
func WorkspaceJSONAPIObject(ws *otf.Workspace) *Workspace {
	obj := &Workspace{
		ID: ws.ID,
		Actions: &otf.WorkspaceActions{
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
		Permissions: &otf.WorkspacePermissions{
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
		obj.Organization = OrganizationJSONAPIObject(ws.Organization)
	}

	return obj
}

// WorkspaceListJSONAPIObject converts a WorkspaceList to
// a struct that can be marshalled into a JSON-API object
func WorkspaceListJSONAPIObject(cvl *otf.WorkspaceList) *WorkspaceList {
	obj := &WorkspaceList{
		Pagination: cvl.Pagination,
	}
	for _, item := range cvl.Items {
		obj.Items = append(obj.Items, WorkspaceJSONAPIObject(item))
	}

	return obj
}
