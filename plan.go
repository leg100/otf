package otf

import (
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
	StatusTimestamps []PlanStatusTimestamp

	// RunID is the ID of the Run the Plan belongs to.
	RunID string
}

type PlanStatusTimestamp struct {
	Status    PlanStatus
	Timestamp time.Time
}

func (p *Plan) GetID() string     { return p.ID }
func (p *Plan) GetStatus() string { return string(p.Status) }
func (p *Plan) String() string    { return p.ID }

type PlanService interface {
	Get(id string) (*Plan, error)

	JobService
}

type PlanLogStore interface {
	ChunkStore
}

func newPlan(runID string) *Plan {
	return &Plan{
		ID:         NewID("plan"),
		Timestamps: NewTimestamps(),
		RunID:      runID,
		// new plans always start off in pending state
		Status: PlanPending,
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
		if err := run.UpdateStatus(RunPlannedAndFinished); err != nil {
			return nil, err
		}
		return &Event{Payload: run, Type: EventRunPlannedAndFinished}, nil
	}

	if !run.Workspace.AutoApply {
		if err := run.UpdateStatus(RunPlanned); err != nil {
			return nil, err
		}
		return &Event{Payload: run, Type: EventRunPlanned}, nil
	}

	if err := run.UpdateStatus(RunApplyQueued); err != nil {
		return nil, err
	}
	return &Event{Type: EventApplyQueued, Payload: run}, nil
}
