package runner

import "github.com/leg100/otf/internal/resource"

type RunnerEvent struct {
	ID     resource.TfeID `json:"runner_id"`
	Status RunnerStatus   `json:"status"`
}

type JobEvent struct {
	ID resource.TfeID `json:"job_id"`
	// ID of the run that this job is for.
	RunID resource.TfeID `jsonapi:"attribute" json:"run_id" db:"run_id"`
	// Current status of job.
	Status JobStatus `json:"status"`
	// ID of runner that this job is allocated to. Only set once job enters
	// JobAllocated state.
	RunnerID *resource.TfeID `json:"runner_id"`
	// Signaled is non-nil when a cancelation signal has been sent to the job
	// and it is true when it has been forceably canceled.
	Signaled *bool `json:"signaled"`
}
