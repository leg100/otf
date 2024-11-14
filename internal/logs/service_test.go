package logs

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTail(t *testing.T) {
	ctx := context.Background()

	t.Run("receive chunk event", func(t *testing.T) {
		sub := make(chan pubsub.Event[Chunk])
		app := &Service{
			chunkproxy: &fakeTailProxy{},
			broker:     &fakeSubService{stream: sub},
			Logger:     logr.Discard(),
			Authorizer: &fakeAuthorizer{},
		}

		stream, err := app.Tail(ctx, GetChunkOptions{
			RunID: testutils.ParseID(t, "run-123"),
			Phase: internal.PlanPhase,
		})
		require.NoError(t, err)

		want := Chunk{
			RunID:  testutils.ParseID(t, "run-123"),
			Phase:  internal.PlanPhase,
			Data:   []byte("\x02hello world\x03"),
			Offset: 6,
		}
		sub <- pubsub.Event[Chunk]{Payload: want}
		require.Equal(t, want, <-stream)
	})

	t.Run("receive existing chunk", func(t *testing.T) {
		want := Chunk{
			RunID: testutils.ParseID(t, "run-123"),
			Phase: internal.PlanPhase,
			Data:  []byte("\x02hello"),
		}
		svc := &Service{
			chunkproxy: &fakeTailProxy{chunk: want},
			broker: &fakeSubService{
				stream: make(chan pubsub.Event[Chunk]),
			},
			Logger:     logr.Discard(),
			Authorizer: &fakeAuthorizer{},
		}
		stream, err := svc.Tail(ctx, GetChunkOptions{
			RunID: testutils.ParseID(t, "run-123"),
			Phase: internal.PlanPhase,
		})
		require.NoError(t, err)
		require.Equal(t, want, <-stream)
	})

	t.Run("receive existing chunk and overlapping published chunk", func(t *testing.T) {
		// send first chunk
		want := Chunk{
			RunID: testutils.ParseID(t, "run-123"),
			Phase: internal.PlanPhase,
			Data:  []byte("\x02hello"),
		}
		sub := make(chan pubsub.Event[Chunk])
		svc := &Service{
			chunkproxy: &fakeTailProxy{chunk: want},
			broker:     &fakeSubService{stream: sub},
			Logger:     logr.Discard(),
			Authorizer: &fakeAuthorizer{},
		}

		stream, err := svc.Tail(ctx, GetChunkOptions{
			RunID: testutils.ParseID(t, "run-123"),
			Phase: internal.PlanPhase,
		})
		require.NoError(t, err)

		// receive first chunk
		require.Equal(t, want, <-stream)

		// send second, overlapping, chunk
		sub <- pubsub.Event[Chunk]{
			Payload: Chunk{
				RunID:  testutils.ParseID(t, "run-123"),
				Phase:  internal.PlanPhase,
				Data:   []byte("lo world\x03"),
				Offset: 4,
			},
		}

		// receive non-overlapping part of second chunk.
		want = Chunk{
			RunID:  testutils.ParseID(t, "run-123"),
			Phase:  internal.PlanPhase,
			Data:   []byte(" world\x03"),
			Offset: 6,
		}
		assert.Equal(t, want, <-stream)
	})

	t.Run("ignore duplicate chunk", func(t *testing.T) {
		want := Chunk{
			RunID: testutils.ParseID(t, "run-123"),
			Phase: internal.PlanPhase,
			Data:  []byte("\x02hello"),
		}
		sub := make(chan pubsub.Event[Chunk])
		svc := &Service{
			chunkproxy: &fakeTailProxy{chunk: want},
			broker:     &fakeSubService{stream: sub},
			Logger:     logr.Discard(),
			Authorizer: &fakeAuthorizer{},
		}

		stream, err := svc.Tail(ctx, GetChunkOptions{
			RunID: testutils.ParseID(t, "run-123"),
			Phase: internal.PlanPhase,
		})
		require.NoError(t, err)

		// receive existing chunk
		require.Equal(t, want, <-stream)

		// publish duplicate chunk
		sub <- pubsub.Event[Chunk]{
			Payload: Chunk{
				RunID: testutils.ParseID(t, "run-123"),
				Phase: internal.PlanPhase,
				Data:  []byte("\x02hello"),
			},
		}

		// publish non-duplicate chunk
		want = Chunk{
			RunID:  testutils.ParseID(t, "run-123"),
			Phase:  internal.PlanPhase,
			Data:   []byte(" world\x03"),
			Offset: 6,
		}
		sub <- pubsub.Event[Chunk]{Payload: want}
		// dup event is skipped and non-dup is received
		assert.Equal(t, want, <-stream)
	})

	t.Run("ignore chunk for other run", func(t *testing.T) {
		sub := make(chan pubsub.Event[Chunk])
		svc := &Service{
			chunkproxy: &fakeTailProxy{},
			broker:     &fakeSubService{stream: sub},
			Logger:     logr.Discard(),
			Authorizer: &fakeAuthorizer{},
		}

		stream, err := svc.Tail(ctx, GetChunkOptions{
			RunID: testutils.ParseID(t, "run-123"),
			Phase: internal.PlanPhase,
		})
		require.NoError(t, err)

		// publish chunk for other run
		sub <- pubsub.Event[Chunk]{
			Payload: Chunk{
				RunID: testutils.ParseID(t, "run-456"),
				Phase: internal.PlanPhase,
				Data:  []byte("workers of the world, unite"),
			},
		}

		// publish chunk for tailed run
		want := Chunk{
			RunID: testutils.ParseID(t, "run-123"),
			Phase: internal.PlanPhase,
			Data:  []byte("\x02hello"),
		}
		sub <- pubsub.Event[Chunk]{Payload: want}
		// chunk for other run is skipped but chunk for this run is received
		assert.Equal(t, want, <-stream)
	})
}
