package run

import "time"

// Status represents a run state.
type Status string

const (
	// List all available run statuses supported in OTF.
	RunApplied            Status = "applied"
	RunApplyQueued        Status = "apply_queued"
	RunApplying           Status = "applying"
	RunCanceled           Status = "canceled"
	RunForceCanceled      Status = "force_canceled"
	RunConfirmed          Status = "confirmed"
	RunDiscarded          Status = "discarded"
	RunErrored            Status = "errored"
	RunPending            Status = "pending"
	RunPlanQueued         Status = "plan_queued"
	RunPlanned            Status = "planned"
	RunPlannedAndFinished Status = "planned_and_finished"
	RunPlanning           Status = "planning"

	// OTF doesn't support cost estimation but go-tfe API tests expect this
	// status so it is included expressly to pass the tests.
	RunCostEstimated Status = "cost_estimated"
)

func (r Status) String() string { return string(r) }

type (
	// StatusPeriod is the duration over which a run has had a status.
	StatusPeriod struct {
		Status Status        `json:"status"`
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
	ActiveRun = []Status{
		RunApplyQueued,
		RunApplying,
		RunConfirmed,
		RunPlanQueued,
		RunPlanned,
		RunPlanning,
	}
	IncompleteRun = append(ActiveRun, RunPending)
)
