package otf

const (
	PendingPhase PhaseType = "pending"
	PlanPhase    PhaseType = "plan"
	ApplyPhase   PhaseType = "apply"
	FinalPhase   PhaseType = "final"
	UnknownPhase PhaseType = "unknown"

	// List all available run statuses supported in OTF.
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

var (
	ActiveRun = []RunStatus{
		RunApplyQueued,
		RunApplying,
		RunConfirmed,
		RunPlanQueued,
		RunPlanned,
		RunPlanning,
	}
	IncompleteRun = append(ActiveRun, RunPending)
	CompletedRun  = []RunStatus{
		RunApplied,
		RunErrored,
		RunDiscarded,
		RunCanceled,
		RunForceCanceled,
	}
)

type (
	PhaseType string

	// RunStatus represents a run state.
	RunStatus string
)

func (r RunStatus) String() string { return string(r) }

// RunStatusPtr returns a pointer to a run status
func RunStatusPtr(s RunStatus) *RunStatus {
	return &s
}
