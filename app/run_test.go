package app

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTail(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		opts otf.GetChunkOptions
		// existing chunk
		existing otf.Chunk
		// incoming event
		event otf.Event
		// expected stream of logs
		want [][]byte
	}{
		{
			name: "receive event",
			opts: otf.GetChunkOptions{
				RunID: "run-123",
				Phase: otf.PlanPhase,
			},
			event: otf.Event{
				Payload: otf.PersistedChunk{
					Chunk: otf.Chunk{
						RunID:  "run-123",
						Phase:  otf.PlanPhase,
						Data:   []byte("\x02hello world\x03"),
						Offset: 6,
					},
				},
			},
			want: [][]byte{
				[]byte("\x02hello world\x03"),
			},
		},
		{
			name: "retrieve from db and receive event",
			opts: otf.GetChunkOptions{
				RunID: "run-123",
				Phase: otf.PlanPhase,
			},
			existing: otf.Chunk{
				RunID: "run-123",
				Phase: otf.PlanPhase,
				Data:  []byte("\x02hello"),
			},
			event: otf.Event{
				Payload: otf.PersistedChunk{
					Chunk: otf.Chunk{
						RunID:  "run-123",
						Phase:  otf.PlanPhase,
						Data:   []byte(" world\x03"),
						Offset: 6,
					},
				},
			},
			want: [][]byte{
				[]byte("\x02hello"),
				[]byte(" world\x03"),
			},
		},
		{
			name: "handle partially overlapping event",
			existing: otf.Chunk{
				RunID: "run-123",
				Phase: otf.PlanPhase,
				Data:  []byte("\x02hello"),
			},
			event: otf.Event{
				Payload: &otf.PersistedChunk{
					Chunk: otf.Chunk{
						RunID:  "run-123",
						Phase:  otf.PlanPhase,
						Data:   []byte("lo world\x03"),
						Offset: 4,
					},
				},
			},
			want: [][]byte{
				[]byte("\x02hello"),
				[]byte(" world\x03"),
			},
		},
		{
			name: "ignore entirely overlapping event",
			existing: otf.Chunk{
				RunID: "run-123",
				Phase: otf.PlanPhase,
				Data:  []byte("\x02hello world\x03"),
			},
			event: otf.Event{
				Payload: &otf.PersistedChunk{
					Chunk: otf.Chunk{
						RunID: "run-123",
						Phase: otf.PlanPhase,
						Data:  []byte("\x02hello world\x03"),
					},
				},
			},
			want: [][]byte{
				[]byte("\x02hello world\x03"),
			},
		},
		{
			name: "ignore event from other run",
			event: otf.Event{
				Payload: otf.PersistedChunk{
					Chunk: otf.Chunk{
						RunID:  "run-other",
						Phase:  otf.PlanPhase,
						Data:   []byte("lo world\x03"),
						Offset: 4,
					},
				},
			},
			want: [][]byte{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &Application{
				proxy:         &fakeTailProxy{chunk: tt.existing},
				Mapper:        &fakeTailMapper{},
				PubSubService: &fakePubSubTailService{event: tt.event},
				Logger:        logr.Discard(),
			}
			stream, err := app.Tail(ctx, tt.opts)
			require.NoError(t, err)

			received := 0
			for got := range stream {
				assert.Equal(t, tt.want[received], got.Data)
				received++
			}
		})
	}
}

type fakeTailMapper struct {
	Mapper
}

func (f *fakeTailMapper) CanAccessRun(context.Context, string) bool { return true }

type fakeTailProxy struct {
	// fake chunk to return
	chunk otf.Chunk
	otf.ChunkStore
}

func (f *fakeTailProxy) GetChunk(ctx context.Context, opts otf.GetChunkOptions) (otf.Chunk, error) {
	return f.chunk, nil
}

type fakePubSubTailService struct {
	event otf.Event
	otf.PubSubService
}

func (f *fakePubSubTailService) Subscribe(context.Context) <-chan otf.Event {
	ch := make(chan otf.Event)
	go func() {
		ch <- f.event
		close(ch)
	}()
	return ch
}
