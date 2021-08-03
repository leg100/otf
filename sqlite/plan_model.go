package sqlite

import (
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
	StatusTimestamps     tfe.PlanStatusTimestamps `gorm:"embedded;embeddedPrefix:timestamp_"`

	Logs []byte

	// The blob ID of the execution plan file
	PlanFileBlobID string

	// The blob ID of the execution plan file in json format
	PlanJSONBlobID string

	// Plan belongs to a run
	RunID uint
}

// PlanList is a list of run models
type PlanList []Plan

// Update updates the model with the supplied fn. The fn operates on the domain
// obj, so Update handles converting to and from a domain obj.
func (r *Plan) Update(fn func(*ots.Plan) error) error {
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

func (model *Plan) ToDomain() *ots.Plan {
	domain := ots.Plan{
		ID:                   model.ExternalID,
		Model:                model.Model,
		ResourceAdditions:    model.ResourceAdditions,
		ResourceChanges:      model.ResourceChanges,
		ResourceDestructions: model.ResourceDestructions,
		Status:               model.Status,
		StatusTimestamps:     &model.StatusTimestamps,
		Logs:                 model.Logs,
		PlanFileBlobID:       model.PlanFileBlobID,
		PlanJSONBlobID:       model.PlanJSONBlobID,
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
	model.StatusTimestamps = *domain.StatusTimestamps
	model.Logs = domain.Logs
	model.PlanFileBlobID = domain.PlanFileBlobID
	model.PlanJSONBlobID = domain.PlanJSONBlobID
}

func (l PlanList) ToDomain() (dl []*ots.Plan) {
	for _, i := range l {
		dl = append(dl, i.ToDomain())
	}
	return
}
