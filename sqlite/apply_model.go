package sqlite

import (
	"database/sql"

	"github.com/leg100/go-tfe"
	"github.com/leg100/otf"
	"gorm.io/gorm"
)

// Apply models a row in an applies table.
type Apply struct {
	gorm.Model

	ExternalID string `gorm:"uniqueIndex"`

	otf.Resources

	Status           tfe.ApplyStatus
	StatusTimestamps *ApplyStatusTimestamps `gorm:"embedded;embeddedPrefix:timestamp_"`

	LogsBlobID string

	// Apply belongs to a run
	RunID uint
}

// ApplyList is a list of run models
type ApplyList []Apply

// ApplyStatusTimestamps holds the timestamps for individual apply statuses.
type ApplyStatusTimestamps struct {
	CanceledAt      sql.NullTime
	ErroredAt       sql.NullTime
	FinishedAt      sql.NullTime
	ForceCanceledAt sql.NullTime
	QueuedAt        sql.NullTime
	StartedAt       sql.NullTime
}

// Update updates the model with the supplied fn. The fn operates on the domain
// obj, so Update handles converting to and from a domain obj.
func (model *Apply) Update(fn func(*otf.Apply) error) error {
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

func (model *Apply) ToDomain() *otf.Apply {
	domain := otf.Apply{
		ID:               model.ExternalID,
		Model:            model.Model,
		Resources:        model.Resources,
		Status:           model.Status,
		StatusTimestamps: &tfe.ApplyStatusTimestamps{},
		LogsBlobID:       model.LogsBlobID,
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
func (model *Apply) FromDomain(domain *otf.Apply) {
	model.ExternalID = domain.ID
	model.Model = domain.Model
	model.ResourceAdditions = domain.ResourceAdditions
	model.ResourceChanges = domain.ResourceChanges
	model.ResourceDestructions = domain.ResourceDestructions
	model.Status = domain.Status
	model.LogsBlobID = domain.LogsBlobID

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

func (l ApplyList) ToDomain() (dl []*otf.Apply) {
	for _, i := range l {
		dl = append(dl, i.ToDomain())
	}
	return
}
