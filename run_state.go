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

type Phase string

const (
	PlanPhase Phase = iota
	ApplyPhase
)

type runState struct {
	status     RunStatus
	phase      Phase     // phase type
	phaseState JobStatus // phase state
	final      bool
}

var (
	RunPendingState = runState{
		status:     RunPending,
		phase:      PlanPhase,
		phaseState: JobPending,
	}
	RunPlanQueuedState = runState{
		status:     RunPlanQueued,
		phase:      PlanPhase,
		phaseState: JobQueued,
	}
	RunPlannedAndFinishedState = runState{
		status:     RunPlannedAndFinished,
		phase:      PlanPhase,
		phaseState: JobFinished,
		final:      true,
	}
	RunApplyingState = runState{
		status:     RunApplying,
		phase:      ApplyPhase,
		phaseState: JobRunning,
	}
	RunErrorState = runState{
		status:     RunApplying,
		phase:      ApplyPhase,
		phaseState: JobErrored,
	}
)

type RunStatus string

func (r RunStatus) String() string { return string(r) }
