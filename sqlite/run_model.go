package sqlite

import (
	"database/sql"
	"strings"

	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"gorm.io/gorm"
)

// Run models a row in a runs table.
type Run struct {
	gorm.Model

	ExternalID string `gorm:"uniqueIndex"`

	ForceCancelAvailableAt sql.NullTime
	IsDestroy              bool
	Message                string
	Permissions            *tfe.RunPermissions `gorm:"embedded;embeddedPrefix:permission_"`
	PositionInQueue        int
	Refresh                bool
	RefreshOnly            bool
	Status                 tfe.RunStatus
	StatusTimestamps       RunStatusTimestamps `gorm:"embedded;embeddedPrefix:timestamp_"`

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

// RunStatusTimestamps holds the timestamps for individual run statuses.
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

func (r *Run) ToDomain() *ots.Run {
	domain := ots.Run{
		ID:      r.ExternalID,
		Refresh: r.Refresh,
	}

	if r.Plan != nil {
		domain.Plan = r.Plan.ToDomain()
	}

	if r.ReplaceAddrs != "" {
		domain.ReplaceAddrs = strings.Split(r.ReplaceAddrs, ",")
	}

	if r.TargetAddrs != "" {
		domain.TargetAddrs = strings.Split(r.TargetAddrs, ",")
	}

	return &domain
}

// FromDomain updates run model fields with a run domain object's fields
func (model *Run) FromDomain(domain *ots.Run) {
	model.ExternalID = domain.ID
	model.Status = domain.Status
	model.Message = domain.Message
	model.Refresh = domain.Refresh
	model.RefreshOnly = domain.RefreshOnly
	model.ReplaceAddrs = strings.Join(domain.ReplaceAddrs, ",")
	model.TargetAddrs = strings.Join(domain.TargetAddrs, ",")

	if domain.Plan != nil {
		model.Plan = &Plan{}
		model.Plan.FromDomain(domain.Plan)
	}

	model.Workspace = &Workspace{}
	model.Workspace.FromDomain(domain.Workspace)
	model.WorkspaceID = domain.Workspace.Model.ID

	model.ConfigurationVersion = &ConfigurationVersion{}
	model.ConfigurationVersion.FromDomain(domain.ConfigurationVersion)
	model.ConfigurationVersionID = domain.ConfigurationVersion.Model.ID
}

func (l RunList) ToDomain() (dl []*ots.Run) {
	for _, i := range l {
		dl = append(dl, i.ToDomain())
	}
	return
}
