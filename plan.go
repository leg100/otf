package ots

import (
	tfe "github.com/leg100/go-tfe"
	"gorm.io/gorm"
)

const (
	MaxPlanLogsLimit = 65536
)

// Plan represents a Terraform Enterprise plan.
type Plan struct {
	ID string

	gorm.Model

	ResourceAdditions    int
	ResourceChanges      int
	ResourceDestructions int
	Status               tfe.PlanStatus
	StatusTimestamps     *tfe.PlanStatusTimestamps

	Logs

	// The blob ID of the execution plan file
	PlanFileBlobID string

	// The blob ID of the execution plan file in json format
	PlanJSONBlobID string
}

type PlanService interface {
	Get(id string) (*Plan, error)
	GetPlanJSON(id string) ([]byte, error)
}

// PlanFinishOptions represents the options for finishing a plan.
type PlanFinishOptions struct {
	// Type is a public field utilized by JSON:API to set the resource type via
	// the field tag.  It is not a user-defined value and does not need to be
	// set.  https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,plans"`

	ResourceAdditions    int `jsonapi:"attr,resource-additions"`
	ResourceChanges      int `jsonapi:"attr,resource-changes"`
	ResourceDestructions int `jsonapi:"attr,resource-destructions"`
}

func newPlan() *Plan {
	return &Plan{
		ID:               GenerateID("plan"),
		StatusTimestamps: &tfe.PlanStatusTimestamps{},
	}
}

// HasChanges determines whether plan has any changes (adds/changes/deletions).
func (p *Plan) HasChanges() bool {
	if p.ResourceAdditions > 0 || p.ResourceChanges > 0 || p.ResourceDestructions > 0 {
		return true
	}
	return false
}

func (p *Plan) UpdateStatus(status tfe.PlanStatus) {
	p.Status = status
	p.setTimestamp(status)
}

func (p *Plan) setTimestamp(status tfe.PlanStatus) {
	switch status {
	case tfe.PlanCanceled:
		p.StatusTimestamps.CanceledAt = TimeNow()
	case tfe.PlanErrored:
		p.StatusTimestamps.ErroredAt = TimeNow()
	case tfe.PlanFinished:
		p.StatusTimestamps.FinishedAt = TimeNow()
	case tfe.PlanQueued:
		p.StatusTimestamps.QueuedAt = TimeNow()
	case tfe.PlanRunning:
		p.StatusTimestamps.StartedAt = TimeNow()
	}
}
