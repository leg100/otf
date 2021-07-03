package ots

import (
	"fmt"

	tfe "github.com/leg100/go-tfe"
)

const (
	MaxPlanLogsLimit = 65536
)

type PlanService interface {
	GetPlan(id string) (*tfe.Plan, error)
	UpdatePlanStatus(id string, status tfe.PlanStatus) (*tfe.Plan, error)
	FinishPlan(id string, opts *PlanFinishOptions) (*tfe.Plan, error)
	GetPlanLogs(id string, opts PlanLogOptions) ([]byte, error)
	UploadPlanLogs(id string, logs []byte) error
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

func NewPlanID() string {
	return fmt.Sprintf("plan-%s", GenerateRandomString(16))
}
