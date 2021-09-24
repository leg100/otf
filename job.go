package otf

import (
	"context"
)

type ErrJobAlreadyStarted error

// Job represents a piece of work to be done
type Job interface {
	// GetID gets the ID of the job
	GetID() string
	// GetStatus gets the status of the job
	GetStatus() string
	// Do does the piece of work in an execution environment
	Do(*Executor) error
}

type JobService interface {
	// Start is called by an agent when it starts a job. ErrJobAlreadyStarted
	// should be returned if another agent has already started it.
	Start(id string, opts JobStartOptions) (Job, error)
	// Finish is called by an agent when it finishes a job.
	Finish(id string, opts JobFinishOptions) (Job, error)

	JobLogsUploader
}

type JobLogsUploader interface {
	// UploadLogs uploads a chunk of output from the job.
	UploadLogs(ctx context.Context, id string, logs []byte, opts RunUploadLogsOptions) error
}

type JobStartOptions struct {
	AgentID string
}

type JobFinishOptions struct {
	Errored bool
}

// RunUploadLogsOptions represents the options for uploading logs for a run.
type RunUploadLogsOptions struct {
	// End indicates this is the last and final chunk
	End bool `schema:"end"`
}
