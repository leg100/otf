package otf

import "context"

type ErrJobAlreadyStarted error

// Job is either a Run's Plan or Apply.
type Job interface {
	// Do performs the task of work in an execution environment
	Do(*Execution) error

	// JobService provides methods for updating the status of the job.
	JobService

	// GetID gets the ID of the Job
	GetID() string

	// GetStatus gets the status of the Job
	GetStatus() string
}

type JobService interface {
	// Start is called by an agent when it starts a job. ErrJobAlreadyStarted
	// should be returned if another agent has already started it.
	Start(ctx context.Context, id string, opts JobStartOptions) (*Run, error)

	// Finish is called by an agent when it finishes a job.
	Finish(ctx context.Context, id string, opts JobFinishOptions) (*Run, error)

	// ChunkStore handles putting and getting chunks of logs
	ChunkStore
}

type JobStartOptions struct {
	AgentID string
}

type JobFinishOptions struct {
	Errored bool
}
