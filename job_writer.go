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
	// is this the first chunk to be written?
	var firstChunk bool

	if !w.started {
		firstChunk = true
	}

	if err := w.PutChunk(context.Background(), w.ID, p, PutChunkOptions{Start: firstChunk}); err != nil {
		w.Error(err, "unable to write logs")
		return 0, err
	}

	if firstChunk {
		w.started = true
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
