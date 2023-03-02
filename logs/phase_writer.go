package logs

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

type logWriter interface {
	PutChunk(ctx context.Context, chunk Chunk) error
}

// PhaseWriter writes logs on behalf of a run phase.
type PhaseWriter struct {
	logr.Logger

	// started is used internally by the writer to determine whether the first
	// write has been prefixed with the start marker (STX).
	started   bool
	id        string          // ID of run to write logs on behalf of.
	phase     otf.PhaseType   // run phase
	offset    int             // current position in stream
	ctx       context.Context // permits canceling mid-flow
	logWriter                 // for uploading logs to server
}

// NewPhaseWriter returns a new writer for writing logs on behalf of a run.

func NewPhaseWriter(ctx context.Context, logger logr.Logger, w logWriter, run otf.Run) *PhaseWriter {
	return &PhaseWriter{
		id:        run.ID,
		phase:     run.Phase(),
		logWriter: w,
		Logger:    logger,
		ctx:       ctx,
	}
}

// Write uploads a chunk of logs to the server.
func (w *PhaseWriter) Write(p []byte) (int, error) {
	data := make([]byte, len(p))
	copy(data, p)

	chunk := Chunk{
		RunID:  w.id,
		Phase:  w.phase,
		Data:   data,
		Offset: w.offset,
	}

	if !w.started {
		w.started = true
		chunk = chunk.AddStartMarker()
	}
	w.offset = chunk.NextOffset()

	if err := w.PutChunk(w.ctx, chunk); err != nil {
		w.Error(err, "writing log stream")
		return 0, err
	}

	return len(p), nil
}

// Close must be called to complete writing job logs
func (w *PhaseWriter) Close() error {
	chunk := Chunk{
		RunID:  w.id,
		Phase:  w.phase,
		Offset: w.offset,
	}
	chunk = chunk.AddEndMarker()
	if !w.started {
		chunk = chunk.AddStartMarker()
	}
	w.offset += chunk.NextOffset()

	if err := w.PutChunk(w.ctx, chunk); err != nil {
		w.Error(err, "closing log stream")
		return err
	}
	return nil
}
