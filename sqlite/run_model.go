package sqlite

import (
	"database/sql"
	"strings"
	"time"

	"github.com/leg100/go-tfe"
	"github.com/leg100/otf"
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
	StatusTimestamps       *RunStatusTimestamps `gorm:"embedded;embeddedPrefix:timestamp_"`

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

type RunStatusTimestamps struct {
	AppliedAt            sql.NullTime
	ApplyQueuedAt        sql.NullTime
	ApplyingAt           sql.NullTime
	CanceledAt           sql.NullTime
	ConfirmedAt          sql.NullTime
	CostEstimatedAt      sql.NullTime
	CostEstimatingAt     sql.NullTime
	DiscardedAt          sql.NullTime
	ErroredAt            sql.NullTime
	ForceCanceledAt      sql.NullTime
	PlanQueueableAt      sql.NullTime
	PlanQueuedAt         sql.NullTime
	PlannedAndFinishedAt sql.NullTime
	PlannedAt            sql.NullTime
	PlanningAt           sql.NullTime
	PolicyCheckedAt      sql.NullTime
	PolicySoftFailedAt   sql.NullTime
}

// RunList is a list of run models
type RunList []Run

// Update updates the model with the supplied fn. The fn operates on the domain
// obj, so Update handles converting to and from a domain obj.
func (model *Run) Update(fn func(*otf.Run) error) error {
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

func (model *Run) ToDomain() *otf.Run {
	domain := otf.Run{
		ID:                     model.ExternalID,
		ForceCancelAvailableAt: model.ForceCancelAvailableAt,
		Message:                model.Message,
		Permissions:            model.Permissions,
		Refresh:                model.Refresh,
		RefreshOnly:            model.RefreshOnly,
		Status:                 model.Status,
		StatusTimestamps:       &tfe.RunStatusTimestamps{},
	}

	if model.StatusTimestamps.AppliedAt.Valid {
		domain.StatusTimestamps.AppliedAt = &model.StatusTimestamps.AppliedAt.Time
	}

	if model.StatusTimestamps.ApplyQueuedAt.Valid {
		domain.StatusTimestamps.ApplyQueuedAt = &model.StatusTimestamps.ApplyQueuedAt.Time
	}

	if model.StatusTimestamps.ApplyingAt.Valid {
		domain.StatusTimestamps.ApplyingAt = &model.StatusTimestamps.ApplyingAt.Time
	}

	if model.StatusTimestamps.CanceledAt.Valid {
		domain.StatusTimestamps.CanceledAt = &model.StatusTimestamps.CanceledAt.Time
	}

	if model.StatusTimestamps.ConfirmedAt.Valid {
		domain.StatusTimestamps.ConfirmedAt = &model.StatusTimestamps.ConfirmedAt.Time
	}

	if model.StatusTimestamps.DiscardedAt.Valid {
		domain.StatusTimestamps.DiscardedAt = &model.StatusTimestamps.DiscardedAt.Time
	}

	if model.StatusTimestamps.ErroredAt.Valid {
		domain.StatusTimestamps.ErroredAt = &model.StatusTimestamps.ErroredAt.Time
	}

	if model.StatusTimestamps.ForceCanceledAt.Valid {
		domain.StatusTimestamps.ForceCanceledAt = &model.StatusTimestamps.ForceCanceledAt.Time
	}

	if model.StatusTimestamps.PlanQueueableAt.Valid {
		domain.StatusTimestamps.PlanQueueableAt = &model.StatusTimestamps.PlanQueueableAt.Time
	}

	if model.StatusTimestamps.PlanQueuedAt.Valid {
		domain.StatusTimestamps.PlanQueuedAt = &model.StatusTimestamps.PlanQueuedAt.Time
	}

	if model.StatusTimestamps.PlannedAndFinishedAt.Valid {
		domain.StatusTimestamps.PlannedAndFinishedAt = &model.StatusTimestamps.PlannedAndFinishedAt.Time
	}

	if model.StatusTimestamps.PlannedAt.Valid {
		domain.StatusTimestamps.PlannedAt = &model.StatusTimestamps.PlannedAt.Time
	}

	if model.StatusTimestamps.PlanningAt.Valid {
		domain.StatusTimestamps.PlanningAt = &model.StatusTimestamps.PlanningAt.Time
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
func NewFromDomain(domain *otf.Run) *Run {
	model := &Run{
		Apply: &Apply{
			StatusTimestamps: &ApplyStatusTimestamps{},
		},
		ConfigurationVersion: &ConfigurationVersion{},
		Plan: &Plan{
			StatusTimestamps: &PlanStatusTimestamps{},
		},
		Workspace:        &Workspace{},
		StatusTimestamps: &RunStatusTimestamps{},
	}
	model.FromDomain(domain)

	return model
}

// FromDomain updates run model fields with a run domain object's fields
func (model *Run) FromDomain(domain *otf.Run) {
	model.ExternalID = domain.ID
	model.ForceCancelAvailableAt = domain.ForceCancelAvailableAt
	model.Status = domain.Status
	model.Message = domain.Message
	model.Permissions = domain.Permissions
	model.Refresh = domain.Refresh
	model.RefreshOnly = domain.RefreshOnly
	model.ReplaceAddrs = strings.Join(domain.ReplaceAddrs, ",")
	model.TargetAddrs = strings.Join(domain.TargetAddrs, ",")

	if domain.StatusTimestamps.AppliedAt != nil {
		model.StatusTimestamps.AppliedAt.Time = *domain.StatusTimestamps.AppliedAt
		model.StatusTimestamps.AppliedAt.Valid = true
	}

	if domain.StatusTimestamps.ApplyQueuedAt != nil {
		model.StatusTimestamps.ApplyQueuedAt.Time = *domain.StatusTimestamps.ApplyQueuedAt
		model.StatusTimestamps.ApplyQueuedAt.Valid = true
	}

	if domain.StatusTimestamps.ApplyingAt != nil {
		model.StatusTimestamps.ApplyingAt.Time = *domain.StatusTimestamps.ApplyingAt
		model.StatusTimestamps.ApplyingAt.Valid = true
	}

	if domain.StatusTimestamps.CanceledAt != nil {
		model.StatusTimestamps.CanceledAt.Time = *domain.StatusTimestamps.CanceledAt
		model.StatusTimestamps.CanceledAt.Valid = true
	}

	if domain.StatusTimestamps.ConfirmedAt != nil {
		model.StatusTimestamps.ConfirmedAt.Time = *domain.StatusTimestamps.ConfirmedAt
		model.StatusTimestamps.ConfirmedAt.Valid = true
	}

	if domain.StatusTimestamps.DiscardedAt != nil {
		model.StatusTimestamps.DiscardedAt.Time = *domain.StatusTimestamps.DiscardedAt
		model.StatusTimestamps.DiscardedAt.Valid = true
	}

	if domain.StatusTimestamps.ErroredAt != nil {
		model.StatusTimestamps.ErroredAt.Time = *domain.StatusTimestamps.ErroredAt
		model.StatusTimestamps.ErroredAt.Valid = true
	}

	if domain.StatusTimestamps.ForceCanceledAt != nil {
		model.StatusTimestamps.ForceCanceledAt.Time = *domain.StatusTimestamps.ForceCanceledAt
		model.StatusTimestamps.ForceCanceledAt.Valid = true
	}

	if domain.StatusTimestamps.PlanQueueableAt != nil {
		model.StatusTimestamps.PlanQueueableAt.Time = *domain.StatusTimestamps.PlanQueueableAt
		model.StatusTimestamps.PlanQueueableAt.Valid = true
	}

	if domain.StatusTimestamps.PlanQueuedAt != nil {
		model.StatusTimestamps.PlanQueuedAt.Time = *domain.StatusTimestamps.PlanQueuedAt
		model.StatusTimestamps.PlanQueuedAt.Valid = true
	}

	if domain.StatusTimestamps.PlannedAt != nil {
		model.StatusTimestamps.PlannedAt.Time = *domain.StatusTimestamps.PlannedAt
		model.StatusTimestamps.PlannedAt.Valid = true
	}

	if domain.StatusTimestamps.PlanningAt != nil {
		model.StatusTimestamps.PlanningAt.Time = *domain.StatusTimestamps.PlanningAt
		model.StatusTimestamps.PlanningAt.Valid = true
	}

	if domain.StatusTimestamps.PlannedAndFinishedAt != nil {
		model.StatusTimestamps.PlannedAndFinishedAt.Time = *domain.StatusTimestamps.PlannedAndFinishedAt
		model.StatusTimestamps.PlannedAndFinishedAt.Valid = true
	}

	model.Apply.FromDomain(domain.Apply)

	model.Plan.FromDomain(domain.Plan)

	model.Workspace.FromDomain(domain.Workspace)
	model.WorkspaceID = domain.Workspace.Model.ID

	model.ConfigurationVersion.FromDomain(domain.ConfigurationVersion)
	model.ConfigurationVersionID = domain.ConfigurationVersion.Model.ID
}

func (l RunList) ToDomain() (dl []*otf.Run) {
	for _, i := range l {
		dl = append(dl, i.ToDomain())
	}
	return
}
