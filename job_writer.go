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
	// current position in the stream
	offset int

	logr.Logger
}

// Write uploads a chunk of logs to the server.
func (w *JobWriter) Write(p []byte) (int, error) {
	data := make([]byte, len(p))
	copy(data, p)

	chunk := Chunk{
		RunID:  w.ID,
		Phase:  w.Phase,
		Data:   data,
		Offset: w.offset,
	}

	if !w.started {
		w.started = true
		chunk = chunk.AddStartMarker()
	}
	w.offset += chunk.NextOffset()

	if err := w.PutChunk(context.Background(), chunk); err != nil {
		w.Error(err, "writing log stream")
		return 0, err
	}

	return len(p), nil
}

// Close must be called to complete writing job logs
func (w *JobWriter) Close() error {
	chunk := Chunk{
		RunID:  w.ID,
		Phase:  w.Phase,
		Offset: w.offset,
	}
	chunk = chunk.AddEndMarker()
	if !w.started {
		chunk = chunk.AddStartMarker()
	}
	w.offset += chunk.NextOffset()

	if err := w.PutChunk(context.Background(), chunk); err != nil {
		w.Error(err, "closing log stream")
		return err
	}
	return nil
}
