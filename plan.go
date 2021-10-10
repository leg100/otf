package otf

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	LocalStateFilename  = "terraform.tfstate"
	PlanFilename        = "plan.out"
	JSONPlanFilename    = "plan.out.json"
	ApplyOutputFilename = "apply.out"

	//List all available plan statuses.
	PlanCanceled    PlanStatus = "canceled"
	PlanCreated     PlanStatus = "created"
	PlanErrored     PlanStatus = "errored"
	PlanFinished    PlanStatus = "finished"
	PlanMFAWaiting  PlanStatus = "mfa_waiting"
	PlanPending     PlanStatus = "pending"
	PlanQueued      PlanStatus = "queued"
	PlanRunning     PlanStatus = "running"
	PlanUnreachable PlanStatus = "unreachable"
)

// PlanStatus represents a plan state.
type PlanStatus string

// Plan represents a Terraform Enterprise plan.
type Plan struct {
	ID string `db:"external_id"`

	Model

	Resources

	Status           PlanStatus
	StatusTimestamps TimestampMap

	// LogsBlobID is the blob ID for the log output from a terraform plan
	LogsBlobID string

	// PlanFileBlobID is the blob ID of the execution plan file in binary format
	PlanFileBlobID string

	// PlanJSONBlobID is the blob ID of the execution plan file in json format
	PlanJSONBlobID string

	RunID int64
}

// PlanStatusTimestamps holds the timestamps for individual plan statuses.
type PlanStatusTimestamps struct {
	CanceledAt      *time.Time `json:"canceled-at,omitempty"`
	ErroredAt       *time.Time `json:"errored-at,omitempty"`
	FinishedAt      *time.Time `json:"finished-at,omitempty"`
	ForceCanceledAt *time.Time `json:"force-canceled-at,omitempty"`
	QueuedAt        *time.Time `json:"queued-at,omitempty"`
	StartedAt       *time.Time `json:"started-at,omitempty"`
}

type PlanService interface {
	Get(id string) (*Plan, error)
	GetPlanJSON(id string) ([]byte, error)
}

func newPlan() *Plan {
	return &Plan{
		ID:               GenerateID("plan"),
		Model:            NewModel(),
		StatusTimestamps: make(TimestampMap),
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

func (p *Plan) UpdateStatus(status PlanStatus) {
	p.Status = status
	p.setTimestamp(status)
}

func (p *Plan) setTimestamp(status PlanStatus) {
	p.StatusTimestamps[string(status)] = time.Now()
}
