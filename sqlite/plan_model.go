package sqlite

import (
	"database/sql"

	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"gorm.io/gorm"
)

// Plan models a row in a runs table.
type Plan struct {
	gorm.Model

	ExternalID string `gorm:"uniqueIndex"`

	ResourceAdditions    int
	ResourceChanges      int
	ResourceDestructions int
	Status               tfe.PlanStatus
	StatusTimestamps     *PlanStatusTimestamps `gorm:"embedded;embeddedPrefix:timestamp_"`

	Logs []byte

	// The blob ID of the execution plan file
	PlanFileBlobID string

	// The blob ID of the execution plan file in json format
	PlanJSONBlobID string

	// Plan belongs to a run
	RunID uint
}

// PlanStatusTimestamps holds the timestamps for individual plan statuses.
type PlanStatusTimestamps struct {
	CanceledAt      sql.NullTime
	ErroredAt       sql.NullTime
	FinishedAt      sql.NullTime
	ForceCanceledAt sql.NullTime
	QueuedAt        sql.NullTime
	StartedAt       sql.NullTime
}

// PlanList is a list of run models
type PlanList []Plan

// Update updates the model with the supplied fn. The fn operates on the domain
// obj, so Update handles converting to and from a domain obj.
func (model *Plan) Update(fn func(*ots.Plan) error) error {
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

func (model *Plan) ToDomain() *ots.Plan {
	domain := ots.Plan{
		ID:                   model.ExternalID,
		Model:                model.Model,
		ResourceAdditions:    model.ResourceAdditions,
		ResourceChanges:      model.ResourceChanges,
		ResourceDestructions: model.ResourceDestructions,
		Status:               model.Status,
		StatusTimestamps:     &tfe.PlanStatusTimestamps{},
		Logs:                 model.Logs,
		PlanFileBlobID:       model.PlanFileBlobID,
		PlanJSONBlobID:       model.PlanJSONBlobID,
	}

	if model.StatusTimestamps.CanceledAt.Valid {
		domain.StatusTimestamps.CanceledAt = &model.StatusTimestamps.CanceledAt.Time
	}

	if model.StatusTimestamps.ErroredAt.Valid {
		domain.StatusTimestamps.ErroredAt = &model.StatusTimestamps.ErroredAt.Time
	}

	if model.StatusTimestamps.FinishedAt.Valid {
		domain.StatusTimestamps.FinishedAt = &model.StatusTimestamps.FinishedAt.Time
	}

	if model.StatusTimestamps.ForceCanceledAt.Valid {
		domain.StatusTimestamps.ForceCanceledAt = &model.StatusTimestamps.ForceCanceledAt.Time
	}

	if model.StatusTimestamps.QueuedAt.Valid {
		domain.StatusTimestamps.QueuedAt = &model.StatusTimestamps.QueuedAt.Time
	}

	if model.StatusTimestamps.StartedAt.Valid {
		domain.StatusTimestamps.StartedAt = &model.StatusTimestamps.StartedAt.Time
	}

	return &domain
}

// FromDomain updates run model fields with a run domain object's fields
func (model *Plan) FromDomain(domain *ots.Plan) {
	model.ExternalID = domain.ID
	model.Model = domain.Model
	model.ResourceAdditions = domain.ResourceAdditions
	model.ResourceChanges = domain.ResourceChanges
	model.ResourceDestructions = domain.ResourceDestructions
	model.Status = domain.Status
	model.Logs = domain.Logs
	model.PlanFileBlobID = domain.PlanFileBlobID
	model.PlanJSONBlobID = domain.PlanJSONBlobID

	if domain.StatusTimestamps.CanceledAt != nil {
		model.StatusTimestamps.CanceledAt.Time = *domain.StatusTimestamps.CanceledAt
		model.StatusTimestamps.CanceledAt.Valid = true
	}

	if domain.StatusTimestamps.ErroredAt != nil {
		model.StatusTimestamps.ErroredAt.Time = *domain.StatusTimestamps.ErroredAt
		model.StatusTimestamps.ErroredAt.Valid = true
	}

	if domain.StatusTimestamps.FinishedAt != nil {
		model.StatusTimestamps.FinishedAt.Time = *domain.StatusTimestamps.FinishedAt
		model.StatusTimestamps.FinishedAt.Valid = true
	}

	if domain.StatusTimestamps.ForceCanceledAt != nil {
		model.StatusTimestamps.ForceCanceledAt.Time = *domain.StatusTimestamps.ForceCanceledAt
		model.StatusTimestamps.ForceCanceledAt.Valid = true
	}

	if domain.StatusTimestamps.QueuedAt != nil {
		model.StatusTimestamps.QueuedAt.Time = *domain.StatusTimestamps.QueuedAt
		model.StatusTimestamps.QueuedAt.Valid = true
	}

	if domain.StatusTimestamps.StartedAt != nil {
		model.StatusTimestamps.StartedAt.Time = *domain.StatusTimestamps.StartedAt
		model.StatusTimestamps.StartedAt.Valid = true
	}
}

func (l PlanList) ToDomain() (dl []*ots.Plan) {
	for _, i := range l {
		dl = append(dl, i.ToDomain())
	}
	return
}
