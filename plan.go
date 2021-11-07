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
	ID string `db:"plan_id"`

	// Timestamps records timestamps of lifecycle transitions
	Timestamps

	// Resources is a summary of planned resource changes
	Resources

	// Status is the current status
	Status PlanStatus

	// StatusTimestamps records timestamps of status transitions
	StatusTimestamps TimestampMap

	// PlanFile is the blob ID of the execution plan file in binary format
	PlanFile []byte

	// PlanJSON is the blob ID of the execution plan file in json format
	PlanJSON []byte

	// RunID is the ID of the Run the Plan belongs to.
	RunID string
}

func (p *Plan) GetID() string     { return p.ID }
func (p *Plan) GetStatus() string { return string(p.Status) }
func (p *Plan) String() string    { return p.ID }

type PlanService interface {
	Get(id string) (*Plan, error)
	GetPlanJSON(id string) ([]byte, error)

	JobService
}

func newPlan(runID string) *Plan {
	return &Plan{
		ID:               NewID("plan"),
		Timestamps:       NewTimestamps(),
		StatusTimestamps: make(TimestampMap),
		RunID:            runID,
	}
}

// HasChanges determines whether plan has any changes (adds/changes/deletions).
func (p *Plan) HasChanges() bool {
	if p.ResourceAdditions > 0 || p.ResourceChanges > 0 || p.ResourceDestructions > 0 {
		return true
	}
	return false
}

// Do performs a terraform plan
func (p *Plan) Do(run *Run, env Environment) error {
	if err := run.Do(env); err != nil {
		return err
	}

	if err := env.RunCLI("terraform", "plan", fmt.Sprintf("-out=%s", PlanFilename)); err != nil {
		return err
	}

	if err := env.RunCLI("sh", "-c", fmt.Sprintf("terraform show -json %s > %s", PlanFilename, JSONPlanFilename)); err != nil {
		return err
	}

	if err := env.RunFunc(run.uploadPlan); err != nil {
		return err
	}

	if err := env.RunFunc(run.uploadJSONPlan); err != nil {
		return err
	}

	return nil
}

// CalculateTotals produces a summary of planned changes and updates the object
// with the summary.
func (p *Plan) CalculateTotals() error {
	if p.PlanJSON == nil {
		return fmt.Errorf("plan obj is missing the json formatted plan file")
	}

	planFile := PlanFile{}
	if err := json.Unmarshal(p.PlanJSON, &planFile); err != nil {
		return err
	}

	// Parse plan output
	p.Resources = planFile.Changes()

	return nil
}

// Start updates the plan to reflect its plan having started
func (p *Plan) Start(run *Run) error {
	if run.Status != RunPlanQueued {
		return fmt.Errorf("run cannot be started: invalid status: %s", run.Status)
	}

	run.UpdateStatus(RunPlanning)

	return nil
}

// Finish updates the run to reflect its plan having finished. An event is
// returned reflecting the run's new status.
func (p *Plan) Finish(run *Run) (*Event, error) {
	if !p.HasChanges() || run.IsSpeculative() {
		run.UpdateStatus(RunPlannedAndFinished)
		return &Event{Payload: run, Type: EventRunPlannedAndFinished}, nil
	}

	run.UpdateStatus(RunPlanned)

	if run.Workspace.AutoApply {
		run.UpdateStatus(RunApplyQueued)
		return &Event{Type: EventApplyQueued, Payload: run}, nil
	}

	return &Event{Payload: run, Type: EventRunPlanned}, nil
}

func (p *Plan) UpdateStatus(status PlanStatus) {
	p.Status = status
	p.setTimestamp(status)
}

func (p *Plan) setTimestamp(status PlanStatus) {
	p.StatusTimestamps[string(status)] = time.Now()
}
