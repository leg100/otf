package workspace

import (
	"time"

	"github.com/leg100/otf"
)

// run is a workspace run, a slimmed down version of run.Run, redefined here in
// order to avoid an import cycle on the run pkg.
type run struct {
	ID                     string
	CreatedAt              time.Time
	IsDestroy              bool
	ForceCancelAvailableAt *time.Time
	Message                string
	Organization           string
	Refresh                bool
	RefreshOnly            bool
	ReplaceAddrs           []string
	PositionInQueue        int
	TargetAddrs            []string
	AutoApply              bool
	Speculative            bool
	Status                 otf.RunStatus
	WorkspaceID            string
	ConfigurationVersionID string

	Latest bool    // is latest run for workspace
	Commit *string // commit sha that triggered this run
}

func (r run) PlanOnly() bool {
	return r.Status == otf.RunPlannedAndFinished
}

// Cancelable determines whether run can be cancelled.
func (r run) Cancelable() bool {
	switch r.Status {
	case otf.RunPending, otf.RunPlanQueued, otf.RunPlanning, otf.RunPlanned, otf.RunApplyQueued, otf.RunApplying:
		return true
	default:
		return false
	}
}
