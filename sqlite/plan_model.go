package sqlite

import (
	"database/sql"
	"strings"

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
	StatusTimestamps     PlanStatusTimestamps `gorm:"embedded;embeddedPrefix:timestamp_"`

	Logs []byte

	// Plan belongs to a run
	RunID uint

	// The execution plan file
	Plan []byte `jsonapi:"attr,plan"`

	// The execution plan file in json format
	PlanJSON []byte `jsonapi:"attr,plan-json"`
}

// PlanStatusTimestamps holds the timestamps for individual run statuses.
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

func (r *Plan) ToDomain() *ots.Plan {
	domain := ots.Plan{
		ID:                   r.ExternalID,
		ResourceAdditions:    r.ResourceAdditions,
		ResourceChanges:      r.ResourceChanges,
		ResourceDestructions: r.ResourceDestructions,
		Status:               r.Status,
	}

	domain.StatusTimestamps: &tfe.PlanStatusTimestamps{
			CanceledAt     :  
			ErroredAt      :  
			FinishedAt     :  
			ForceCanceledAt:  
			QueuedAt       :  
			StartedAt      :  

	}

	return &domain
}

// FromDomain updates run model fields with a run domain object's fields
func (r *Plan) FromDomain(domain *ots.Plan) {
	r.ExternalID = domain.ID
	r.Status = domain.Status
	r.Message = domain.Message
	r.Refresh = domain.Refresh
	r.RefreshOnly = domain.RefreshOnly
	r.ReplaceAddrs = strings.Join(domain.ReplaceAddrs, ",")

	r.Plan.FromDomain(domain.Plan)
}

func (l PlanList) ToDomain() (dl []*ots.Plan) {
	for _, i := range l {
		dl = append(dl, i.ToDomain())
	}
	return
}
