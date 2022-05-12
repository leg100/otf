package sql

import (
	"time"

	"github.com/leg100/otf"
)

type planResult interface {
	GetPlanID() *string
	Timestamps
	GetStatus() *string
}

type planStatusTimestamp interface {
	GetPlanID() *string
	GetStatus() *string
	GetTimestamp() time.Time
}

func addResultToPlan(plan *otf.Plan, result planResult) {
	plan.ID = *result.GetPlanID()
	plan.Timestamps = convertTimestamps(result)
	plan.Status = otf.PlanStatus(*result.GetStatus())
}

func convertPlan(result planResult) *otf.Plan {
	var plan otf.Plan
	addResultToPlan(&plan, result)
	return &plan
}

func convertPlanStatusTimestamp(r planStatusTimestamp) otf.PlanStatusTimestamp {
	return otf.PlanStatusTimestamp{
		Status:    otf.PlanStatus(*r.GetStatus()),
		Timestamp: r.GetTimestamp(),
	}
}
