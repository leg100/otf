package ots

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/jsonapi"
	tfe "github.com/hashicorp/go-tfe"
)

const (
	DefaultAllowDestroyPlan    = true
	DefaultFileTriggersEnabled = true
	DefaultTerraformVersion    = "0.15.4"
)

var (
	ErrWorkspaceAlreadyLocked   = errors.New("workspace already locked")
	ErrWorkspaceAlreadyUnlocked = errors.New("workspace already unlocked")
)

// Workspace represents a Terraform Enterprise workspace.
type Workspace struct {
	ID                   string                `jsonapi:"primary,workspaces"`
	Actions              *WorkspaceActions     `jsonapi:"attr,actions"`
	AgentPoolID          string                `jsonapi:"attr,agent-pool-id"`
	AllowDestroyPlan     bool                  `jsonapi:"attr,allow-destroy-plan"`
	AutoApply            bool                  `jsonapi:"attr,auto-apply"`
	CanQueueDestroyPlan  bool                  `jsonapi:"attr,can-queue-destroy-plan"`
	CreatedAt            time.Time             `jsonapi:"attr,created-at,iso8601"`
	Description          string                `jsonapi:"attr,description"`
	Environment          string                `jsonapi:"attr,environment"`
	ExecutionMode        string                `jsonapi:"attr,execution-mode"`
	FileTriggersEnabled  bool                  `jsonapi:"attr,file-triggers-enabled"`
	GlobalRemoteState    bool                  `jsonapi:"attr,global-remote-state"`
	Locked               bool                  `jsonapi:"attr,locked"`
	MigrationEnvironment string                `jsonapi:"attr,migration-environment"`
	Name                 string                `jsonapi:"attr,name"`
	Operations           bool                  `jsonapi:"attr,operations"`
	Permissions          *WorkspacePermissions `jsonapi:"attr,permissions"`
	QueueAllRuns         bool                  `jsonapi:"attr,queue-all-runs"`
	SpeculativeEnabled   bool                  `jsonapi:"attr,speculative-enabled"`
	TerraformVersion     string                `jsonapi:"attr,terraform-version"`
	TriggerPrefixes      []string              `jsonapi:"attr,trigger-prefixes"`
	VCSRepo              *VCSRepo              `jsonapi:"attr,vcs-repo"`
	WorkingDirectory     string                `jsonapi:"attr,working-directory"`
	UpdatedAt            time.Time             `jsonapi:"attr,updated-at,iso8601"`
	ResourceCount        int                   `jsonapi:"attr,resource-count"`
	ApplyDurationAverage time.Duration         `jsonapi:"attr,apply-duration-average"`
	PlanDurationAverage  time.Duration         `jsonapi:"attr,plan-duration-average"`
	PolicyCheckFailures  int                   `jsonapi:"attr,policy-check-failures"`
	RunFailures          int                   `jsonapi:"attr,run-failures"`
	RunsCount            int                   `jsonapi:"attr,workspace-kpis-runs-count"`

	// Relations
	Organization *Organization `jsonapi:"relation,organization"`
}

// WorkspaceActions represents the workspace actions.
type WorkspaceActions struct {
	IsDestroyable bool `json:"is-destroyable"`
}

// WorkspacePermissions represents the workspace permissions.
type WorkspacePermissions struct {
	CanDestroy        bool `json:"can-destroy"`
	CanForceUnlock    bool `json:"can-force-unlock"`
	CanLock           bool `json:"can-lock"`
	CanQueueApply     bool `json:"can-queue-apply"`
	CanQueueDestroy   bool `json:"can-queue-destroy"`
	CanQueueRun       bool `json:"can-queue-run"`
	CanReadSettings   bool `json:"can-read-settings"`
	CanUnlock         bool `json:"can-unlock"`
	CanUpdate         bool `json:"can-update"`
	CanUpdateVariable bool `json:"can-update-variable"`
}

// VCSRepo contains the configuration of a VCS integration.
type VCSRepo struct {
	Branch            string `json:"branch"`
	DisplayIdentifier string `json:"display-identifier"`
	Identifier        string `json:"identifier"`
	IngressSubmodules bool   `json:"ingress-submodules"`
	OAuthTokenID      string `json:"oauth-token-id"`
	RepositoryHTTPURL string `json:"repository-http-url"`
	ServiceProvider   string `json:"service-provider"`
}

type WorkspaceList struct {
	Items        []*Workspace
	Organization string
	WorkspaceListOptions
}

// WorkspaceListOptions represents the options for listing workspaces.
type WorkspaceListOptions struct {
	ListOptions

	// A search string (partial workspace name) used to filter the results.
	Search *string `schema:"search[name]"`

	// A list of relations to include. See available resources https://www.terraform.io/docs/cloud/api/workspaces.html#available-related-resources
	Include *string `schema:"include"`
}

func (l *WorkspaceList) GetItems() interface{} {
	return l.Items
}

func (l *WorkspaceList) GetPath() string {
	return fmt.Sprintf("/api/v2/organizations/%s/workspaces", l.Organization)
}

var _ Paginated = (*WorkspaceList)(nil)

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

	// The version of Terraform to use for this workspace. Upon creating a
	// workspace, the latest version is selected unless otherwise specified.
	TerraformVersion *string `jsonapi:"attr,terraform-version,omitempty"`

	// List of repository-root-relative paths which list all locations to be
	// tracked for changes. See FileTriggersEnabled above for more details.
	TriggerPrefixes []string `jsonapi:"attr,trigger-prefixes,omitempty"`

	// Settings for the workspace's VCS repository. If omitted, the workspace is
	// created without a VCS repo. If included, you must specify at least the
	// oauth-token-id and identifier keys below.
	VCSRepo *tfe.VCSRepoOptions `jsonapi:"attr,vcs-repo,omitempty"`

	// A relative path that Terraform will execute within. This defaults to the
	// root of your repository and is typically set to a subdirectory matching the
	// environment when multiple environments exist within the same repository.
	WorkingDirectory *string `jsonapi:"attr,working-directory,omitempty"`
}

// WorkspaceLockOptions represents the options for locking a workspace.
type WorkspaceLockOptions struct {
	// Specifies the reason for locking the workspace.
	Reason *string `jsonapi:"attr,reason,omitempty"`
}

type WorkspaceService interface {
	CreateWorkspace(org string, opts *WorkspaceCreateOptions) (*Workspace, error)
	GetWorkspace(name, org string) (*Workspace, error)
	GetWorkspaceByID(id string) (*Workspace, error)
	ListWorkspaces(org string, opts WorkspaceListOptions) (*WorkspaceList, error)
	UpdateWorkspace(name, org string, opts *tfe.WorkspaceUpdateOptions) (*Workspace, error)
	UpdateWorkspaceByID(id string, opts *tfe.WorkspaceUpdateOptions) (*Workspace, error)
	DeleteWorkspace(name, org string) error
	DeleteWorkspaceByID(id string) error
	LockWorkspace(id string, opts WorkspaceLockOptions) (*Workspace, error)
	UnlockWorkspace(id string) (*Workspace, error)
}

func (ws *Workspace) JSONAPILinks() *jsonapi.Links {
	return &jsonapi.Links{
		"self": fmt.Sprintf("/api/v2/organizations/%s/workspaces/%s", ws.Organization.Name, ws.Name),
	}
}

func (opts *WorkspaceCreateOptions) Validate() error {
	if !validString(opts.Name) {
		return tfe.ErrRequiredName
	}
	if !validStringID(opts.Name) {
		return tfe.ErrInvalidName
	}
	if opts.AgentPoolID != nil {
		return errors.New("unsupported")
	}

	return nil
}

func NewWorkspaceID() string {
	return fmt.Sprintf("ws-%s", GenerateRandomString(16))
}
