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

// Apply represents a terraform apply
type Apply struct {
	id string
	// ResourcesReport is a report of applied resource changes
	*ResourceReport
	// Status is the current status
	status ApplyStatus
	// StatusTimestamps records timestamps of status transitions
	statusTimestamps []ApplyStatusTimestamp
	// run is the parent run
	run *Run
}

func (a *Apply) ID() string          { return a.id }
func (a *Apply) JobID() string       { return a.id }
func (a *Apply) String() string      { return a.id }
func (a *Apply) Status() ApplyStatus { return a.status }

func (a *Apply) GetService(app Application) JobService {
	return app.ApplyService()
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
	a.run.updateStatus(RunApplying)
	return nil
}

// Finish updates the run to reflect its apply having finished. An event is
// returned reflecting the run's new status.
func (a *Apply) Finish() error {
	return a.run.updateStatus(RunApplied)
}

func (a *Apply) StatusTimestamps() []ApplyStatusTimestamp { return a.statusTimestamps }

func (a *Apply) updateStatus(status ApplyStatus) {
	a.status = status
	a.statusTimestamps = append(a.statusTimestamps, ApplyStatusTimestamp{
		Status:    status,
		Timestamp: CurrentTimestamp(),
	})
}

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

type ApplyStatusTimestamp struct {
	Status    ApplyStatus
	Timestamp time.Time
}

func newApply(run *Run) *Apply {
	return &Apply{
		id:             NewID("apply"),
		run:            run,
		status:         ApplyPending,
		ResourceReport: &ResourceReport{},
	}
}
