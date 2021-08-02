package sqlite

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"gorm.io/gorm"
)

// Apply models a row in an applies table.
type Apply struct {
	gorm.Model

	ExternalID string `gorm:"uniqueIndex"`

	ResourceAdditions    int
	ResourceChanges      int
	ResourceDestructions int
	Status               tfe.ApplyStatus
	StatusTimestamps     tfe.ApplyStatusTimestamps `gorm:"embedded;embeddedPrefix:timestamp_"`

	Logs []byte

	// Apply belongs to a run
	RunID uint
}

// ApplyList is a list of run models
type ApplyList []Apply

// Update updates the model with the supplied fn. The fn operates on the domain
// obj, so Update handles converting to and from a domain obj.
func (r *Apply) Update(fn func(*ots.Apply) error) error {
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

func (r *Apply) ToDomain() *ots.Apply {
	domain := ots.Apply{
		ID:                   r.ExternalID,
		Model:                r.Model,
		ResourceAdditions:    r.ResourceAdditions,
		ResourceChanges:      r.ResourceChanges,
		ResourceDestructions: r.ResourceDestructions,
		Status:               r.Status,
		StatusTimestamps:     &r.StatusTimestamps,
		Logs:                 r.Logs,
	}

	return &domain
}

// FromDomain updates run model fields with a run domain object's fields
func (r *Apply) FromDomain(domain *ots.Apply) {
	r.ExternalID = domain.ID
	r.Model = domain.Model
	r.ResourceAdditions = domain.ResourceAdditions
	r.ResourceChanges = domain.ResourceChanges
	r.ResourceDestructions = domain.ResourceDestructions
	r.Status = domain.Status
	r.StatusTimestamps = *domain.StatusTimestamps
	r.Logs = domain.Logs
}

func (l ApplyList) ToDomain() (dl []*ots.Apply) {
	for _, i := range l {
		dl = append(dl, i.ToDomain())
	}
	return
}
