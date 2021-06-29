package ots

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/jsonapi"
	tfe "github.com/leg100/go-tfe"
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
	Organization *tfe.Organization `jsonapi:"relation,organization"`
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
	*Pagination
	Items []*Workspace
}

// WorkspaceListOptions represents the options for listing workspaces.
type WorkspaceListOptions struct {
	ListOptions

	// A search string (partial workspace name) used to filter the results.
	Search *string `schema:"search[name]"`

	// A list of relations to include. See available resources https://www.terraform.io/docs/cloud/api/workspaces.html#available-related-resources
	Include *string `schema:"include"`
}

// WorkspaceLockOptions represents the options for locking a workspace.
type WorkspaceLockOptions struct {
	// Specifies the reason for locking the workspace.
	Reason *string `jsonapi:"attr,reason,omitempty"`
}

type WorkspaceService interface {
	CreateWorkspace(org string, opts *tfe.WorkspaceCreateOptions) (*Workspace, error)
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

func NewWorkspaceID() string {
	return fmt.Sprintf("ws-%s", GenerateRandomString(16))
}
