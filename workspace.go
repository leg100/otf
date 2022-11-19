package otf

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	DefaultAllowDestroyPlan    = true
	DefaultFileTriggersEnabled = true
	DefaultTerraformVersion    = "1.0.10"

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
	organization               *Organization
	latestRunID                *string
	repo                       *VCSRepo
}

func (ws *Workspace) ID() string                       { return ws.id }
func (ws *Workspace) CreatedAt() time.Time             { return ws.createdAt }
func (ws *Workspace) UpdatedAt() time.Time             { return ws.updatedAt }
func (ws *Workspace) String() string                   { return ws.organization.Name() + "/" + ws.name }
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
func (ws *Workspace) OrganizationID() string           { return ws.organization.id }
func (ws *Workspace) OrganizationName() string         { return ws.organization.name }
func (ws *Workspace) Organization() *Organization      { return ws.organization }
func (ws *Workspace) LatestRunID() *string             { return ws.latestRunID }
func (ws *Workspace) VCSRepo() *VCSRepo                { return ws.repo }

// ExecutionModes returns a list of possible execution modes
func (ws *Workspace) ExecutionModes() []string {
	return []string{"local", "remote", "agent"}
}

// QualifiedName returns the workspace's qualified name including the name of
// its organization
func (ws *Workspace) QualifiedName() WorkspaceQualifiedName {
	return WorkspaceQualifiedName{
		Organization: ws.OrganizationName(),
		Name:         ws.Name(),
	}
}

func (ws *Workspace) SpecID() WorkspaceSpec {
	return WorkspaceSpec{ID: &ws.id}
}

func (ws *Workspace) SpecName() WorkspaceSpec {
	return WorkspaceSpec{Name: &ws.name, OrganizationName: &ws.organization.name}
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
	if opts.VCSRepo != nil {
		if ws.repo != nil {
			return fmt.Errorf("updating workspace vcs repo not supported")
		}
		ws.repo = opts.VCSRepo
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
	*VCSRepo
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
		o.VCSRepo == nil {
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
	GetWorkspace(ctx context.Context, spec WorkspaceSpec) (*Workspace, error)
	ListWorkspaces(ctx context.Context, opts WorkspaceListOptions) (*WorkspaceList, error)
	UpdateWorkspace(ctx context.Context, spec WorkspaceSpec, opts WorkspaceUpdateOptions) (*Workspace, error)
	DeleteWorkspace(ctx context.Context, spec WorkspaceSpec) error

	WorkspaceLockService
	CurrentRunService
	WorkspacePermissionService
	WorkspaceRepoService
}

type WorkspacePermissionService interface {
	SetWorkspacePermission(ctx context.Context, spec WorkspaceSpec, team string, role WorkspaceRole) error
	ListWorkspacePermissions(ctx context.Context, spec WorkspaceSpec) ([]*WorkspacePermission, error)
	UnsetWorkspacePermission(ctx context.Context, spec WorkspaceSpec, team string) error
}

type WorkspaceLockService interface {
	LockWorkspace(ctx context.Context, spec WorkspaceSpec, opts WorkspaceLockOptions) (*Workspace, error)
	UnlockWorkspace(ctx context.Context, spec WorkspaceSpec, opts WorkspaceUnlockOptions) (*Workspace, error)
}

// WorkspaceStore is a persistence store for workspaces.
type WorkspaceStore interface {
	CreateWorkspace(ctx context.Context, ws *Workspace) error
	GetWorkspace(ctx context.Context, spec WorkspaceSpec) (*Workspace, error)
	ListWorkspaces(ctx context.Context, opts WorkspaceListOptions) (*WorkspaceList, error)
	ListWorkspacesByUserID(ctx context.Context, userID string, organization string, opts ListOptions) (*WorkspaceList, error)
	UpdateWorkspace(ctx context.Context, spec WorkspaceSpec, ws func(ws *Workspace) error) (*Workspace, error)
	DeleteWorkspace(ctx context.Context, spec WorkspaceSpec) error
	GetWorkspaceID(ctx context.Context, spec WorkspaceSpec) (string, error)
	GetWorkspaceIDByRunID(ctx context.Context, runID string) (string, error)
	GetWorkspaceIDByStateVersionID(ctx context.Context, svID string) (string, error)
	GetWorkspaceIDByCVID(ctx context.Context, cvID string) (string, error)

	WorkspaceLockService
	CurrentRunService
	WorkspacePermissionService
	WorkspaceRepoService
}

// WorkspaceRepoService manages a workspace's connection to a VCS repository.
type WorkspaceRepoService interface {
	// ConnectWorkspaceRepo connects a workspace to a VCS repository using a VCS
	// provider.
	//
	// TODO: Rename to ConnectWorkspace
	ConnectWorkspaceRepo(ctx context.Context, spec WorkspaceSpec, repo VCSRepo) (*Workspace, error)
	UpdateWorkspaceRepo(ctx context.Context, spec WorkspaceSpec, repo VCSRepo) (*Workspace, error)
	DisconnectWorkspaceRepo(ctx context.Context, spec WorkspaceSpec) (*Workspace, error)
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
	// OrganizationName filters workspaces by organization name.
	OrganizationName *string `schema:"organization_name,omitempty"`
	// Filter by those for which user has workspace-level permissions.
	UserID *string
}

// WorkspaceSpec is used for identifying an individual workspace. Either ID *or*
// both Name and OrganizationName must be specfiied.
type WorkspaceSpec struct {
	// Specify workspace using its ID
	ID *string

	// Specify workspace using its name and organization
	Name             *string `schema:"workspace_name"`
	OrganizationName *string `schema:"organization_name"`
}

// LogFields provides fields for logging
func (spec WorkspaceSpec) LogFields() (fields []interface{}) {
	if spec.ID != nil {
		fields = append(fields, "id", *spec.ID)
	}
	if spec.Name != nil && spec.OrganizationName != nil {
		fields = append(fields, "name", *spec.Name, "organization", *spec.OrganizationName)
	}
	return fields
}

func (spec *WorkspaceSpec) String() string {
	switch {
	case spec.ID != nil:
		return *spec.ID
	case spec.Name != nil && spec.OrganizationName != nil:
		return *spec.OrganizationName + "/" + *spec.Name
	default:
		panic("invalid workspace spec")
	}
}

func (spec *WorkspaceSpec) Valid() error {
	if spec.ID != nil {
		if *spec.ID == "" {
			return fmt.Errorf("id is an empty string")
		}
		return nil
	}

	// No ID specified; both org and workspace name must be specified

	if spec.Name == nil {
		return fmt.Errorf("workspace name nor id specified")
	}

	if spec.OrganizationName == nil {
		return fmt.Errorf("must specify both organization and workspace")
	}

	if *spec.Name == "" {
		return fmt.Errorf("workspace name is an empty string")
	}

	if *spec.OrganizationName == "" {
		return fmt.Errorf("organization name is an empty string")
	}

	return nil
}
