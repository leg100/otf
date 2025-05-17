// Package runstatus provides run statuses.
//
// NOTE: placed in separate package from `run` to avoid import cycles.
package runstatus

// Status represents a run state.
type Status string

const (
	// List all available run statuses supported in OTF.
	Applied            Status = "applied"
	ApplyQueued        Status = "apply_queued"
	Applying           Status = "applying"
	Canceled           Status = "canceled"
	Confirmed          Status = "confirmed"
	Discarded          Status = "discarded"
	Errored            Status = "errored"
	ForceCanceled      Status = "force_canceled"
	Pending            Status = "pending"
	PlanQueued         Status = "plan_queued"
	Planned            Status = "planned"
	PlannedAndFinished Status = "planned_and_finished"
	Planning           Status = "planning"

	// OTF doesn't support cost estimation but go-tfe API tests expect this
	// status so it is included expressly to pass the tests.
	CostEstimated Status = "cost_estimated"
)

func (s Status) String() string { return string(s) }

func All() []Status {
	return []Status{
		Applied,
		ApplyQueued,
		Applying,
		Canceled,
		Confirmed,
		Discarded,
		Errored,
		ForceCanceled,
		Pending,
		PlanQueued,
		Planned,
		PlannedAndFinished,
		Planning,
	}
}

// Done determines whether status is an end state, e.g. applied, discarded, etc.
func Done(status Status) bool {
	switch status {
	case Applied, PlannedAndFinished, Discarded, Canceled, ForceCanceled, Errored:
		return true
	default:
		return false
	}
}

// Queued determines whether status is a queued state.
func Queued(status Status) bool {
	switch status {
	case PlanQueued, ApplyQueued:
		return true
	default:
		return false
	}
}
