package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/daemon"
	"github.com/leg100/otf/internal/logs"
	"github.com/leg100/otf/internal/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogs(t *testing.T) {
	integrationTest(t)

	t.Run("upload chunk", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		run := svc.createRun(t, ctx, nil, nil)

		err := svc.Logs.PutChunk(ctx, logs.PutChunkOptions{
			RunID: run.ID,
			Phase: internal.PlanPhase,
			Data:  []byte("\x02hello world\x03"),
		})
		require.NoError(t, err)
	})

	t.Run("reject empty chunk", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		run := svc.createRun(t, ctx, nil, nil)

		err := svc.Logs.PutChunk(ctx, logs.PutChunkOptions{
			RunID: run.ID,
			Phase: internal.PlanPhase,
		})
		assert.Error(t, err)
	})

	t.Run("get chunk", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		run := svc.createRun(t, ctx, nil, nil)

		err := svc.Logs.PutChunk(ctx, logs.PutChunkOptions{
			RunID: run.ID,
			Phase: internal.PlanPhase,
			Data:  []byte("\x02hello world\x03"),
		})
		require.NoError(t, err)

		tests := []struct {
			name string
			opts logs.GetChunkOptions
			want logs.Chunk
		}{
			{
				name: "entire chunk",
				opts: logs.GetChunkOptions{
					RunID: run.ID,
					Phase: internal.PlanPhase,
				},
				want: logs.Chunk{
					RunID:  run.ID,
					Phase:  internal.PlanPhase,
					Data:   []byte("\x02hello world\x03"),
					Offset: 0,
				},
			},
			{
				name: "first chunk",
				opts: logs.GetChunkOptions{
					RunID: run.ID,
					Phase: internal.PlanPhase,
					Limit: 4,
				},
				want: logs.Chunk{
					RunID:  run.ID,
					Phase:  internal.PlanPhase,
					Data:   []byte("\x02hel"),
					Offset: 0,
				},
			},
			{
				name: "intermediate chunk",
				opts: logs.GetChunkOptions{
					RunID:  run.ID,
					Phase:  internal.PlanPhase,
					Offset: 4,
					Limit:  3,
				},
				want: logs.Chunk{
					RunID:  run.ID,
					Phase:  internal.PlanPhase,
					Data:   []byte("lo "),
					Offset: 4,
				},
			},
			{
				name: "last chunk",
				opts: logs.GetChunkOptions{
					RunID:  run.ID,
					Phase:  internal.PlanPhase,
					Offset: 7,
				},
				want: logs.Chunk{
					RunID:  run.ID,
					Phase:  internal.PlanPhase,
					Data:   []byte("world\x03"),
					Offset: 7,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := svc.Logs.GetChunk(ctx, tt.opts)
				require.NoError(t, err)

				assert.Equal(t, tt.want, got)
			})
		}
	})
}

// TestClusterLogs tests the relaying of logs across a cluster of otfd nodes.
func TestClusterLogs(t *testing.T) {
	integrationTest(t)

	// simulate a cluster of two otfd nodes
	connstr := sql.NewTestDB(t)
	local, _, ctx := setup(t, &config{Config: daemon.Config{
		Database:         connstr,
		DisableScheduler: true, // don't start run
	}})
	remote, _, _ := setup(t, &config{Config: daemon.Config{
		Database:         connstr,
		DisableScheduler: true, // don't start run
	}})

	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(func() { cancel() })

	// create run on local node
	run := local.createRun(t, ctx, nil, nil)

	// follow run's plan logs on remote node
	sub, err := remote.Logs.Tail(ctx, logs.GetChunkOptions{
		RunID: run.ID,
		Phase: internal.PlanPhase,
	})
	require.NoError(t, err)

	// upload first chunk to local node
	err = local.Logs.PutChunk(ctx, logs.PutChunkOptions{
		RunID: run.ID,
		Phase: internal.PlanPhase,
		Data:  []byte("\x02hello"),
	})
	require.NoError(t, err)

	// upload second and last chunk to local node
	err = local.Logs.PutChunk(ctx, logs.PutChunkOptions{
		RunID:  run.ID,
		Phase:  internal.PlanPhase,
		Data:   []byte(" world\x03"),
		Offset: 6,
	})
	require.NoError(t, err)

	got := <-sub
	assert.Equal(t, run.ID, got.RunID)
	assert.Equal(t, internal.PlanPhase, got.Phase)
	assert.Equal(t, []byte("\x02hello"), got.Data)

	got = <-sub
	assert.Equal(t, run.ID, got.RunID)
	assert.Equal(t, internal.PlanPhase, got.Phase)
	assert.Equal(t, []byte(" world\x03"), got.Data)
	assert.Equal(t, 6, got.Offset)
}
