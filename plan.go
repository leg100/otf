package ots

import (
	"fmt"

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

	Logs []byte

	// The execution plan file
	Plan []byte

	// The execution plan file in json format
	PlanJSON []byte
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
		ID:               NewPlanID(),
		StatusTimestamps: &tfe.PlanStatusTimestamps{},
	}
}

func NewPlanID() string {
	return fmt.Sprintf("plan-%s", GenerateRandomString(16))
}

// HasChanges determines whether plan has any changes (adds/changes/deletions).
func (p *Plan) HasChanges() bool {
	if p.ResourceAdditions > 0 || p.ResourceChanges > 0 || p.ResourceDestructions > 0 {
		return true
	}
	return false
}
