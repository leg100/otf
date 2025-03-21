package workspace

import (
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runstatus"
)

// LatestRun is a summary of the latest run for a workspace
type LatestRun struct {
	ID     resource.TfeID
	Status runstatus.Status
}
