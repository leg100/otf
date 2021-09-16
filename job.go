package ots

type ErrJobAlreadyStarted error

// Job represents a piece of work to be done
type Job interface {
	// GetID gets the ID of the job
	GetID() string
	// GetStatus gets the status of the job
	GetStatus() string
	// Do does the piece of work in an execution environment
	Do(*Environment) error
}

type JobService interface {
	// Start is called by an agent when it starts a job. ErrJobAlreadyStarted
	// should be returned if another agent has already started it.
	Start(id string, opts JobStartOptions) (Job, error)
	// Finish is called by an agent when it finishes a job.
	Finish(id string, opts JobFinishOptions) (Job, error)
}

type JobStartOptions struct {
	AgentID string
}

type JobFinishOptions struct {
	Errored bool
}
