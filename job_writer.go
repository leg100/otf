package otf

import (
	"context"

	"github.com/go-logr/logr"
)

// JobWriter writes logs on behalf of a run phase.
//
// TODO: rename to LogWriter or PhaseWriter
type JobWriter struct {
	// ID of run to write logs on behalf of.
	ID string

	// run phase
	Phase PhaseType

	// LogService for uploading logs to server
	LogService

	// started is used internally by the writer to determine whether the first
	// write has been prefixed with the start marker (STX).
	started bool

	logr.Logger
}

// Write uploads a chunk of logs to the server.
func (w *JobWriter) Write(p []byte) (int, error) {
	data := make([]byte, len(p))
	copy(data, p)
	chunk := Chunk{Data: data}

	if !w.started {
		w.started = true
		chunk.Start = true
	}

	if err := w.PutChunk(context.Background(), w.ID, w.Phase, chunk); err != nil {
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

	if err := w.PutChunk(context.Background(), w.ID, w.Phase, chunk); err != nil {
		w.Error(err, "unable to close logs")

		return err
	}
	return nil
}
