package logs

import (
	"context"
	"fmt"

	"github.com/leg100/otf"
)

type (
	chunkWriter interface {
		PutChunk(ctx context.Context, chunk otf.Chunk) error
	}

	// PhaseWriter writes logs on behalf of a run phase.
	PhaseWriter struct {
		ctx         context.Context // permits canceling mid-flow
		started     bool            // has first chunk been sent?
		id          string          // ID of run to write logs on behalf of.
		phase       otf.PhaseType   // run phase
		offset      int             // current position in stream
		chunkWriter                 // for uploading logs to server
	}

	PhaseWriterOptions struct {
		RunID  string
		Phase  otf.PhaseType
		Writer chunkWriter
	}
)

// NewPhaseWriter returns a new writer for writing logs on behalf of a run.

func NewPhaseWriter(ctx context.Context, opts PhaseWriterOptions) *PhaseWriter {
	return &PhaseWriter{
		ctx:         ctx,
		id:          opts.RunID,
		phase:       opts.Phase,
		chunkWriter: opts.Writer,
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
		data = append([]byte{otf.STX}, data...)
	}

	chunk := otf.Chunk{
		RunID:  w.id,
		Phase:  w.phase,
		Data:   data,
		Offset: w.offset,
	}
	w.offset = chunk.NextOffset()

	if err := w.PutChunk(w.ctx, chunk); err != nil {
		return 0, fmt.Errorf("writing log stream: %w", err)
	}

	return len(p), nil
}

// Close must be called to complete writing job logs
func (w *PhaseWriter) Close() error {
	chunk := otf.Chunk{
		RunID:  w.id,
		Phase:  w.phase,
		Offset: w.offset,
	}
	if w.started {
		chunk.Data = []byte{otf.ETX}
	} else {
		chunk.Data = []byte{otf.STX, otf.ETX}
	}

	err := w.PutChunk(w.ctx, chunk)
	if err != nil {
		return fmt.Errorf("closing log stream: %w", err)
	}
	return nil
}
