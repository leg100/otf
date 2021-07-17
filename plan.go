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

	HasChanges           bool
	LogReadURL           string
	ResourceAdditions    int
	ResourceChanges      int
	ResourceDestructions int
	Status               tfe.PlanStatus
	StatusTimestamps     *tfe.PlanStatusTimestamps `gorm:"embedded;embeddedPrefix:timestamp_"`

	Logs []byte

	RunID uint
}

func (a *Plan) DTO() interface{} {
	return &tfe.Plan{
		ID:                   a.ExternalID,
		HasChanges:           a.HasChanges,
		LogReadURL:           a.LogReadURL,
		ResourceAdditions:    a.ResourceAdditions,
		ResourceChanges:      a.ResourceChanges,
		ResourceDestructions: a.ResourceDestructions,
		Status:               a.Status,
		StatusTimestamps:     a.StatusTimestamps,
	}
}

type PlanService interface {
	Get(id string) (*Plan, error)
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
