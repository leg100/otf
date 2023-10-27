package internal

type PhaseType string

const (
	PendingPhase PhaseType = "pending"
	PlanPhase    PhaseType = "plan"
	ApplyPhase   PhaseType = "apply"
	FinalPhase   PhaseType = "final"
	UnknownPhase PhaseType = "unknown"
)

// RunStatus represents a run state.
type RunStatus string

const (
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

	// OTF doesn't support cost estimation but go-tfe API tests expect this
	// status so it is included expressly to pass the tests.
	RunCostEstimated RunStatus = "cost_estimated"
)

func (r RunStatus) String() string { return string(r) }

func (r RunStatus) Color() string {
	switch r {
	case RunPending:
		return "yellow-50"
	case RunPlanQueued:
		return "yellow-200"
	case RunPlanning:
		return "violet-100"
	case RunPlanned:
		return "violet-400"
	case RunPlannedAndFinished:
		return "green-100"
	case RunApplyQueued:
		return "yellow-200"
	case RunApplying:
		return "cyan-200"
	case RunApplied:
		return "teal-400"
	case RunErrored:
		return "red-100"
	case RunDiscarded:
		return "gray-200"
	case RunCanceled:
		return "red-200"
	default:
		return "inherit"
	}
}

// StatusPeriod provides the percent of a run's total elapsed time in which
// it has had the given status.
type StatusPeriod struct {
	Status  RunStatus
	Percent int
}

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
)
