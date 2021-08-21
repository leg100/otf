package sqlite

import (
	"strings"
	"time"

	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"gorm.io/gorm"
)

// Workspace models a row in a runs table.
type Workspace struct {
	gorm.Model

	ExternalID string `gorm:"uniqueIndex"`

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
	Permissions                *tfe.WorkspacePermissions `gorm:"embedded;embeddedPrefix:permission_"`
	QueueAllRuns               bool
	SpeculativeEnabled         bool
	SourceName                 string
	SourceURL                  string
	StructuredRunOutputEnabled bool
	TerraformVersion           string
	TriggerPrefixes            string
	VCSRepo                    *tfe.VCSRepo `gorm:"-"`
	WorkingDirectory           string
	ResourceCount              int
	ApplyDurationAverage       time.Duration
	PlanDurationAverage        time.Duration
	PolicyCheckFailures        int
	RunFailures                int
	RunsCount                  int

	OrganizationID uint
	Organization   *Organization
}

// WorkspaceList is a list of run models
type WorkspaceList []Workspace

// Update updates the model with the supplied fn. The fn operates on the domain
// obj, so Update handles converting to and from a domain obj.
func (model *Workspace) Update(fn func(*ots.Workspace) error) error {
	// model -> domain
	domain := model.ToDomain()

	// invoke user fn
	if err := fn(domain); err != nil {
		return err
	}

	// domain -> model
	model.FromDomain(domain)

	return nil
}

func (model *Workspace) ToDomain() *ots.Workspace {
	domain := ots.Workspace{
		ID:                         model.ExternalID,
		Model:                      model.Model,
		AllowDestroyPlan:           model.AllowDestroyPlan,
		AutoApply:                  model.AutoApply,
		CanQueueDestroyPlan:        model.CanQueueDestroyPlan,
		Description:                model.Description,
		Environment:                model.Environment,
		ExecutionMode:              model.ExecutionMode,
		FileTriggersEnabled:        model.FileTriggersEnabled,
		GlobalRemoteState:          model.GlobalRemoteState,
		Locked:                     model.Locked,
		MigrationEnvironment:       model.MigrationEnvironment,
		Name:                       model.Name,
		Permissions:                model.Permissions,
		QueueAllRuns:               model.QueueAllRuns,
		SpeculativeEnabled:         model.SpeculativeEnabled,
		SourceName:                 model.SourceName,
		SourceURL:                  model.SourceURL,
		StructuredRunOutputEnabled: model.StructuredRunOutputEnabled,
		TerraformVersion:           model.TerraformVersion,
		VCSRepo:                    model.VCSRepo,
		WorkingDirectory:           model.WorkingDirectory,
		ResourceCount:              model.ResourceCount,
		ApplyDurationAverage:       model.ApplyDurationAverage,
		PlanDurationAverage:        model.PlanDurationAverage,
		PolicyCheckFailures:        model.PolicyCheckFailures,
		RunFailures:                model.RunFailures,
		RunsCount:                  model.RunsCount,
	}

	// A model doesn't necessarily include its relations
	if model.Organization != nil {
		domain.Organization = model.Organization.ToDomain()
	}

	if model.TriggerPrefixes != "" {
		domain.TriggerPrefixes = strings.Split(model.TriggerPrefixes, ",")
	}

	return &domain
}

// FromDomain updates workspace model fields with a workspace domain object's
// fields
func (model *Workspace) FromDomain(domain *ots.Workspace) {
	model.ExternalID = domain.ID
	model.Model = domain.Model
	model.AllowDestroyPlan = domain.AllowDestroyPlan
	model.AutoApply = domain.AutoApply
	model.CanQueueDestroyPlan = domain.CanQueueDestroyPlan
	model.Description = domain.Description
	model.Environment = domain.Environment
	model.ExecutionMode = domain.ExecutionMode
	model.FileTriggersEnabled = domain.FileTriggersEnabled
	model.GlobalRemoteState = domain.GlobalRemoteState
	model.Locked = domain.Locked
	model.MigrationEnvironment = domain.MigrationEnvironment
	model.Name = domain.Name
	model.Permissions = domain.Permissions
	model.QueueAllRuns = domain.QueueAllRuns
	model.SpeculativeEnabled = domain.SpeculativeEnabled
	model.SourceName = domain.SourceName
	model.SourceURL = domain.SourceURL
	model.StructuredRunOutputEnabled = domain.StructuredRunOutputEnabled
	model.TerraformVersion = domain.TerraformVersion
	model.TriggerPrefixes = strings.Join(domain.TriggerPrefixes, ",")
	model.VCSRepo = domain.VCSRepo
	model.WorkingDirectory = domain.WorkingDirectory
	model.ResourceCount = domain.ResourceCount
	model.ApplyDurationAverage = domain.ApplyDurationAverage
	model.PlanDurationAverage = domain.PlanDurationAverage
	model.PolicyCheckFailures = domain.PolicyCheckFailures
	model.RunFailures = domain.RunFailures
	model.RunsCount = domain.RunsCount

	// A ws domain obj doesn't necessarily include an org obj
	if domain.Organization != nil {
		model.Organization = &Organization{}
		model.Organization.FromDomain(domain.Organization)
		model.OrganizationID = domain.Organization.Model.ID
	}
}

func (l WorkspaceList) ToDomain() (dl []*ots.Workspace) {
	for _, i := range l {
		dl = append(dl, i.ToDomain())
	}
	return
}
