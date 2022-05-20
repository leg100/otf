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
	ChunkStore
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

	// run is the parent run
	run *Run
}

type ApplyStatusTimestamp struct {
	Status    ApplyStatus
	Timestamp time.Time
}

func (a *Apply) GetID() string     { return a.ID }
func (a *Apply) GetStatus() string { return string(a.Status) }
func (a *Apply) String() string    { return a.ID }

func (a *Apply) GetService(app Application) JobService {
	return app.ApplyService()
}

func newApply(run *Run) *Apply {
	return &Apply{
		ID:     NewID("apply"),
		run:    run,
		Status: ApplyPending,
	}
}

// Do performs a terraform apply
func (a *Apply) Do(env Environment) error {
	if err := a.run.setupEnv(env); err != nil {
		return err
	}

	if err := env.RunFunc(a.run.downloadPlanFile); err != nil {
		return err
	}

	if err := env.RunCLI("sh", "-c", fmt.Sprintf("terraform apply %s | tee %s", PlanFilename, ApplyOutputFilename)); err != nil {
		return err
	}

	if err := env.RunFunc(a.run.uploadState); err != nil {
		return err
	}

	return nil
}

// Start updates the run to reflect its apply having started
func (a *Apply) Start() error {
	if a.run.Status() == RunApplying {
		return ErrJobAlreadyClaimed
	}

	if a.run.Status() != RunApplyQueued {
		return fmt.Errorf("run cannot be started: invalid status: %s", a.run.Status())
	}

	a.run.UpdateStatus(RunApplying)

	return nil
}

// Finish updates the run to reflect its apply having finished. An event is
// returned reflecting the run's new status.
func (a *Apply) Finish() error {
	return a.run.UpdateStatus(RunApplied)
}
