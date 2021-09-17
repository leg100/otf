package ots

import (
	"encoding/json"
	"fmt"

	tfe "github.com/leg100/go-tfe"
	"gorm.io/gorm"
)

const (
	LocalStateFilename  = "terraform.tfstate"
	PlanFilename        = "plan.out"
	JSONPlanFilename    = "plan.out.json"
	ApplyOutputFilename = "apply.out"
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

	// LogsBlobID is the blob ID for the log output from a terraform plan
	LogsBlobID string

	// PlanFileBlobID is the blob ID of the execution plan file in binary format
	PlanFileBlobID string

	// PlanJSONBlobID is the blob ID of the execution plan file in json format
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
		LogsBlobID:       NewBlobID(),
		PlanFileBlobID:   NewBlobID(),
		PlanJSONBlobID:   NewBlobID(),
	}
}

// HasChanges determines whether plan has any changes (adds/changes/deletions).
func (p *Plan) HasChanges() bool {
	if p.ResourceAdditions > 0 || p.ResourceChanges > 0 || p.ResourceDestructions > 0 {
		return true
	}
	return false
}

func (p *Plan) GetLogsBlobID() string {
	return p.LogsBlobID
}

func (p *Plan) Do(run *Run, exe *Executor) error {
	if err := exe.RunCLI("terraform", "plan", "-no-color", fmt.Sprintf("-out=%s", PlanFilename)); err != nil {
		return err
	}

	if err := exe.RunCLI("sh", "-c", fmt.Sprintf("terraform show -json %s > %s", PlanFilename, JSONPlanFilename)); err != nil {
		return err
	}

	if err := exe.RunFunc(run.uploadPlan); err != nil {
		return err
	}

	if err := exe.RunFunc(run.uploadJSONPlan); err != nil {
		return err
	}

	return nil
}

// UpdateResources parses the plan file produced from terraform plan to
// determine the number and type of resource changes planned and updates the
// plan object accordingly.
func (p *Plan) UpdateResources(bs BlobStore) error {
	jsonFile, err := bs.Get(p.PlanJSONBlobID)
	if err != nil {
		return err
	}

	planFile := PlanFile{}
	if err := json.Unmarshal(jsonFile, &planFile); err != nil {
		return err
	}

	// Parse plan output
	adds, updates, deletes := planFile.Changes()

	// Update status
	p.ResourceAdditions = adds
	p.ResourceChanges = updates
	p.ResourceDestructions = deletes

	return nil
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
