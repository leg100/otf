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

// WorkspaceCreateOptions represents the options for creating a new workspace.
type WorkspaceCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,workspaces"`

	// Required when execution-mode is set to agent. The ID of the agent pool
	// belonging to the workspace's organization. This value must not be specified
	// if execution-mode is set to remote or local or if operations is set to true.
	AgentPoolID *string `jsonapi:"attr,agent-pool-id,omitempty"`

	// Whether destroy plans can be queued on the workspace.
	AllowDestroyPlan *bool `jsonapi:"attr,allow-destroy-plan,omitempty"`

	// Whether to automatically apply changes when a Terraform plan is successful.
	AutoApply *bool `jsonapi:"attr,auto-apply,omitempty"`

	// A description for the workspace.
	Description *string `jsonapi:"attr,description,omitempty"`

	// Which execution mode to use. Valid values are remote, local, and agent.
	// When set to local, the workspace will be used for state storage only.
	// This value must not be specified if operations is specified.
	// 'agent' execution mode is not available in Terraform Enterprise.
	ExecutionMode *string `jsonapi:"attr,execution-mode,omitempty"`

	// Whether to filter runs based on the changed files in a VCS push. If
	// enabled, the working directory and trigger prefixes describe a set of
	// paths which must contain changes for a VCS push to trigger a run. If
	// disabled, any push will trigger a run.
	FileTriggersEnabled *bool `jsonapi:"attr,file-triggers-enabled,omitempty"`

	GlobalRemoteState *bool `jsonapi:"attr,global-remote-state,omitempty"`

	// The legacy TFE environment to use as the source of the migration, in the
	// form organization/environment. Omit this unless you are migrating a legacy
	// environment.
	MigrationEnvironment *string `jsonapi:"attr,migration-environment,omitempty"`

	// The name of the workspace, which can only include letters, numbers, -,
	// and _. This will be used as an identifier and must be unique in the
	// organization.
	Name *string `jsonapi:"attr,name"`

	// DEPRECATED. Whether the workspace will use remote or local execution mode.
	// Use ExecutionMode instead.
	Operations *bool `jsonapi:"attr,operations,omitempty"`

	// Whether to queue all runs. Unless this is set to true, runs triggered by
	// a webhook will not be queued until at least one run is manually queued.
	QueueAllRuns *bool `jsonapi:"attr,queue-all-runs,omitempty"`

	// Whether this workspace allows speculative plans. Setting this to false
	// prevents Terraform Cloud or the Terraform Enterprise instance from
	// running plans on pull requests, which can improve security if the VCS
	// repository is public or includes untrusted contributors.
	SpeculativeEnabled *bool `jsonapi:"attr,speculative-enabled,omitempty"`

	// BETA. A friendly name for the application or client creating this
	// workspace. If set, this will be displayed on the workspace as
	// "Created via <SOURCE NAME>".
	SourceName *string `jsonapi:"attr,source-name,omitempty"`

	// BETA. A URL for the application or client creating this workspace. This
	// can be the URL of a related resource in another app, or a link to
	// documentation or other info about the client.
	SourceURL *string `jsonapi:"attr,source-url,omitempty"`

	// BETA. Enable the experimental advanced run user interface.
	// This only applies to runs using Terraform version 0.15.2 or newer,
	// and runs executed using older versions will see the classic experience
	// regardless of this setting.
	StructuredRunOutputEnabled *bool `jsonapi:"attr,structured-run-output-enabled,omitempty"`

	// The version of Terraform to use for this workspace. Upon creating a
	// workspace, the latest version is selected unless otherwise specified.
	TerraformVersion *string `jsonapi:"attr,terraform-version,omitempty"`

	// List of repository-root-relative paths which list all locations to be
	// tracked for changes. See FileTriggersEnabled above for more details.
	TriggerPrefixes []string `jsonapi:"attr,trigger-prefixes,omitempty"`

	// Settings for the workspace's VCS repository. If omitted, the workspace is
	// created without a VCS repo. If included, you must specify at least the
	// oauth-token-id and identifier keys below.
	VCSRepo *VCSRepoOptions `jsonapi:"attr,vcs-repo,omitempty"`

	// A relative path that Terraform will execute within. This defaults to the
	// root of your repository and is typically set to a subdirectory matching the
	// environment when multiple environments exist within the same repository.
	WorkingDirectory *string `jsonapi:"attr,working-directory,omitempty"`
}

// WorkspaceUpdateOptions represents the options for updating a workspace.
type WorkspaceUpdateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,workspaces"`

	// Required when execution-mode is set to agent. The ID of the agent pool
	// belonging to the workspace's organization. This value must not be specified
	// if execution-mode is set to remote or local or if operations is set to true.
	AgentPoolID *string `jsonapi:"attr,agent-pool-id,omitempty"`

	// Whether destroy plans can be queued on the workspace.
	AllowDestroyPlan *bool `jsonapi:"attr,allow-destroy-plan,omitempty"`

	// Whether to automatically apply changes when a Terraform plan is successful.
	AutoApply *bool `jsonapi:"attr,auto-apply,omitempty"`

	// A new name for the workspace, which can only include letters, numbers, -,
	// and _. This will be used as an identifier and must be unique in the
	// organization. Warning: Changing a workspace's name changes its URL in the
	// API and UI.
	Name *string `jsonapi:"attr,name,omitempty"`

	// A description for the workspace.
	Description *string `jsonapi:"attr,description,omitempty"`

	// Which execution mode to use. Valid values are remote, local, and agent.
	// When set to local, the workspace will be used for state storage only.
	// This value must not be specified if operations is specified.
	// 'agent' execution mode is not available in Terraform Enterprise.
	ExecutionMode *string `jsonapi:"attr,execution-mode,omitempty"`

	// Whether to filter runs based on the changed files in a VCS push. If
	// enabled, the working directory and trigger prefixes describe a set of
	// paths which must contain changes for a VCS push to trigger a run. If
	// disabled, any push will trigger a run.
	FileTriggersEnabled *bool `jsonapi:"attr,file-triggers-enabled,omitempty"`

	GlobalRemoteState *bool `jsonapi:"attr,global-remote-state,omitempty"`

	// DEPRECATED. Whether the workspace will use remote or local execution mode.
	// Use ExecutionMode instead.
	Operations *bool `jsonapi:"attr,operations,omitempty"`

	// Whether to queue all runs. Unless this is set to true, runs triggered by
	// a webhook will not be queued until at least one run is manually queued.
	QueueAllRuns *bool `jsonapi:"attr,queue-all-runs,omitempty"`

	// Whether this workspace allows speculative plans. Setting this to false
	// prevents Terraform Cloud or the Terraform Enterprise instance from
	// running plans on pull requests, which can improve security if the VCS
	// repository is public or includes untrusted contributors.
	SpeculativeEnabled *bool `jsonapi:"attr,speculative-enabled,omitempty"`

	// BETA. Enable the experimental advanced run user interface.
	// This only applies to runs using Terraform version 0.15.2 or newer,
	// and runs executed using older versions will see the classic experience
	// regardless of this setting.
	StructuredRunOutputEnabled *bool `jsonapi:"attr,structured-run-output-enabled,omitempty"`

	// The version of Terraform to use for this workspace.
	TerraformVersion *string `jsonapi:"attr,terraform-version,omitempty"`

	// List of repository-root-relative paths which list all locations to be
	// tracked for changes. See FileTriggersEnabled above for more details.
	TriggerPrefixes []string `jsonapi:"attr,trigger-prefixes,omitempty"`

	// To delete a workspace's existing VCS repo, specify null instead of an
	// object. To modify a workspace's existing VCS repo, include whichever of
	// the keys below you wish to modify. To add a new VCS repo to a workspace
	// that didn't previously have one, include at least the oauth-token-id and
	// identifier keys.
	VCSRepo *VCSRepoOptions `jsonapi:"attr,vcs-repo,omitempty"`

	// A relative path that Terraform will execute within. This defaults to the
	// root of your repository and is typically set to a subdirectory matching
	// the environment when multiple environments exist within the same
	// repository.
	WorkingDirectory *string `jsonapi:"attr,working-directory,omitempty"`
}

// WorkspaceList represents a list of workspaces.
type WorkspaceList struct {
	*otf.Pagination
	Items []*Workspace
}

func (o *WorkspaceCreateOptions) ToDomain() otf.WorkspaceCreateOptions {
	return otf.WorkspaceCreateOptions{
		AllowDestroyPlan:           o.AllowDestroyPlan,
		AutoApply:                  o.AutoApply,
		Description:                o.Description,
		ExecutionMode:              o.ExecutionMode,
		FileTriggersEnabled:        o.FileTriggersEnabled,
		GlobalRemoteState:          o.GlobalRemoteState,
		Name:                       o.Name,
		Operations:                 o.Operations,
		QueueAllRuns:               o.QueueAllRuns,
		SpeculativeEnabled:         o.SpeculativeEnabled,
		SourceName:                 o.SourceName,
		SourceURL:                  o.SourceURL,
		StructuredRunOutputEnabled: o.StructuredRunOutputEnabled,
		TerraformVersion:           o.TerraformVersion,
		TriggerPrefixes:            o.TriggerPrefixes,
		WorkingDirectory:           o.WorkingDirectory,
	}
}

func (o *WorkspaceUpdateOptions) ToDomain() otf.WorkspaceUpdateOptions {
	return otf.WorkspaceUpdateOptions{
		AllowDestroyPlan:           o.AllowDestroyPlan,
		AutoApply:                  o.AutoApply,
		Description:                o.Description,
		ExecutionMode:              o.ExecutionMode,
		FileTriggersEnabled:        o.FileTriggersEnabled,
		GlobalRemoteState:          o.GlobalRemoteState,
		Name:                       o.Name,
		Operations:                 o.Operations,
		QueueAllRuns:               o.QueueAllRuns,
		SpeculativeEnabled:         o.SpeculativeEnabled,
		StructuredRunOutputEnabled: o.StructuredRunOutputEnabled,
		TerraformVersion:           o.TerraformVersion,
		TriggerPrefixes:            o.TriggerPrefixes,
		WorkingDirectory:           o.WorkingDirectory,
	}
}

// ToDomain converts http workspace obj to a domain workspace obj.
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
		Permissions:                w.Permissions,
		QueueAllRuns:               w.QueueAllRuns,
		SpeculativeEnabled:         w.SpeculativeEnabled,
		SourceName:                 w.SourceName,
		SourceURL:                  w.SourceURL,
		StructuredRunOutputEnabled: w.StructuredRunOutputEnabled,
		TerraformVersion:           w.TerraformVersion,
		VCSRepo:                    w.VCSRepo,
		WorkingDirectory:           w.WorkingDirectory,
		ResourceCount:              w.ResourceCount,
		ApplyDurationAverage:       w.ApplyDurationAverage,
		PlanDurationAverage:        w.PlanDurationAverage,
		PolicyCheckFailures:        w.PolicyCheckFailures,
		RunFailures:                w.RunFailures,
		RunsCount:                  w.RunsCount,
		TriggerPrefixes:            w.TriggerPrefixes,
	}

	if w.Organization != nil {
		domain.Organization = w.Organization.ToDomain()
	}

	return &domain
}

func (s *Server) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := WorkspaceCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.WorkspaceService.Create(vars["org"], opts.ToDomain())
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.WorkspaceJSONAPIObject(obj), WithCode(http.StatusCreated))
}

func (s *Server) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	spec := otf.WorkspaceSpecifier{
		Name:             otf.String(vars["name"]),
		OrganizationName: otf.String(vars["org"]),
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
	spec := otf.WorkspaceSpecifier{
		ID: otf.String(vars["id"]),
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
	var opts otf.WorkspaceListOptions
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

	opts := WorkspaceUpdateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	spec := otf.WorkspaceSpecifier{
		Name:             otf.String(vars["name"]),
		OrganizationName: otf.String(vars["org"]),
	}

	obj, err := s.WorkspaceService.Update(spec, opts.ToDomain())
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.WorkspaceJSONAPIObject(obj))
}

func (s *Server) UpdateWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := WorkspaceUpdateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	spec := otf.WorkspaceSpecifier{
		ID: otf.String(vars["id"]),
	}

	obj, err := s.WorkspaceService.Update(spec, opts.ToDomain())
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.WorkspaceJSONAPIObject(obj))
}

func (s *Server) LockWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	opts := otf.WorkspaceLockOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.WorkspaceService.Lock(vars["id"], opts)
	if err == otf.ErrWorkspaceAlreadyLocked {
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
	if err == otf.ErrWorkspaceAlreadyUnlocked {
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
	spec := otf.WorkspaceSpecifier{
		Name:             otf.String(vars["name"]),
		OrganizationName: otf.String(vars["org"]),
	}

	if err := s.WorkspaceService.Delete(spec); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) DeleteWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	spec := otf.WorkspaceSpecifier{ID: otf.String(vars["id"])}

	if err := s.WorkspaceService.Delete(spec); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// WorkspaceJSONAPIObject converts a Workspace to a struct that can be marshalled into a
// JSON-API object
func (s *Server) WorkspaceJSONAPIObject(ws *otf.Workspace) *Workspace {
	obj := &Workspace{
		ID: ws.ID,
		Actions: &otf.WorkspaceActions{
			IsDestroyable: false,
		},
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

	if ws.ExecutionMode == "remote" {
		// Operations is deprecated but clients and go-tfe tests still use it
		obj.Operations = true
	}

	if ws.Organization != nil {
		obj.Organization = s.OrganizationJSONAPIObject(ws.Organization)
	}

	return obj
}

// WorkspaceListJSONAPIObject converts a WorkspaceList to
// a struct that can be marshalled into a JSON-API object
func (s *Server) WorkspaceListJSONAPIObject(cvl *otf.WorkspaceList) *WorkspaceList {
	obj := &WorkspaceList{
		Pagination: cvl.Pagination,
	}
	for _, item := range cvl.Items {
		obj.Items = append(obj.Items, s.WorkspaceJSONAPIObject(item))
	}

	return obj
}
