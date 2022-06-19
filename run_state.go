package otf

import "errors"

const (
	RunApplied            RunStatus = "applied"
	RunApplyQueued        RunStatus = "apply_queued"
	RunApplying           RunStatus = "applying"
	RunCanceled           RunStatus = "canceled"
	RunForceCanceled      RunStatus = "force_canceled"
	RunConfirmed          RunStatus = "confirmed"
	RunDiscarded          RunStatus = "discarded"
	RunErrored            RunStatus = "errored"
	RunPending            RunStatus = "pending"
	RunPlanQueued         RunStatus = "plan_queued"
	RunPlanned            RunStatus = "planned"
	RunPlannedAndFinished RunStatus = "planned_and_finished"
	RunPlanning           RunStatus = "planning"
)

var ErrRunInvalidStateTransition = errors.New("invalid run state transition")

type RunStatus string

func (r RunStatus) String() string { return string(r) }

type runState interface {
	Status() RunStatus
	// Start run action (either a plan or an apply)
	Start() error
	// Finish run action (either a plan or apply)
	Finish(ReportService, JobFinishOptions) error
	// Cancel run action
	Cancel() error
	// Discard run action
	Discard() error
	// Enqueue action (either an plan or apply depending on current state)
	Enqueue() error
	// Apply run action
	ApplyRun() error
	// Cancelable determines whether run is in a state to be canceled
	Cancelable() bool
	// Confirmable determines whether run is in a state to be confirmed
	Confirmable() bool
	// Discardable determines whether run is in a state to be discarded
	Discardable() bool
	// Done determines whether run is in a final completed state.
	Done() bool
	// run state can have a job associated with it
	Job
}
