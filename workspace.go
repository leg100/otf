package otf

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultAllowDestroyPlan    = true
	DefaultFileTriggersEnabled = true
	DefaultTerraformVersion    = "1.3.7"

	RemoteExecutionMode ExecutionMode = "remote"
	LocalExecutionMode  ExecutionMode = "local"
	AgentExecutionMode  ExecutionMode = "agent"
)

type ExecutionMode string

func ValidateExecutionMode(m ExecutionMode) error {
	if m == RemoteExecutionMode || m == LocalExecutionMode || m == AgentExecutionMode {
		return nil
	}
	return fmt.Errorf("invalid execution mode: %s", m)
}

// ExecutionModePtr returns a pointer to an execution mode.
func ExecutionModePtr(m ExecutionMode) *ExecutionMode {
	return &m
}

var ErrInvalidWorkspaceSpec = errors.New("invalid workspace spec options")

// Workspace represents a Terraform Enterprise workspace.
type Workspace struct {
	id                         string
	createdAt                  time.Time
	updatedAt                  time.Time
	allowDestroyPlan           bool
	autoApply                  bool
	canQueueDestroyPlan        bool
	description                string
	environment                string
	executionMode              ExecutionMode
	fileTriggersEnabled        bool
	globalRemoteState          bool
	lock                       WorkspaceLockState
	migrationEnvironment       string
	name                       string
	queueAllRuns               bool
	speculativeEnabled         bool
	structuredRunOutputEnabled bool
	sourceName                 string
	sourceURL                  string
	terraformVersion           string
	triggerPrefixes            []string
	workingDirectory           string
	organization               string
	latestRunID                *string
	repo                       *WorkspaceRepo
}

func (ws *Workspace) ID() string                       { return ws.id }
func (ws *Workspace) CreatedAt() time.Time             { return ws.createdAt }
func (ws *Workspace) UpdatedAt() time.Time             { return ws.updatedAt }
func (ws *Workspace) String() string                   { return ws.organization + "/" + ws.name }
func (ws *Workspace) Name() string                     { return ws.name }
func (ws *Workspace) WorkspaceName() string            { return ws.name }
func (ws *Workspace) AllowDestroyPlan() bool           { return ws.allowDestroyPlan }
func (ws *Workspace) AutoApply() bool                  { return ws.autoApply }
func (ws *Workspace) CanQueueDestroyPlan() bool        { return ws.canQueueDestroyPlan }
func (ws *Workspace) Environment() string              { return ws.environment }
func (ws *Workspace) Description() string              { return ws.description }
func (ws *Workspace) ExecutionMode() ExecutionMode     { return ws.executionMode }
func (ws *Workspace) FileTriggersEnabled() bool        { return ws.fileTriggersEnabled }
func (ws *Workspace) GlobalRemoteState() bool          { return ws.globalRemoteState }
func (ws *Workspace) GetLock() WorkspaceLockState      { return ws.lock }
func (ws *Workspace) MigrationEnvironment() string     { return ws.migrationEnvironment }
func (ws *Workspace) QueueAllRuns() bool               { return ws.queueAllRuns }
func (ws *Workspace) SourceName() string               { return ws.sourceName }
func (ws *Workspace) SourceURL() string                { return ws.sourceURL }
func (ws *Workspace) SpeculativeEnabled() bool         { return ws.speculativeEnabled }
func (ws *Workspace) StructuredRunOutputEnabled() bool { return ws.structuredRunOutputEnabled }
func (ws *Workspace) TerraformVersion() string         { return ws.terraformVersion }
func (ws *Workspace) TriggerPrefixes() []string        { return ws.triggerPrefixes }
func (ws *Workspace) WorkingDirectory() string         { return ws.workingDirectory }
func (ws *Workspace) Organization() string             { return ws.organization }
func (ws *Workspace) LatestRunID() *string             { return ws.latestRunID }
func (ws *Workspace) Repo() *WorkspaceRepo             { return ws.repo }

// ExecutionModes returns a list of possible execution modes
func (ws *Workspace) ExecutionModes() []string {
	return []string{"local", "remote", "agent"}
}

// QualifiedName returns the workspace's qualified name including the name of
// its organization
func (ws *Workspace) QualifiedName() WorkspaceQualifiedName {
	return WorkspaceQualifiedName{
		Organization: ws.Organization(),
		Name:         ws.Name(),
	}
}

// Locked determines whether workspace is locked.
func (ws *Workspace) Locked() bool {
	_, ok := ws.lock.(*Unlocked)
	return !ok
}

// Lock transfers a workspace into the given lock state
func (ws *Workspace) Lock(lock WorkspaceLockState) error {
	if err := ws.lock.CanLock(lock); err != nil {
		return err
	}
	ws.lock = lock
	return nil
}

// Unlock the workspace using the given identity.
func (ws *Workspace) Unlock(iden Identity, force bool) error {
	if err := ws.lock.CanUnlock(iden, force); err != nil {
		return err
	}
	ws.lock = &Unlocked{}
	return nil
}

func (ws *Workspace) MarshalLog() any {
	log := struct {
		Name         string `json:"name"`
		Organization string `json:"organization"`
	}{
		Name:         ws.name,
		Organization: ws.organization,
	}
	return log
}

// UpdateWithOptions updates the workspace with the given options.
//
// TODO: validate options
func (ws *Workspace) UpdateWithOptions(ctx context.Context, opts WorkspaceUpdateOptions) error {
	if opts.Name != nil {
		ws.name = *opts.Name
		ws.updatedAt = CurrentTimestamp()
	}
	if opts.AllowDestroyPlan != nil {
		ws.allowDestroyPlan = *opts.AllowDestroyPlan
		ws.updatedAt = CurrentTimestamp()
	}
	if opts.AutoApply != nil {
		ws.autoApply = *opts.AutoApply
		ws.updatedAt = CurrentTimestamp()
	}
	if opts.Description != nil {
		ws.description = *opts.Description
		ws.updatedAt = CurrentTimestamp()
	}
	if opts.ExecutionMode != nil {
		if err := ValidateExecutionMode(*opts.ExecutionMode); err != nil {
			return err
		}
		ws.executionMode = *opts.ExecutionMode
		ws.updatedAt = CurrentTimestamp()
	}
	if opts.FileTriggersEnabled != nil {
		ws.fileTriggersEnabled = *opts.FileTriggersEnabled
		ws.updatedAt = CurrentTimestamp()
	}
	if opts.Operations != nil {
		if *opts.Operations {
			ws.executionMode = "remote"
		} else {
			ws.executionMode = "local"
		}
		ws.updatedAt = CurrentTimestamp()
	}
	if opts.QueueAllRuns != nil {
		ws.queueAllRuns = *opts.QueueAllRuns
		ws.updatedAt = CurrentTimestamp()
	}
	if opts.SpeculativeEnabled != nil {
		ws.speculativeEnabled = *opts.SpeculativeEnabled
		ws.updatedAt = CurrentTimestamp()
	}
	if opts.StructuredRunOutputEnabled != nil {
		ws.structuredRunOutputEnabled = *opts.StructuredRunOutputEnabled
		ws.updatedAt = CurrentTimestamp()
	}
	if opts.TerraformVersion != nil {
		ws.terraformVersion = *opts.TerraformVersion
		ws.updatedAt = CurrentTimestamp()
	}
	if opts.TriggerPrefixes != nil {
		ws.triggerPrefixes = opts.TriggerPrefixes
		ws.updatedAt = CurrentTimestamp()
	}
	if opts.WorkingDirectory != nil {
		ws.workingDirectory = *opts.WorkingDirectory
		ws.updatedAt = CurrentTimestamp()
	}
	if opts.WorkspaceRepo != nil {
		if ws.repo != nil {
			return fmt.Errorf("updating workspace vcs repo not supported")
		}
		ws.repo = opts.WorkspaceRepo
	}

	return nil
}

// WorkspaceQualifiedName is the workspace's fully qualified name including the
// name of its organization
type WorkspaceQualifiedName struct {
	Organization string
	Name         string
}

// WorkspaceUpdateOptions represents the options for updating a workspace.
type WorkspaceUpdateOptions struct {
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
	*WorkspaceRepo
}

func (o WorkspaceUpdateOptions) Valid() error {
	if o.AllowDestroyPlan == nil &&
		o.AutoApply == nil &&
		o.Name == nil &&
		o.Description == nil &&
		o.ExecutionMode == nil &&
		o.FileTriggersEnabled == nil &&
		o.GlobalRemoteState == nil &&
		o.Operations == nil &&
		o.QueueAllRuns == nil &&
		o.SpeculativeEnabled == nil &&
		o.StructuredRunOutputEnabled == nil &&
		o.TerraformVersion == nil &&
		o.TriggerPrefixes == nil &&
		o.WorkingDirectory == nil &&
		o.WorkspaceRepo == nil {
		return fmt.Errorf("must set at least one option to update")
	}
	if o.Name != nil && !ValidStringID(o.Name) {
		return ErrInvalidName
	}
	if o.TerraformVersion != nil && !validSemanticVersion(*o.TerraformVersion) {
		return ErrInvalidTerraformVersion
	}
	if o.ExecutionMode != nil {
		if err := ValidateExecutionMode(*o.ExecutionMode); err != nil {
			return err
		}
	}
	return nil
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

// WorkspaceList represents a list of Workspaces.
type WorkspaceList struct {
	*Pagination
	Items []*Workspace
}

type WorkspaceService interface {
	CreateWorkspace(ctx context.Context, opts WorkspaceCreateOptions) (*Workspace, error)
	GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error)
	GetWorkspaceByName(ctx context.Context, organization, workspace string) (*Workspace, error)
	ListWorkspaces(ctx context.Context, opts WorkspaceListOptions) (*WorkspaceList, error)
	ListWorkspacesByWebhookID(ctx context.Context, id uuid.UUID) ([]*Workspace, error)
	UpdateWorkspace(ctx context.Context, workspaceID string, opts WorkspaceUpdateOptions) (*Workspace, error)
	DeleteWorkspace(ctx context.Context, workspaceID string) (*Workspace, error)

	WorkspaceLockService
	CurrentRunService
	WorkspacePermissionService
	WorkspaceConnectionService
}

type WorkspaceConnectionService interface {
	ConnectWorkspace(ctx context.Context, workspaceID string, opts ConnectWorkspaceOptions) (*Workspace, error)
	UpdateWorkspaceRepo(ctx context.Context, workspaceID string, repo WorkspaceRepo) (*Workspace, error)
	DisconnectWorkspace(ctx context.Context, workspaceID string) (*Workspace, error)
}

type WorkspacePermissionService interface {
	SetWorkspacePermission(ctx context.Context, workspaceID, team string, role Role) error
	ListWorkspacePermissions(ctx context.Context, workspaceID string) ([]*WorkspacePermission, error)
	UnsetWorkspacePermission(ctx context.Context, workspaceID, team string) error
}

type WorkspaceLockService interface {
	LockWorkspace(ctx context.Context, workspaceID string, opts WorkspaceLockOptions) (*Workspace, error)
	UnlockWorkspace(ctx context.Context, workspaceID string, opts WorkspaceUnlockOptions) (*Workspace, error)
}

// WorkspaceStore is a persistence store for workspaces.
type WorkspaceStore interface {
	CreateWorkspace(ctx context.Context, ws *Workspace) error
	GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error)
	GetWorkspaceByName(ctx context.Context, organization, workspace string) (*Workspace, error)
	ListWorkspaces(ctx context.Context, opts WorkspaceListOptions) (*WorkspaceList, error)
	ListWorkspacesByUserID(ctx context.Context, userID string, organization string, opts ListOptions) (*WorkspaceList, error)
	ListWorkspacesByWebhookID(ctx context.Context, id uuid.UUID) ([]*Workspace, error)
	UpdateWorkspace(ctx context.Context, workspaceID string, ws func(ws *Workspace) error) (*Workspace, error)
	DeleteWorkspace(ctx context.Context, workspaceID string) error
	GetWorkspaceIDByRunID(ctx context.Context, runID string) (string, error)
	GetWorkspaceIDByStateVersionID(ctx context.Context, svID string) (string, error)
	GetWorkspaceIDByCVID(ctx context.Context, cvID string) (string, error)

	// CreateWorkspaceRepo creates a workspace repo in the persistence store.
	CreateWorkspaceRepo(ctx context.Context, workspaceID string, repo WorkspaceRepo) (*Workspace, error)
	// UpdateWorkspaceRepo updates a workspace's repo in the persistence store.
	UpdateWorkspaceRepo(ctx context.Context, workspaceID string, repo WorkspaceRepo) (*Workspace, error)
	// DeleteWorkspaceRepo deletes a workspace's repo from the persistence
	// store, returning the workspace without the repo as well the original repo, or an
	// error.
	DeleteWorkspaceRepo(ctx context.Context, workspaceID string) (*Workspace, error)

	WorkspaceLockService
	CurrentRunService
	WorkspacePermissionService
}

// CurrentRunService provides interaction with the current run for a workspace,
// i.e. the current, or most recently current, non-speculative, run.
type CurrentRunService interface {
	// SetCurrentRun sets the ID of the latest run for a workspace.
	//
	// Take full run obj as param
	SetCurrentRun(ctx context.Context, workspaceID, runID string) error
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
