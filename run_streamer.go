package otf

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"
)

type RunStreamer struct {
	run *Run

	planLogsStore, applyLogsStore ChunkStore

	interval time.Duration

	r io.Reader
	w io.WriteCloser
}

func NewRunStreamer(run *Run, planLogsStore, applyLogsStore ChunkStore, interval time.Duration) *RunStreamer {
	streamer := &RunStreamer{
		run:            run,
		planLogsStore:  planLogsStore,
		applyLogsStore: applyLogsStore,
		interval:       interval,
	}

	streamer.r, streamer.w = io.Pipe()

	return streamer
}

// Stream streams logs from a Run's plan, and then, if applicable, from its
// apply.
func (s *RunStreamer) Stream(ctx context.Context) {
	defer s.w.Close()

	// stream plan logs, ignore EOF because we may want to stream apply logs too
	err := s.stream(ctx, s.planLogsStore, s.run.Plan.ID, s.w)
	if err != nil && err != io.EOF {
		s.w.Write([]byte(fmt.Sprintf("stream error: %s", err.Error())))
		return
	}

	if !s.run.Plan.HasChanges() || s.run.IsSpeculative() {
		return
	}

	s.w.Write([]byte("\n"))

	err = s.stream(ctx, s.applyLogsStore, s.run.Apply.ID, s.w)
	if err != nil && err != io.EOF {
		s.w.Write([]byte(fmt.Sprintf("stream error: %s", err.Error())))
		return
	}
}

// Read reads from the read end of the streaming pipe.
func (s *RunStreamer) Read(p []byte) (int, error) {
	return s.r.Read(p)
}

// stream streams chunks of data, from a store at intermittent intervals to
// prevent thrashing the store.
func (s *RunStreamer) stream(ctx context.Context, store ChunkStore, id string, w io.Writer) error {
	offset := 0

	for {
		chunk, err := store.GetChunk(ctx, id, GetChunkOptions{Offset: offset})
		if err != nil {
			return fmt.Errorf("retrieving chunk: %w", err)
		}

		offset += len(chunk)

		if err := writeChunk(chunk, w); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(s.interval):
			// Pause to prevent thrashing store
		}
	}
}

// writeChunk writes a chunk to the writer, stripping any ascii markers
// beforehand. io.EOF is returned when the ETX marker is encountered.
func writeChunk(chunk []byte, w io.Writer) error {
	var eof error

	if len(chunk) == 0 {
		return nil
	}

	if bytes.HasPrefix(chunk, []byte{ChunkStartMarker}) {
		// Strip STX byte from chunk
		chunk = chunk[1:]
	}

	if bytes.HasSuffix(chunk, []byte{ChunkEndMarker}) {
		// Strip ETX byte from chunk
		chunk = chunk[:len(chunk)-1]
		eof = io.EOF
	}

	if _, err := w.Write(chunk); err != nil {
		return fmt.Errorf("writing chunk: %w", err)
	}

	return eof
}
