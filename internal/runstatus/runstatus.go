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
	PolicyChecking     Status = "policy_checking"
	PolicyChecked      Status = "policy_checked"
	PolicySoftFailed   Status = "policy_soft_failed"
	PolicyFailed       Status = "policy_failed"
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
		PolicyChecking,
		PolicyChecked,
		PolicySoftFailed,
		PolicyFailed,
		Planned,
		PlannedAndFinished,
		Planning,
	}
}

// PlannedCompatible determines whether status represents a completed plan that
// is still compatible with planned-era callers and filters.
func PlannedCompatible(status Status) bool {
	switch status {
	case Planned, CostEstimated, PolicyChecked, PolicySoftFailed:
		return true
	default:
		return false
	}
}

// ExpandPlannedCompatible expands planned to the set of statuses that should
// be treated as planned-compatible. All other statuses are returned unchanged.
func ExpandPlannedCompatible(statuses []Status) []Status {
	if len(statuses) == 0 {
		return nil
	}

	seen := make(map[Status]struct{}, len(statuses))
	expanded := make([]Status, 0, len(statuses)+3)
	appendStatus := func(status Status) {
		if _, ok := seen[status]; ok {
			return
		}
		seen[status] = struct{}{}
		expanded = append(expanded, status)
	}

	for _, status := range statuses {
		if status == Planned {
			for _, compatible := range []Status{Planned, CostEstimated, PolicyChecked, PolicySoftFailed} {
				appendStatus(compatible)
			}
			continue
		}
		appendStatus(status)
	}

	return expanded
}

// Done determines whether status is an end state, e.g. applied, discarded, etc.
func Done(status Status) bool {
	switch status {
	case Applied, PlannedAndFinished, PolicyFailed, Discarded, Canceled, ForceCanceled, Errored:
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
