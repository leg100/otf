package logs

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

// JobWriter writes logs on behalf of a run phase.
//
// TODO: rename to LogWriter or PhaseWriter
type JobWriter struct {
	ID    string        // ID of run to write logs on behalf of.
	Phase otf.PhaseType // run phase

	ChunkService // for uploading logs to server
	logr.Logger

	// started is used internally by the writer to determine whether the first
	// write has been prefixed with the start marker (STX).
	started bool
	offset  int             // current position in stream
	ctx     context.Context // permits canceling mid-flow
}

func NewJobWriter(ctx context.Context, app ChunkService, logger logr.Logger, run otf.Run) *JobWriter {
	return &JobWriter{
		ID:           run.ID(),
		Phase:        run.Phase(),
		ChunkService: app,
		Logger:       logger,
		ctx:          ctx,
	}
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
	w.offset = chunk.NextOffset()

	if err := w.PutChunk(w.ctx, chunk); err != nil {
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

	if err := w.PutChunk(w.ctx, chunk); err != nil {
		w.Error(err, "closing log stream")
		return err
	}
	return nil
}
