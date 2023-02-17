package logs

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTail(t *testing.T) {
	ctx := context.Background()

	t.Run("receive published chunk", func(t *testing.T) {
		app := fakeTailApp(Chunk{})

		stream, err := app.tail(ctx, GetChunkOptions{
			RunID: "run-123",
			Phase: otf.PlanPhase,
		})
		require.NoError(t, err)

		want := Chunk{
			RunID:  "run-123",
			Phase:  otf.PlanPhase,
			Data:   []byte("\x02hello world\x03"),
			Offset: 6,
		}
		app.Publish(otf.Event{
			Payload: PersistedChunk{
				Chunk: want,
			},
		})
		require.Equal(t, want, <-stream)
	})

	t.Run("receive existing chunk", func(t *testing.T) {
		want := Chunk{
			RunID: "run-123",
			Phase: otf.PlanPhase,
			Data:  []byte("\x02hello"),
		}
		app := fakeTailApp(want)

		stream, err := app.tail(ctx, GetChunkOptions{
			RunID: "run-123",
			Phase: otf.PlanPhase,
		})
		require.NoError(t, err)

		require.Equal(t, want, <-stream)
	})

	t.Run("receive existing chunk and overlapping published chunk", func(t *testing.T) {
		want := Chunk{
			RunID: "run-123",
			Phase: otf.PlanPhase,
			Data:  []byte("\x02hello"),
		}
		app := fakeTailApp(want)

		stream, err := app.tail(ctx, GetChunkOptions{
			RunID: "run-123",
			Phase: otf.PlanPhase,
		})
		require.NoError(t, err)

		// receive existing chunk
		require.Equal(t, want, <-stream)

		// receive overlapping chunk
		app.Publish(otf.Event{
			Payload: PersistedChunk{
				Chunk: Chunk{
					RunID:  "run-123",
					Phase:  otf.PlanPhase,
					Data:   []byte("lo world\x03"),
					Offset: 4,
				},
			},
		})

		// receive existing chunk
		want = Chunk{
			RunID:  "run-123",
			Phase:  otf.PlanPhase,
			Data:   []byte("world\x03"),
			Offset: 6,
		}
		assert.Equal(t, want, <-stream)
	})

	t.Run("ignore duplicate chunk", func(t *testing.T) {
		want := Chunk{
			RunID: "run-123",
			Phase: otf.PlanPhase,
			Data:  []byte("\x02hello"),
		}
		app := fakeTailApp(want)

		stream, err := app.tail(ctx, GetChunkOptions{
			RunID: "run-123",
			Phase: otf.PlanPhase,
		})
		require.NoError(t, err)

		// receive existing chunk
		require.Equal(t, want, <-stream)

		// publish duplicate chunk
		app.Publish(otf.Event{
			Payload: PersistedChunk{
				Chunk: Chunk{
					RunID: "run-123",
					Phase: otf.PlanPhase,
					Data:  []byte("\x02hello"),
				},
			},
		})

		// publish non-duplicate chunk
		want = Chunk{
			RunID: "run-123",
			Phase: otf.PlanPhase,
			Data:  []byte(" world\x03"),
		}
		app.Publish(otf.Event{
			Payload: PersistedChunk{
				Chunk: want,
			},
		})
		// dup event is skipped and non-dup is received
		assert.Equal(t, want, <-stream)
	})

	t.Run("ignore chunk for other run", func(t *testing.T) {
		app := fakeTailApp(Chunk{})

		stream, err := app.tail(ctx, GetChunkOptions{
			RunID: "run-123",
			Phase: otf.PlanPhase,
		})
		require.NoError(t, err)

		// publish chunk for other run
		app.Publish(otf.Event{
			Payload: PersistedChunk{
				Chunk: Chunk{
					RunID: "run-456",
					Phase: otf.PlanPhase,
					Data:  []byte("workers of the world, unite"),
				},
			},
		})

		// publish chunk for tailed run
		want := Chunk{
			RunID: "run-123",
			Phase: otf.PlanPhase,
			Data:  []byte("\x02hello"),
		}
		app.Publish(otf.Event{
			Payload: PersistedChunk{
				Chunk: want,
			},
		})
		// chunk for other run is skipped but chunk for this run is received
		assert.Equal(t, want, <-stream)
	})
}

func fakeTailApp(existing Chunk) *Application {
	return &Application{
		proxy:         &fakeTailProxy{chunk: existing},
		PubSubService: newFakePubSubTailService(),
		Logger:        logr.Discard(),
		Authorizer:    &testutil.AllowAllAuthorizer{&otf.Superuser{}},
	}
}

type fakeTailProxy struct {
	// fake chunk to return
	chunk Chunk

	db
}

func (f *fakeTailProxy) get(ctx context.Context, opts GetChunkOptions) (Chunk, error) {
	return f.chunk, nil
}

type fakePubSubTailService struct {
	stream chan otf.Event
}

func newFakePubSubTailService() *fakePubSubTailService {
	return &fakePubSubTailService{stream: make(chan otf.Event)}
}

func (f *fakePubSubTailService) Subscribe(context.Context, string) (<-chan otf.Event, error) {
	return f.stream, nil
}

func (f *fakePubSubTailService) Publish(event otf.Event) {
	f.stream <- event
}
