package runner

type RunnerStatus string

const (
	RunnerIdle    RunnerStatus = "idle"
	RunnerBusy    RunnerStatus = "busy"
	RunnerExited  RunnerStatus = "exited"
	RunnerErrored RunnerStatus = "errored"
	RunnerUnknown RunnerStatus = "unknown"
)
