package workspace

type (
	// LatestRun is a summary of the latest run for a workspace
	LatestRun struct {
		ID     string
		Status runStatus
	}

	// runStatus is the status of a run. Duplicated here rather than use
	// run.runStatus in order to avoid an import cycle.
	runStatus string
)

func (s runStatus) String() string { return string(s) }
