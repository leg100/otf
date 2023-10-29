package run

import "time"

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

type (
	// StatusPeriod is the duration over which a run has had a status.
	StatusPeriod struct {
		Status RunStatus     `json:"status"`
		Period time.Duration `json:"period"`
	}

	PeriodReport struct {
		TotalTime time.Duration  `json:"total_time"`
		Periods   []StatusPeriod `json:"periods"`
	}
)

func (r PeriodReport) Percentage(i int) float64 {
	return (r.Periods[i].Period.Seconds() / r.TotalTime.Seconds()) * 100
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
