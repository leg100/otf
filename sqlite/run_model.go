package sqlite

import (
	"strings"
	"time"

	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"gorm.io/gorm"
)

// Run models a row in a runs table.
type Run struct {
	gorm.Model

	ExternalID string `gorm:"uniqueIndex"`

	ForceCancelAvailableAt time.Time
	IsDestroy              bool
	Message                string
	Permissions            *tfe.RunPermissions `gorm:"embedded;embeddedPrefix:permission_"`
	PositionInQueue        int
	Refresh                bool
	RefreshOnly            bool
	Status                 tfe.RunStatus
	StatusTimestamps       *tfe.RunStatusTimestamps `gorm:"embedded;embeddedPrefix:timestamp_"`

	// Comma separated list of replace addresses
	ReplaceAddrs string
	// Comma separated list of target addresses
	TargetAddrs string

	// Run has one plan
	Plan *Plan

	// Run has one apply
	Apply *Apply

	// Run belongs to a workspace
	WorkspaceID uint
	Workspace   *Workspace

	// Run belongs to a configuration version
	ConfigurationVersionID uint
	ConfigurationVersion   *ConfigurationVersion
}

// RunList is a list of run models
type RunList []Run

// Update updates the model with the supplied fn. The fn operates on the domain
// obj, so Update handles converting to and from a domain obj.
func (r *Run) Update(fn func(*ots.Run) error) error {
	// model -> domain
	domain := r.ToDomain()

	// invoke user fn
	if err := fn(domain); err != nil {
		return err
	}

	// domain -> model
	r.FromDomain(domain)

	return nil
}

func (model *Run) ToDomain() *ots.Run {
	domain := ots.Run{
		ID:                     model.ExternalID,
		ForceCancelAvailableAt: model.ForceCancelAvailableAt,
		Message:                model.Message,
		Permissions:            model.Permissions,
		Refresh:                model.Refresh,
		RefreshOnly:            model.RefreshOnly,
		Status:                 model.Status,
		StatusTimestamps:       model.StatusTimestamps,
	}

	if model.Apply != nil {
		domain.Apply = model.Apply.ToDomain()
	}

	if model.ConfigurationVersion != nil {
		domain.ConfigurationVersion = model.ConfigurationVersion.ToDomain()
	}

	if model.Plan != nil {
		domain.Plan = model.Plan.ToDomain()
	}

	if model.Workspace != nil {
		domain.Workspace = model.Workspace.ToDomain()
	}

	if model.ReplaceAddrs != "" {
		domain.ReplaceAddrs = strings.Split(model.ReplaceAddrs, ",")
	}

	if model.TargetAddrs != "" {
		domain.TargetAddrs = strings.Split(model.TargetAddrs, ",")
	}

	return &domain
}

// NewFromDomain constructs a model obj from a domain obj
func NewFromDomain(domain *ots.Run) *Run {
	model := &Run{
		Apply:                &Apply{},
		ConfigurationVersion: &ConfigurationVersion{},
		Plan:                 &Plan{},
		Workspace:            &Workspace{},
	}
	model.FromDomain(domain)

	return model
}

// FromDomain updates run model fields with a run domain object's fields
func (model *Run) FromDomain(domain *ots.Run) {
	model.ExternalID = domain.ID
	model.ForceCancelAvailableAt = domain.ForceCancelAvailableAt
	model.Status = domain.Status
	model.Message = domain.Message
	model.Permissions = domain.Permissions
	model.Refresh = domain.Refresh
	model.RefreshOnly = domain.RefreshOnly
	model.ReplaceAddrs = strings.Join(domain.ReplaceAddrs, ",")
	model.TargetAddrs = strings.Join(domain.TargetAddrs, ",")
	model.StatusTimestamps = domain.StatusTimestamps

	model.Apply.FromDomain(domain.Apply)

	model.Plan.FromDomain(domain.Plan)

	model.Workspace.FromDomain(domain.Workspace)
	model.WorkspaceID = domain.Workspace.Model.ID

	model.ConfigurationVersion.FromDomain(domain.ConfigurationVersion)
	model.ConfigurationVersionID = domain.ConfigurationVersion.Model.ID
}

func (l RunList) ToDomain() (dl []*ots.Run) {
	for _, i := range l {
		dl = append(dl, i.ToDomain())
	}
	return
}
