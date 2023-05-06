package logs

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	internal "github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTail(t *testing.T) {
	ctx := context.Background()

	t.Run("receive chunk event", func(t *testing.T) {
		app := fakeService(internal.Chunk{})

		stream, err := app.Tail(ctx, internal.GetChunkOptions{
			RunID: "run-123",
			Phase: internal.PlanPhase,
		})
		require.NoError(t, err)

		want := internal.Chunk{
			RunID:  "run-123",
			Phase:  internal.PlanPhase,
			Data:   []byte("\x02hello world\x03"),
			Offset: 6,
		}
		app.Publish(internal.Event{Payload: want})
		require.Equal(t, want, <-stream)
	})

	t.Run("receive existing chunk", func(t *testing.T) {
		want := internal.Chunk{
			RunID: "run-123",
			Phase: internal.PlanPhase,
			Data:  []byte("\x02hello"),
		}
		svc := fakeService(want)

		stream, err := svc.Tail(ctx, internal.GetChunkOptions{
			RunID: "run-123",
			Phase: internal.PlanPhase,
		})
		require.NoError(t, err)

		require.Equal(t, want, <-stream)
	})

	t.Run("receive existing chunk and overlapping published chunk", func(t *testing.T) {
		// send first chunk
		want := internal.Chunk{
			RunID: "run-123",
			Phase: internal.PlanPhase,
			Data:  []byte("\x02hello"),
		}
		svc := fakeService(want)

		stream, err := svc.Tail(ctx, internal.GetChunkOptions{
			RunID: "run-123",
			Phase: internal.PlanPhase,
		})
		require.NoError(t, err)

		// receive first chunk
		require.Equal(t, want, <-stream)

		// send second, overlapping, chunk
		svc.Publish(internal.Event{
			Payload: internal.Chunk{
				RunID:  "run-123",
				Phase:  internal.PlanPhase,
				Data:   []byte("lo world\x03"),
				Offset: 4,
			},
		})

		// receive non-overlapping part of second chunk.
		want = internal.Chunk{
			RunID:  "run-123",
			Phase:  internal.PlanPhase,
			Data:   []byte(" world\x03"),
			Offset: 6,
		}
		assert.Equal(t, want, <-stream)
	})

	t.Run("ignore duplicate chunk", func(t *testing.T) {
		want := internal.Chunk{
			RunID: "run-123",
			Phase: internal.PlanPhase,
			Data:  []byte("\x02hello"),
		}
		app := fakeService(want)

		stream, err := app.Tail(ctx, internal.GetChunkOptions{
			RunID: "run-123",
			Phase: internal.PlanPhase,
		})
		require.NoError(t, err)

		// receive existing chunk
		require.Equal(t, want, <-stream)

		// publish duplicate chunk
		app.Publish(internal.Event{
			Payload: internal.Chunk{
				RunID: "run-123",
				Phase: internal.PlanPhase,
				Data:  []byte("\x02hello"),
			},
		})

		// publish non-duplicate chunk
		want = internal.Chunk{
			RunID:  "run-123",
			Phase:  internal.PlanPhase,
			Data:   []byte(" world\x03"),
			Offset: 6,
		}
		app.Publish(internal.Event{Payload: want})
		// dup event is skipped and non-dup is received
		assert.Equal(t, want, <-stream)
	})

	t.Run("ignore chunk for other run", func(t *testing.T) {
		app := fakeService(internal.Chunk{})

		stream, err := app.Tail(ctx, internal.GetChunkOptions{
			RunID: "run-123",
			Phase: internal.PlanPhase,
		})
		require.NoError(t, err)

		// publish chunk for other run
		app.Publish(internal.Event{
			Payload: internal.Chunk{
				RunID: "run-456",
				Phase: internal.PlanPhase,
				Data:  []byte("workers of the world, unite"),
			},
		})

		// publish chunk for tailed run
		want := internal.Chunk{
			RunID: "run-123",
			Phase: internal.PlanPhase,
			Data:  []byte("\x02hello"),
		}
		app.Publish(internal.Event{
			Payload: want,
		})
		// chunk for other run is skipped but chunk for this run is received
		assert.Equal(t, want, <-stream)
	})
}

func fakeService(existing internal.Chunk) *service {
	return &service{
		proxy:         &fakeTailProxy{chunk: existing},
		PubSubService: newFakePubSubService(),
		Logger:        logr.Discard(),
		run:           &fakeAuthorizer{},
	}
}
