package ots

import (
	"fmt"
	"time"

	tfe "github.com/leg100/go-tfe"
)

const (
	MaxPlanLogsLimit = 65536
)

// Plan represents a Terraform Enterprise plan.
type Plan struct {
	ExternalID string `gorm:"uniqueIndex"`
	InternalID uint   `gorm:"primaryKey;column:id"`

	ResourceAdditions    int
	ResourceChanges      int
	ResourceDestructions int
	Status               tfe.PlanStatus
	StatusTimestamps     *tfe.PlanStatusTimestamps `gorm:"embedded;embeddedPrefix:timestamp_"`

	Logs []byte

	RunID uint

	// The execution plan file
	Plan []byte `jsonapi:"attr,plan"`

	// The execution plan file in json format
	PlanJSON []byte `jsonapi:"attr,plan-json"`
}

func (p *Plan) DTO() interface{} {
	return &tfe.Plan{
		ID:                   p.ExternalID,
		HasChanges:           p.HasChanges(),
		LogReadURL:           GetPlanLogsUrl(p.ExternalID),
		ResourceAdditions:    p.ResourceAdditions,
		ResourceChanges:      p.ResourceChanges,
		ResourceDestructions: p.ResourceDestructions,
		Status:               p.Status,
		StatusTimestamps:     p.StatusTimestamps,
	}
}

type PlanService interface {
	Get(id string) (*Plan, error)
	GetPlanJSON(id string) ([]byte, error)
}

type PlanLogOptions struct {
	// The maximum number of bytes of logs to return to the client
	Limit int `schema:"limit"`

	// The start position in the logs from which to send to the client
	Offset int `schema:"offset"`
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

	// The execution plan file
	Plan []byte `jsonapi:"attr,plan"`

	// The execution plan file in json format
	PlanJSON []byte `jsonapi:"attr,plan-json"`
}

func newPlan() *Plan {
	return &Plan{
		ExternalID: NewPlanID(),
	}
}

func NewPlanID() string {
	return fmt.Sprintf("plan-%s", GenerateRandomString(16))
}

// UpdateStatus updates the status of the plan. It'll also update the
// appropriate timestamp and set any other appropriate fields for the given
// status.
func (p *Plan) UpdateStatus(status tfe.PlanStatus) {
	// Copy timestamps from plan
	timestamps := &tfe.PlanStatusTimestamps{}
	if p.StatusTimestamps != nil {
		timestamps = p.StatusTimestamps
	}

	switch status {
	case tfe.PlanQueued:
		timestamps.QueuedAt = time.Now()
	case tfe.PlanCanceled:
		timestamps.CanceledAt = time.Now()
	case tfe.PlanErrored:
		timestamps.ErroredAt = time.Now()
	case tfe.PlanFinished:
		timestamps.FinishedAt = time.Now()
	default:
		// Don't set a timestamp
		return
	}

	p.Status = status

	// Set timestamps on plan
	p.StatusTimestamps = timestamps
}

// HasChanges determines whether plan has any changes (adds/changes/deletions).
func (p *Plan) HasChanges() bool {
	if p.ResourceAdditions > 0 || p.ResourceChanges > 0 || p.ResourceDestructions > 0 {
		return true
	}
	return false
}
