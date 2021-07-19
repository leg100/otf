package ots

import (
	"errors"
	"fmt"
	"strings"
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
	ExternalID string `gorm:"uniqueIndex"`
	InternalID uint   `gorm:"primaryKey;column:id"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

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
	Permissions                *tfe.WorkspacePermissions `gorm:"embedded;embeddedPrefix:permission_"`
	QueueAllRuns               bool
	SpeculativeEnabled         bool
	SourceName                 string
	SourceURL                  string
	StructuredRunOutputEnabled bool
	TerraformVersion           string
	VCSRepo                    *tfe.VCSRepo `gorm:"-"`
	WorkingDirectory           string
	ResourceCount              int
	ApplyDurationAverage       time.Duration
	PlanDurationAverage        time.Duration
	PolicyCheckFailures        int
	RunFailures                int
	RunsCount                  int

	TriggerPrefixes         []string `gorm:"-"`
	InternalTriggerPrefixes string

	// Relations AgentPool  *tfe.AgentPool CurrentRun *Run

	// Workspace belongs to an organization
	OrganizationID uint
	Organization   *Organization

	//SSHKey *tfe.SSHKey
}

func (ws *Workspace) Unwrap(tx *gorm.DB) (err error) {
	ws.TriggerPrefixes = strings.Split(ws.InternalTriggerPrefixes, ",")
	return
}

func (ws *Workspace) Wrap(tx *gorm.DB) (err error) {
	ws.InternalTriggerPrefixes = strings.Join(ws.TriggerPrefixes, ",")
	return
}

func (ws *Workspace) AfterFind(tx *gorm.DB) (err error) { ws.Unwrap(tx); return }

func (ws *Workspace) BeforeSave(tx *gorm.DB) (err error) { ws.Wrap(tx); return }
func (ws *Workspace) AfterSave(tx *gorm.DB) (err error)  { ws.Unwrap(tx); return }

func (ws *Workspace) DTO() interface{} {
	dto := &tfe.Workspace{
		ID:                         ws.ExternalID,
		Actions:                    ws.Actions(),
		AllowDestroyPlan:           ws.AllowDestroyPlan,
		AutoApply:                  ws.AutoApply,
		CanQueueDestroyPlan:        ws.CanQueueDestroyPlan,
		CreatedAt:                  ws.CreatedAt,
		Description:                ws.Description,
		Environment:                ws.Environment,
		ExecutionMode:              ws.ExecutionMode,
		FileTriggersEnabled:        ws.FileTriggersEnabled,
		GlobalRemoteState:          ws.GlobalRemoteState,
		Locked:                     ws.Locked,
		MigrationEnvironment:       ws.MigrationEnvironment,
		Name:                       ws.Name,
		Operations:                 ws.Operations,
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
		UpdatedAt:                  ws.UpdatedAt,
		ResourceCount:              ws.ResourceCount,
		ApplyDurationAverage:       ws.ApplyDurationAverage,
		PlanDurationAverage:        ws.PlanDurationAverage,
		PolicyCheckFailures:        ws.PolicyCheckFailures,
		RunFailures:                ws.RunFailures,
		RunsCount:                  ws.RunsCount,
	}

	if ws.Organization != nil {
		dto.Organization = ws.Organization.DTO().(*tfe.Organization)
	}

	return dto
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
		ExternalID:          NewWorkspaceID(),
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
		OrganizationID:     org.InternalID,
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

func NewWorkspaceID() string {
	return fmt.Sprintf("ws-%s", GenerateRandomString(16))
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

func (wsl *WorkspaceList) DTO() interface{} {
	l := &tfe.WorkspaceList{
		Pagination: wsl.Pagination,
	}
	for _, item := range wsl.Items {
		l.Items = append(l.Items, item.DTO().(*tfe.Workspace))
	}

	return l
}
