package otf

import (
	"fmt"
	"time"
)

//List all available apply statuses supported in OTF.
const (
	ApplyCanceled    ApplyStatus = "canceled"
	ApplyCreated     ApplyStatus = "created"
	ApplyErrored     ApplyStatus = "errored"
	ApplyFinished    ApplyStatus = "finished"
	ApplyPending     ApplyStatus = "pending"
	ApplyQueued      ApplyStatus = "queued"
	ApplyRunning     ApplyStatus = "running"
	ApplyUnreachable ApplyStatus = "unreachable"
)

// ApplyStatus represents an apply state.
type ApplyStatus string

// ApplyService allows interaction with Applies
type ApplyService interface {
	Get(id string) (*Apply, error)

	JobService
}

type ApplyLogStore interface {
	ChunkStore
}

// Apply represents a terraform apply
type Apply struct {
	ID string `json:"apply_id"`

	// ResourcesReport is a report of applied resource changes
	*ResourceReport

	// Status is the current status
	Status ApplyStatus `json:"apply_status"`

	// StatusTimestamps records timestamps of status transitions
	StatusTimestamps []ApplyStatusTimestamp `json:"apply_status_timestamps"`

	// RunID is the ID of the Run the Apply belongs to.
	RunID string `json:"run_id"`
}

type ApplyStatusTimestamp struct {
	Status    ApplyStatus
	Timestamp time.Time
}

func (a *Apply) GetID() string     { return a.ID }
func (a *Apply) GetStatus() string { return string(a.Status) }
func (a *Apply) String() string    { return a.ID }

func newApply(runID string) *Apply {
	return &Apply{
		ID:    NewID("apply"),
		RunID: runID,
	}
}

// Do performs a terraform apply
func (a *Apply) Do(run *Run, env Environment) error {
	if err := run.Do(env); err != nil {
		return err
	}

	if err := env.RunFunc(run.downloadPlanFile); err != nil {
		return err
	}

	if err := env.RunCLI("sh", "-c", fmt.Sprintf("terraform apply %s | tee %s", PlanFilename, ApplyOutputFilename)); err != nil {
		return err
	}

	if err := env.RunFunc(run.uploadState); err != nil {
		return err
	}

	return nil
}

// Start updates the run to reflect its apply having started
func (a *Apply) Start(run *Run) error {
	if run.Status != RunApplyQueued {
		return fmt.Errorf("run cannot be started: invalid status: %s", run.Status)
	}

	run.UpdateStatus(RunApplying)

	return nil
}

// Finish updates the run to reflect its apply having finished. An event is
// returned reflecting the run's new status.
func (a *Apply) Finish(run *Run) error {
	return run.UpdateStatus(RunApplied)
}

func (a *Apply) updateStatus(status ApplyStatus) {
	a.Status = status
}
