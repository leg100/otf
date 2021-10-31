package otf

import (
	"context"

	"github.com/go-logr/logr"
)

// JobWriter writes logs on behalf of a job.
type JobWriter struct {
	// Job to write logs on behalf of.
	Job

	// started is used internally by the writer to determine whether the first
	// write has been prefixed with the start marker (STX).
	started bool

	// ID of job
	ID string

	logr.Logger
}

// Write uploads a chunk of logs to the server.
func (jw *JobWriter) Write(p []byte) (int, error) {
	opts := PutChunkOptions{}

	// First chunk is prefixed with STX
	if !jw.started {
		jw.started = true
		opts.Start = true
	}

	if err := jw.PutChunk(context.Background(), jw.ID, p, opts); err != nil {
		jw.Error(err, "unable to write logs")
		jw.started = false
		return 0, err
	}

	return len(p), nil
}

// Close must be called to complete writing job logs
func (jw *JobWriter) Close() error {
	opts := PutChunkOptions{End: true}

	if err := jw.PutChunk(context.Background(), jw.ID, nil, opts); err != nil {
		jw.Error(err, "unable to close logs")

		return err
	}
	return nil
}
