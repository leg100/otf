package run

import (
	"time"

	"github.com/leg100/otf/internal/runstatus"
)

type (
	// StatusPeriod is the duration over which a run has had a status.
	StatusPeriod struct {
		Status runstatus.Status `json:"status"`
		Period time.Duration    `json:"period"`
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
	ActiveRun = []runstatus.Status{
		runstatus.ApplyQueued,
		runstatus.Applying,
		runstatus.Confirmed,
		runstatus.PlanQueued,
		runstatus.Planned,
		runstatus.Planning,
	}
	IncompleteRun = append(ActiveRun, runstatus.Pending)
)
