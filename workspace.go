package ots

import (
	"errors"
	"time"

	tfe "github.com/leg100/go-tfe"
	"gorm.io/gorm"
)

const (
	DefaultAllowDestroyPlan    = true
	DefaultFileTriggersEnabled = true
	DefaultTerraformVersion    = "0.15.4"
)

var (
	ErrWorkspaceAlreadyLocked    = errors.New("workspace already locked")
	ErrWorkspaceAlreadyUnlocked  = errors.New("workspace already unlocked")
	ErrInvalidWorkspaceSpecifier = errors.New("invalid workspace specifier options")
)

// Workspace represents a Terraform Enterprise workspace.
type Workspace struct {
	ID string

	gorm.Model

	AllowDestroyPlan           bool
	AutoApply                  bool
	CanQueueDestroyPlan        bool
	Description                string
	Environment                string
	ExecutionMode              string
	FileTriggersEnabled        bool
	GlobalRemoteState          bool
	Locked                     bool
	MigrationEnvironment       string
	Name                       string
	Operations                 bool
	Permissions                *tfe.WorkspacePermissions
	QueueAllRuns               bool
	SpeculativeEnabled         bool
	SourceName                 string
	SourceURL                  string
	StructuredRunOutputEnabled bool
	TerraformVersion           string
	VCSRepo                    *tfe.VCSRepo
	WorkingDirectory           string
	ResourceCount              int
	ApplyDurationAverage       time.Duration
	PlanDurationAverage        time.Duration
	PolicyCheckFailures        int
	RunFailures                int
	RunsCount                  int

	TriggerPrefixes []string

	// Relations AgentPool  *tfe.AgentPool CurrentRun *Run

	// Workspace belongs to an organization
	Organization *Organization

	//SSHKey *tfe.SSHKey
}

// WorkspaceList represents a list of Workspaces.
type WorkspaceList struct {
	*tfe.Pagination
	Items []*Workspace
}

type WorkspaceService interface {
	Create(org string, opts *tfe.WorkspaceCreateOptions) (*Workspace, error)
	Get(name, org string) (*Workspace, error)
	GetByID(id string) (*Workspace, error)
	List(org string, opts tfe.WorkspaceListOptions) (*WorkspaceList, error)
	Update(name, org string, opts *tfe.WorkspaceUpdateOptions) (*Workspace, error)
	UpdateByID(id string, opts *tfe.WorkspaceUpdateOptions) (*Workspace, error)
	Lock(id string, opts tfe.WorkspaceLockOptions) (*Workspace, error)
	Unlock(id string) (*Workspace, error)
	Delete(name, org string) error
	DeleteByID(id string) error
}

type WorkspaceRepository interface {
	Create(ws *Workspace) (*Workspace, error)
	Get(spec WorkspaceSpecifier) (*Workspace, error)
	List(organizationID string, opts WorkspaceListOptions) (*WorkspaceList, error)
	Update(spec WorkspaceSpecifier, fn func(*Workspace) error) (*Workspace, error)
	Delete(spec WorkspaceSpecifier) error
}

// WorkspaceSpecifier is used for identifying an individual workspace. Either ID
// *or* both Name and OrganizationName must be specfiied.
type WorkspaceSpecifier struct {
	// Specify workspace using its ID
	ID *string

	// Specify workspace using its name and organization
	Name             *string
	OrganizationName *string
}

// WorkspaceListOptions are options for paginating and filtering the list of
// Workspaces to retrieve from the WorkspaceRepository ListWorkspaces endpoint
type WorkspaceListOptions struct {
	tfe.ListOptions

	// Optionally filter workspaces with name matching prefix
	Prefix *string
}

func NewWorkspace(opts *tfe.WorkspaceCreateOptions, org *Organization) *Workspace {
	ws := Workspace{
		ID:                  GenerateID("ws"),
		Name:                *opts.Name,
		AllowDestroyPlan:    DefaultAllowDestroyPlan,
		ExecutionMode:       "local", // Only local execution mode is supported
		FileTriggersEnabled: DefaultFileTriggersEnabled,
		GlobalRemoteState:   true, // Only global remote state is supported
		TerraformVersion:    DefaultTerraformVersion,
		Permissions: &tfe.WorkspacePermissions{
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
		SpeculativeEnabled: true,
		Operations:         true,
		Organization:       org,
	}

	if opts.AllowDestroyPlan != nil {
		ws.AllowDestroyPlan = *opts.AllowDestroyPlan
	}
	if opts.AutoApply != nil {
		ws.AutoApply = *opts.AutoApply
	}
	if opts.Description != nil {
		ws.Description = *opts.Description
	}
	if opts.FileTriggersEnabled != nil {
		ws.FileTriggersEnabled = *opts.FileTriggersEnabled
	}
	if opts.Operations != nil {
		ws.Operations = *opts.Operations
	}
	if opts.QueueAllRuns != nil {
		ws.QueueAllRuns = *opts.QueueAllRuns
	}
	if opts.SourceName != nil {
		ws.SourceName = *opts.SourceName
	}
	if opts.SourceURL != nil {
		ws.SourceURL = *opts.SourceURL
	}
	if opts.SpeculativeEnabled != nil {
		ws.SpeculativeEnabled = *opts.SpeculativeEnabled
	}
	if opts.StructuredRunOutputEnabled != nil {
		ws.StructuredRunOutputEnabled = *opts.StructuredRunOutputEnabled
	}
	if opts.TerraformVersion != nil {
		ws.TerraformVersion = *opts.TerraformVersion
	}
	if opts.TriggerPrefixes != nil {
		ws.TriggerPrefixes = opts.TriggerPrefixes
	}
	if opts.WorkingDirectory != nil {
		ws.WorkingDirectory = *opts.WorkingDirectory
	}

	return &ws
}

func UpdateWorkspace(ws *Workspace, opts *tfe.WorkspaceUpdateOptions) (*Workspace, error) {
	if opts.Name != nil {
		ws.Name = *opts.Name
	}

	if opts.AllowDestroyPlan != nil {
		ws.AllowDestroyPlan = *opts.AllowDestroyPlan
	}
	if opts.AutoApply != nil {
		ws.AutoApply = *opts.AutoApply
	}
	if opts.Description != nil {
		ws.Description = *opts.Description
	}
	if opts.FileTriggersEnabled != nil {
		ws.FileTriggersEnabled = *opts.FileTriggersEnabled
	}
	if opts.Operations != nil {
		ws.Operations = *opts.Operations
	}
	if opts.QueueAllRuns != nil {
		ws.QueueAllRuns = *opts.QueueAllRuns
	}
	if opts.SpeculativeEnabled != nil {
		ws.SpeculativeEnabled = *opts.SpeculativeEnabled
	}
	if opts.StructuredRunOutputEnabled != nil {
		ws.StructuredRunOutputEnabled = *opts.StructuredRunOutputEnabled
	}
	if opts.TerraformVersion != nil {
		ws.TerraformVersion = *opts.TerraformVersion
	}
	if opts.TriggerPrefixes != nil {
		ws.TriggerPrefixes = opts.TriggerPrefixes
	}
	if opts.WorkingDirectory != nil {
		ws.WorkingDirectory = *opts.WorkingDirectory
	}

	return ws, nil
}

func (ws *Workspace) Actions() *tfe.WorkspaceActions {
	return &tfe.WorkspaceActions{
		IsDestroyable: false,
	}
}

func (ws *Workspace) ToggleLock(lock bool) error {
	if lock && ws.Locked {
		return ErrWorkspaceAlreadyLocked
	}
	if !lock && !ws.Locked {
		return ErrWorkspaceAlreadyUnlocked
	}

	ws.Locked = lock

	return nil
}
