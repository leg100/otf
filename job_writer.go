package otf

import (
	"context"

	"github.com/go-logr/logr"
)

// JobWriter writes logs on behalf of a job.
type JobWriter struct {
	// ID of Job to write logs on behalf of.
	ID string

	// LogService for uploading logs to server
	LogService

	// started is used internally by the writer to determine whether the first
	// write has been prefixed with the start marker (STX).
	started bool

	logr.Logger
}

// Write uploads a chunk of logs to the server.
func (w *JobWriter) Write(p []byte) (int, error) {
	chunk := Chunk{Data: p}

	if !w.started {
		w.started = true
		chunk.Start = true
	}

	if err := w.PutChunk(context.Background(), w.ID, chunk); err != nil {
		w.Error(err, "unable to write logs")
		w.started = false
		return 0, err
	}

	return len(p), nil
}

// Close must be called to complete writing job logs
func (w *JobWriter) Close() error {
	chunk := Chunk{End: true}

	if !w.started {
		chunk.Start = true
	}

	if err := w.PutChunk(context.Background(), w.ID, chunk); err != nil {
		w.Error(err, "unable to close logs")

		return err
	}
	return nil
}
