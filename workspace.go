package otf

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	RemoteExecutionMode ExecutionMode = "remote"
	LocalExecutionMode  ExecutionMode = "local"
	AgentExecutionMode  ExecutionMode = "agent"

	MinTerraformVersion     = "1.2.0"
	DefaultTerraformVersion = "1.3.7"
)

type ExecutionMode string

// ExecutionModePtr returns a pointer to an execution mode.
func ExecutionModePtr(m ExecutionMode) *ExecutionMode {
	return &m
}

type Workspace interface {
	ID() string
	CreatedAt() time.Time
	UpdatedAt() time.Time
	String() string
	Name() string
}

// WorkspaceList represents a list of Workspaces.
type WorkspaceList struct {
	*Pagination
	Items []Workspace
}

// WorkspaceConnector connects a workspace to a VCS repo, subscribing it to
// VCS events that trigger runs.
type WorkspaceConnector interface {
	Connect(ctx context.Context, workspaceID string, opts ConnectWorkspaceOptions) error
	Disconnect(ctx context.Context, workspaceID string) (*Workspace, error)
}

type ConnectWorkspaceOptions struct {
	Identifier string `schema:"identifier,required"` // repo id: <owner>/<repo>
	ProviderID string `schema:"vcs_provider_id,required"`
	Cloud      string // cloud host of the repo
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
	LockWorkspace(ctx context.Context, workspaceID string, opts WorkspaceLockOptions) (Workspace, error)
	UnlockWorkspace(ctx context.Context, workspaceID string, opts WorkspaceUnlockOptions) (Workspace, error)
}

// WorkspaceLockOptions represents the options for locking a workspace.
type WorkspaceLockOptions struct {
	// Specifies the reason for locking the workspace.
	Reason *string `jsonapi:"attr,reason,omitempty"`
}

// WorkspaceUnlockOptions represents the options for unlocking a workspace.
type WorkspaceUnlockOptions struct {
	// Specifies the reason for locking the workspace.
	Reason *string `jsonapi:"attr,reason,omitempty"`
	// Force unlock of workspace
	Force bool
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

// WorkspaceLockState is the state a workspace lock is currently in (i.e.
// unlocked, run-locked, or user-locked)
type WorkspaceLockState interface {
	// CanLock checks whether it can be locked by subject
	CanLock(subject Identity) error
	// CanUnlock checks whether it can be unlocked by subject
	CanUnlock(subject Identity, force bool) error
	// A lock state has an identity, i.e. the name of the run or user that has
	// locked the workspace
	Identity
}

// WorkspaceStore is a persistence store for workspaces.
type WorkspaceStore interface {
	GetWorkspaceByName(ctx context.Context, organization, workspace string) (Workspace, error)
	GetWorkspaceIDByRunID(ctx context.Context, runID string) (string, error)
	GetWorkspaceIDByStateVersionID(ctx context.Context, svID string) (string, error)
	GetWorkspaceIDByCVID(ctx context.Context, cvID string) (string, error)
	GetOrganizationNameByWorkspaceID(ctx context.Context, workspaceID string) (string, error)

	ListWorkspacePermissions(ctx context.Context, workspaceID string) ([]*WorkspacePermission, error)
}
