package logs

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
)

type (
	// PhaseWriter writes logs on behalf of a run phase.
	PhaseWriter struct {
		ctx     context.Context    // permits canceling mid-flow
		started bool               // has first chunk been sent?
		runID   resource.ID        // ID of run to write logs on behalf of.
		phase   internal.PhaseType // run phase
		offset  int                // current position in stream

		PutChunkService // for uploading logs to server
	}

	PhaseWriterOptions struct {
		RunID  resource.ID
		Phase  internal.PhaseType
		Writer PutChunkService
	}
)

// NewPhaseWriter returns a new writer for writing logs on behalf of a run.

func NewPhaseWriter(ctx context.Context, opts PhaseWriterOptions) *PhaseWriter {
	return &PhaseWriter{
		ctx:             ctx,
		runID:           opts.RunID,
		phase:           opts.Phase,
		PutChunkService: opts.Writer,
	}
}

// Write uploads a chunk of logs to the server.
func (w *PhaseWriter) Write(p []byte) (int, error) {
	// TODO: io.Writer's should not retain p but do we need to copy it? Does
	// this code 'retain' p? Does the cache or the database 'retain' p?
	data := make([]byte, len(p))
	copy(data, p)

	if !w.started {
		w.started = true
		data = append([]byte{STX}, data...)
	}

	chunk := PutChunkOptions{
		RunID:  w.runID,
		Phase:  w.phase,
		Data:   data,
		Offset: w.offset,
	}
	w.offset += len(data)

	if err := w.PutChunk(w.ctx, chunk); err != nil {
		return 0, fmt.Errorf("writing log stream: %w", err)
	}

	return len(p), nil
}

// Close must be called to complete writing job logs
func (w *PhaseWriter) Close() error {
	opts := PutChunkOptions{
		RunID:  w.runID,
		Phase:  w.phase,
		Offset: w.offset,
	}
	if w.started {
		opts.Data = []byte{ETX}
	} else {
		opts.Data = []byte{STX, ETX}
	}

	if err := w.PutChunk(w.ctx, opts); err != nil {
		return err
	}
	return nil
}
