package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogs(t *testing.T) {
	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &otf.Superuser{})

	t.Run("upload chunk", func(t *testing.T) {
		svc := setup(t, nil)
		run := svc.createRun(t, ctx, nil, nil)

		err := svc.PutChunk(ctx, otf.PutChunkOptions{
			RunID: run.ID,
			Phase: otf.PlanPhase,
			Data:  []byte("\x02hello world\x03"),
		})
		require.NoError(t, err)
	})

	t.Run("reject empty chunk", func(t *testing.T) {
		svc := setup(t, nil)
		run := svc.createRun(t, ctx, nil, nil)

		err := svc.PutChunk(ctx, otf.PutChunkOptions{
			RunID: run.ID,
			Phase: otf.PlanPhase,
		})
		assert.Error(t, err)
	})

	t.Run("get chunk", func(t *testing.T) {
		svc := setup(t, nil)
		run := svc.createRun(t, ctx, nil, nil)

		err := svc.PutChunk(ctx, otf.PutChunkOptions{
			RunID: run.ID,
			Phase: otf.PlanPhase,
			Data:  []byte("\x02hello"),
		})
		require.NoError(t, err)

		err = svc.PutChunk(ctx, otf.PutChunkOptions{
			RunID: run.ID,
			Phase: otf.PlanPhase,
			Data:  []byte(" world\x03"),
		})
		require.NoError(t, err)

		tests := []struct {
			name string
			opts otf.GetChunkOptions
			want otf.Chunk
		}{
			{
				name: "all chunks",
				opts: otf.GetChunkOptions{
					RunID: run.ID,
					Phase: otf.PlanPhase,
				},
				want: otf.Chunk{
					RunID:  run.ID,
					Phase:  otf.PlanPhase,
					Data:   []byte("\x02hello world\x03"),
					Offset: 0,
				},
			},
			{
				name: "first chunk",
				opts: otf.GetChunkOptions{
					RunID: run.ID,
					Phase: otf.PlanPhase,
					Limit: 4,
				},
				want: otf.Chunk{
					RunID:  run.ID,
					Phase:  otf.PlanPhase,
					Data:   []byte("\x02hel"),
					Offset: 0,
				},
			},
			{
				name: "intermediate chunk",
				opts: otf.GetChunkOptions{
					RunID:  run.ID,
					Phase:  otf.PlanPhase,
					Offset: 4,
					Limit:  3,
				},
				want: otf.Chunk{
					RunID:  run.ID,
					Phase:  otf.PlanPhase,
					Data:   []byte("lo "),
					Offset: 4,
				},
			},
			{
				name: "last chunk",
				opts: otf.GetChunkOptions{
					RunID:  run.ID,
					Phase:  otf.PlanPhase,
					Offset: 7,
				},
				want: otf.Chunk{
					RunID:  run.ID,
					Phase:  otf.PlanPhase,
					Data:   []byte("world\x03"),
					Offset: 7,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := svc.GetChunk(ctx, tt.opts)
				require.NoError(t, err)

				assert.Equal(t, tt.want, got)
			})
		}
	})
}

// TestClusterLogs tests the relaying of logs across a cluster of otfd nodes.
func TestClusterLogs(t *testing.T) {
	t.Parallel()

	// simulate a cluster of two otfd nodes
	db, _ := sql.NewTestDB(t)
	local := setup(t, &config{db: db})
	remote := setup(t, &config{db: db})

	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &otf.Superuser{})
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(func() { cancel() })

	// start broker and caching proxy for each node
	done := make(chan error)
	go func() {
		done <- local.Broker.Start(ctx)
	}()
	go func() {
		done <- local.StartProxy(ctx)
	}()
	go func() {
		done <- remote.Broker.Start(ctx)
	}()
	go func() {
		done <- remote.StartProxy(ctx)
	}()

	// wait 'til brokers are listening
	local.Broker.WaitUntilListening()
	remote.Broker.WaitUntilListening()

	// create run on local node
	run := local.createRun(t, ctx, nil, nil)

	// follow run's plan logs on remote node
	sub, err := remote.Tail(ctx, otf.GetChunkOptions{
		RunID: run.ID,
		Phase: otf.PlanPhase,
	})
	require.NoError(t, err)

	// upload first chunk to local node
	err = local.PutChunk(ctx, otf.PutChunkOptions{
		RunID: run.ID,
		Phase: otf.PlanPhase,
		Data:  []byte("\x02hello"),
	})
	require.NoError(t, err)

	// upload second and last chunk to local node
	err = local.PutChunk(ctx, otf.PutChunkOptions{
		RunID:  run.ID,
		Phase:  otf.PlanPhase,
		Data:   []byte(" world\x03"),
		Offset: 6,
	})
	require.NoError(t, err)

	want1 := otf.Chunk{
		ID:    "1",
		RunID: run.ID,
		Phase: otf.PlanPhase,
		Data:  []byte("\x02hello"),
	}
	require.Equal(t, want1, <-sub)

	want2 := otf.Chunk{
		ID:     "2",
		RunID:  run.ID,
		Phase:  otf.PlanPhase,
		Data:   []byte(" world\x03"),
		Offset: 6,
	}
	require.Equal(t, want2, <-sub)

	cancel()
	assert.NoError(t, <-done)
	assert.NoError(t, <-done)
}
