package otf

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/rbac"
)

const (
	RemoteExecutionMode ExecutionMode = "remote"
	LocalExecutionMode  ExecutionMode = "local"
	AgentExecutionMode  ExecutionMode = "agent"

	MinTerraformVersion     = "1.2.0"
	DefaultTerraformVersion = "1.3.7"
)

var (
	ErrWorkspaceAlreadyLocked         = errors.New("workspace already locked")
	ErrWorkspaceLockedByDifferentUser = errors.New("workspace locked by different user")
	ErrWorkspaceAlreadyUnlocked       = errors.New("workspace already unlocked")
	ErrWorkspaceUnlockDenied          = errors.New("unauthorized to unlock workspace")
	ErrWorkspaceInvalidLock           = errors.New("invalid workspace lock")
)

type ExecutionMode string

// ExecutionModePtr returns a pointer to an execution mode.
func ExecutionModePtr(m ExecutionMode) *ExecutionMode {
	return &m
}

type Workspace struct {
	ID               string
	String           string
	Name             string
	Repo             *WorkspaceRepo
	TerraformVersion string
	ExecutionMode    ExecutionMode
	AutoApply        bool
	Organization     string
	WorkingDirectory string
}

// WorkspaceResource is a resource that is associated with a workspace,
// (including a workspace), e.g. a Run, StateVersion, etc.
type WorkspaceResource interface {
	WorkspaceID() string
	Organization() string
}

// WorkspaceList represents a list of Workspaces.
type WorkspaceList struct {
	*Pagination
	Items []Workspace
}

// CreateWorkspaceOptions represents the options for creating a new workspace.
type CreateWorkspaceOptions struct {
	AllowDestroyPlan           *bool
	AutoApply                  *bool
	Description                *string
	ExecutionMode              *ExecutionMode
	FileTriggersEnabled        *bool
	GlobalRemoteState          *bool
	MigrationEnvironment       *string
	Name                       *string `schema:"name,required"`
	QueueAllRuns               *bool
	SpeculativeEnabled         *bool
	SourceName                 *string
	SourceURL                  *string
	StructuredRunOutputEnabled *bool
	TerraformVersion           *string
	TriggerPrefixes            []string
	WorkingDirectory           *string
	Organization               *string `schema:"organization_name,required"`
	Repo                       *WorkspaceRepo
}

type UpdateWorkspaceOptions struct {
	AllowDestroyPlan           *bool
	AutoApply                  *bool
	Name                       *string
	Description                *string
	ExecutionMode              *ExecutionMode `schema:"execution_mode"`
	FileTriggersEnabled        *bool
	GlobalRemoteState          *bool
	Operations                 *bool
	QueueAllRuns               *bool
	SpeculativeEnabled         *bool
	StructuredRunOutputEnabled *bool
	TerraformVersion           *string `schema:"terraform_version"`
	TriggerPrefixes            []string
	WorkingDirectory           *string
}

// WorkspaceListOptions are options for paginating and filtering a list of
// Workspaces
type WorkspaceListOptions struct {
	// Pagination
	ListOptions
	// Filter workspaces with name matching prefix.
	Prefix string `schema:"search[name],omitempty"`
	// Organization filters workspaces by organization name.
	Organization *string `schema:"organization_name,omitempty"`
	// Filter by those for which user has workspace-level permissions.
	UserID *string
}

type WorkspaceLockService interface {
	LockWorkspace(ctx context.Context, workspaceID string) (Workspace, error)
	UnlockWorkspace(ctx context.Context, workspaceID string, force bool) (Workspace, error)
}

// WorkspaceRepo represents a connection between a workspace and a VCS
// repository.
//
// TODO: rename WorkspaceConnection
type WorkspaceRepo struct {
	ProviderID  string
	WebhookID   uuid.UUID
	Identifier  string // identifier is <repo_owner>/<repo_name>
	Branch      string // branch for which applies are run
	WorkspaceID string
}

type WorkspaceService interface {
	GetWorkspace(ctx context.Context, workspaceID string) (Workspace, error)
	GetWorkspaceByName(ctx context.Context, organization, workspace string) (Workspace, error)
	GetWorkspaceJSONAPI(ctx context.Context, workspaceID string) (*jsonapi.Workspace, error)
	ListWorkspaces(ctx context.Context, opts WorkspaceListOptions) (WorkspaceList, error)
	// ListWorkspacesByWebhookID retrieves workspaces by webhook ID.
	//
	// TODO: rename to ListConnectedWorkspaces
	ListWorkspacesByWebhookID(ctx context.Context, id uuid.UUID) ([]Workspace, error)
	UpdateWorkspace(ctx context.Context, workspaceID string, opts UpdateWorkspaceOptions) (Workspace, error)
	DeleteWorkspace(ctx context.Context, workspaceID string) (Workspace, error)

	WorkspacePermissionService
}

type WorkspacePermissionService interface {
	SetWorkspacePermission(ctx context.Context, workspaceID, team string, role rbac.Role) error
	ListWorkspacePermissions(ctx context.Context, workspaceID string) ([]WorkspacePermission, error)
	UnsetWorkspacePermission(ctx context.Context, workspaceID, team string) error
}

// WorkspaceDB is a persistence store for workspaces.
type WorkspaceDB interface {
	GetWorkspace(ctx context.Context, workspaceID string) (Workspace, error)
	GetWorkspaceByName(ctx context.Context, organization, workspace string) (Workspace, error)
	GetWorkspaceIDByRunID(ctx context.Context, runID string) (string, error)
	GetWorkspaceIDByStateVersionID(ctx context.Context, svID string) (string, error)
	GetWorkspaceIDByCVID(ctx context.Context, cvID string) (string, error)
	GetOrganizationNameByWorkspaceID(ctx context.Context, workspaceID string) (string, error)

	ListWorkspacePermissions(ctx context.Context, workspaceID string) ([]WorkspacePermission, error)
}
