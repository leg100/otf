package sql

import (
	"time"

	"github.com/leg100/otf"
)

func convertPlan(row Plans) *otf.Plan {
	plan := otf.Plan{}
	plan.ID               =     *row.PlanID
	plan.CreatedAt            = row.CreatedAt            
	plan.UpdatedAt            = row.UpdatedAt            
	plan.ResourceAdditions    = int(*row.ResourceAdditions    )
	plan.ResourceChanges      = int(*row.ResourceChanges      )
	plan.ResourceDestructions = int(*row.ResourceDestructions )
	plan.Status               = otf.PlanStatus(*row.Status               )
	plan.RunID                = *row.RunID                

	for _, p := range row.StatusTimestamps
	plan.StatusTimestamps     = row.StatusTimestamps     
}

func convertPlanStatusTimestamp(r planStatusTimestamp) otf.PlanStatusTimestamp {
	return otf.PlanStatusTimestamp{
		Status:    otf.PlanStatus(*r.GetStatus()),
		Timestamp: r.GetTimestamp(),
	}
}
