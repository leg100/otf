package otf

import (
	"context"
	"errors"
	"fmt"
)

const (
	DefaultAllowDestroyPlan                  = true
	DefaultFileTriggersEnabled               = true
	DefaultTerraformVersion                  = "1.0.10"
	DefaultExecutionMode                     = "remote"
	RemoteExecutionMode        ExecutionMode = "remote"
	LocalExecutionMode         ExecutionMode = "local"
)

var (
	ErrWorkspaceAlreadyLocked   = errors.New("workspace already locked")
	ErrWorkspaceAlreadyUnlocked = errors.New("workspace already unlocked")
	ErrInvalidWorkspaceSpec     = errors.New("invalid workspace spec options")
)

type ExecutionMode string

// Workspace represents a Terraform Enterprise workspace.
type Workspace struct {
	id string `json:"workspace_id" jsonapi:"primary,workspaces" schema:"workspace_id"`

	// Timestamps records timestamps of lifecycle transitions
	Timestamps

	allowDestroyPlan           bool
	autoApply                  bool
	canQueueDestroyPlan        bool
	description                string
	environment                string
	executionMode              string
	fileTriggersEnabled        bool
	globalRemoteState          bool
	locked                     bool
	migrationEnvironment       string
	name                       string `schema:"workspace_name"`
	queueAllRuns               bool
	speculativeEnabled         bool
	structuredRunOutputEnabled bool
	sourceName                 string
	sourceURL                  string
	terraformVersion           string
	triggerPrefixes            []string
	workingDirectory           string

	// Workspace belongs to an organization
	Organization *Organization `json:"organization"`
}

// WorkspaceCreateOptions represents the options for creating a new workspace.
type WorkspaceCreateOptions struct {
	AllowDestroyPlan           *bool
	AutoApply                  *bool
	Description                *string
	ExecutionMode              *string
	FileTriggersEnabled        *bool
	GlobalRemoteState          *bool
	MigrationEnvironment       *string
	Name                       string
	OrganizationName           string
	QueueAllRuns               *bool
	SpeculativeEnabled         *bool
	SourceName                 *string
	SourceURL                  *string
	StructuredRunOutputEnabled *bool
	TerraformVersion           *string
	TriggerPrefixes            []string
	WorkingDirectory           *string
}

// WorkspaceUpdateOptions represents the options for updating a workspace.
type WorkspaceUpdateOptions struct {
	AllowDestroyPlan           *bool
	AutoApply                  *bool
	Name                       *string
	Description                *string
	ExecutionMode              *string
	FileTriggersEnabled        *bool
	GlobalRemoteState          *bool
	Operations                 *bool
	QueueAllRuns               *bool
	SpeculativeEnabled         *bool
	StructuredRunOutputEnabled *bool
	TerraformVersion           *string
	TriggerPrefixes            []string
	WorkingDirectory           *string
}

func (o WorkspaceUpdateOptions) Valid() error {
	if o.Name != nil && !ValidStringID(o.Name) {
		return ErrInvalidName
	}
	if o.TerraformVersion != nil && !validSemanticVersion(*o.TerraformVersion) {
		return ErrInvalidTerraformVersion
	}
	return nil
}

// WorkspaceLockOptions represents the options for locking a workspace.
type WorkspaceLockOptions struct {
	// Specifies the reason for locking the workspace.
	Reason *string `jsonapi:"attr,reason,omitempty"`
}

// WorkspaceList represents a list of Workspaces.
type WorkspaceList struct {
	*Pagination
	Items []*Workspace
}

type WorkspaceService interface {
	Create(ctx context.Context, opts WorkspaceCreateOptions) (*Workspace, error)
	Get(ctx context.Context, spec WorkspaceSpec) (*Workspace, error)
	List(ctx context.Context, opts WorkspaceListOptions) (*WorkspaceList, error)
	Update(ctx context.Context, spec WorkspaceSpec, opts WorkspaceUpdateOptions) (*Workspace, error)
	Lock(ctx context.Context, spec WorkspaceSpec, opts WorkspaceLockOptions) (*Workspace, error)
	Unlock(ctx context.Context, spec WorkspaceSpec) (*Workspace, error)
	Delete(ctx context.Context, spec WorkspaceSpec) error
}

type WorkspaceStore interface {
	Create(ws *Workspace) (*Workspace, error)
	Get(spec WorkspaceSpec) (*Workspace, error)
	List(opts WorkspaceListOptions) (*WorkspaceList, error)
	Update(spec WorkspaceSpec, ws func(ws *Workspace) (bool, error)) (*Workspace, error)
	Delete(spec WorkspaceSpec) error
}

// WorkspaceSpec is used for identifying an individual workspace. Either ID *or*
// both Name and OrganizationName must be specfiied.
type WorkspaceSpec struct {
	// Specify workspace using its ID
	ID *string

	// Specify workspace using its name and organization
	Name             *string `schema:"workspace_name"`
	OrganizationName *string `schema:"organization_name"`

	// A list of relations to include. See available resources
	// https://www.terraform.io/docs/cloud/api/workspaces.html#available-related-resources
	Include *string `schema:"include"`
}

func (spec WorkspaceSpec) LogInfo() (keysAndValues []interface{}) {
	if spec.ID != nil {
		keysAndValues = append(keysAndValues, "id", *spec.ID)
	}
	if spec.Name != nil && spec.OrganizationName != nil {
		keysAndValues = append(keysAndValues, "name", *spec.Name, "organization", *spec.OrganizationName)
	}
	return keysAndValues
}

// WorkspaceListOptions are options for paginating and filtering a list of
// Workspaces
type WorkspaceListOptions struct {
	// Pagination
	ListOptions

	// Filter workspaces with name matching prefix.
	Prefix string `schema:"search[name],omitempty"`

	// OrganizationName filters workspaces by organization name. Required.
	OrganizationName string `schema:"organization_name,omitempty"`

	// A list of relations to include. See available resources https://www.terraform.io/docs/cloud/api/workspaces.html#available-related-resources
	Include *string `schema:"include"`
}

func (ws *Workspace) ID() string                       { return ws.id }
func (ws *Workspace) String() string                   { return ws.id }
func (ws *Workspace) Name() string                     { return ws.name }
func (ws *Workspace) AllowDestroyPlan() bool           { return ws.allowDestroyPlan }
func (ws *Workspace) CanQueueDestroyPlan() bool        { return ws.canQueueDestroyPlan }
func (ws *Workspace) Environment() string              { return ws.environment }
func (ws *Workspace) Description() string              { return ws.description }
func (ws *Workspace) ExecutionMode() string            { return ws.executionMode }
func (ws *Workspace) FileTriggersEnabled() bool        { return ws.fileTriggersEnabled }
func (ws *Workspace) GlobalRemoteState() bool          { return ws.globalRemoteState }
func (ws *Workspace) Locked() bool                     { return ws.locked }
func (ws *Workspace) MigrationEnvironment() string     { return ws.migrationEnvironment }
func (ws *Workspace) SourceName() string               { return ws.sourceName }
func (ws *Workspace) SourceURL() string                { return ws.sourceURL }
func (ws *Workspace) SpeculativeEnabled() bool         { return ws.speculativeEnabled }
func (ws *Workspace) StructuredRunOutputEnabled() bool { return ws.structuredRunOutputEnabled }
func (ws *Workspace) TerraformVersion() string         { return ws.terraformVersion }
func (ws *Workspace) TriggerPrefixes() []string        { return ws.triggerPrefixes }
func (ws *Workspace) QueueAllRuns() bool               { return ws.queueAllRuns }
func (ws *Workspace) AutoApply() bool                  { return ws.autoApply }
func (ws *Workspace) WorkingDirectory() string         { return ws.workingDirectory }
func (ws *Workspace) OrganizationID() string           { return ws.Organization.ID() }

func (o WorkspaceCreateOptions) Valid() error {
	if !ValidStringID(&o.Name) {
		return ErrInvalidName
	}
	if o.TerraformVersion != nil && !validSemanticVersion(*o.TerraformVersion) {
		return ErrInvalidTerraformVersion
	}

	return nil
}

// ToggleLock toggles the workspace lock.
func (ws *Workspace) ToggleLock(lock bool) error {
	if lock && ws.locked {
		return ErrWorkspaceAlreadyLocked
	}
	if !lock && !ws.locked {
		return ErrWorkspaceAlreadyUnlocked
	}

	ws.locked = lock

	return nil
}

func (ws *Workspace) UpdateWithOptions(ctx context.Context, opts WorkspaceUpdateOptions) (updated bool, err error) {
	if opts.Name != nil {
		ws.name = *opts.Name
		updated = true
	}
	if opts.AllowDestroyPlan != nil {
		ws.allowDestroyPlan = *opts.AllowDestroyPlan
		updated = true
	}
	if opts.AutoApply != nil {
		ws.autoApply = *opts.AutoApply
		updated = true
	}
	if opts.Description != nil {
		ws.description = *opts.Description
		updated = true
	}
	if opts.ExecutionMode != nil {
		ws.executionMode = *opts.ExecutionMode
		updated = true
	}
	if opts.FileTriggersEnabled != nil {
		ws.fileTriggersEnabled = *opts.FileTriggersEnabled
		updated = true
	}
	if opts.Operations != nil {
		if *opts.Operations {
			ws.executionMode = "remote"
		} else {
			ws.executionMode = "local"
		}
		updated = true
	}
	if opts.QueueAllRuns != nil {
		ws.queueAllRuns = *opts.QueueAllRuns
		updated = true
	}
	if opts.SpeculativeEnabled != nil {
		ws.speculativeEnabled = *opts.SpeculativeEnabled
		updated = true
	}
	if opts.StructuredRunOutputEnabled != nil {
		ws.structuredRunOutputEnabled = *opts.StructuredRunOutputEnabled
		updated = true
	}
	if opts.TerraformVersion != nil {
		ws.terraformVersion = *opts.TerraformVersion
		updated = true
	}
	if opts.TriggerPrefixes != nil {
		ws.triggerPrefixes = opts.TriggerPrefixes
		updated = true
	}
	if opts.WorkingDirectory != nil {
		ws.workingDirectory = *opts.WorkingDirectory
		updated = true
	}

	return updated, nil
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
