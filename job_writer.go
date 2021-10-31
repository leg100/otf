package otf

import (
	"context"

	"github.com/go-logr/logr"
)

// JobWriter writes logs on behalf of a job.
type JobWriter struct {
	// ID of Job to write logs on behalf of.
	ID string

	// JobService for uploading logs to server
	JobService

	// started is used internally by the writer to determine whether the first
	// write has been prefixed with the start marker (STX).
	started bool

	logr.Logger
}

// Write uploads a chunk of logs to the server.
func (w *JobWriter) Write(p []byte) (int, error) {
	opts := PutChunkOptions{}

	// First chunk is prefixed with STX
	if !w.started {
		w.started = true
		opts.Start = true
	}

	if err := w.PutChunk(context.Background(), w.ID, p, opts); err != nil {
		w.Error(err, "unable to write logs")
		w.started = false
		return 0, err
	}

	return len(p), nil
}

// Close must be called to complete writing job logs
func (w *JobWriter) Close() error {
	opts := PutChunkOptions{End: true}

	if err := w.PutChunk(context.Background(), w.ID, nil, opts); err != nil {
		w.Error(err, "unable to close logs")

		return err
	}
	return nil
}
