package ots

import "context"

const (
	JobPending   JobStatus = "pending"
	JobStarted   JobStatus = "started"
	JobCompleted JobStatus = "completed"
	JobErrored   JobStatus = "errored"
)

type ErrJobAlreadyStarted error

type JobStatus string

// Job represents a piece of work to be done
type Job interface {
	// Do does the piece of work
	Do(ctx context.Context) error
}

type JobService interface {
	// Start is called by an agent when it starts the job. ErrJobAlreadyStarted
	// should be returned if another agent has already started it.
	Start(id string) error
}
