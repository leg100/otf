package integration

import (
	"context"
	"testing"

	runpkg "github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogs(t *testing.T) {
	integrationTest(t)

	t.Run("upload chunk", func(t *testing.T) {
		svc, _, ctx := setup(t)
		run := svc.createRun(t, ctx, nil, nil, nil)

		err := svc.Runs.PutChunk(ctx, runpkg.PutChunkOptions{
			RunID: run.ID,
			Phase: runpkg.PlanPhase,
			Data:  []byte("\x02hello world\x03"),
		})
		require.NoError(t, err)
	})

	t.Run("reject empty chunk", func(t *testing.T) {
		svc, _, ctx := setup(t)
		run := svc.createRun(t, ctx, nil, nil, nil)

		err := svc.Runs.PutChunk(ctx, runpkg.PutChunkOptions{
			RunID: run.ID,
			Phase: runpkg.PlanPhase,
		})
		assert.Error(t, err)
	})

	t.Run("get chunk", func(t *testing.T) {
		svc, _, ctx := setup(t)
		run := svc.createRun(t, ctx, nil, nil, nil)

		err := svc.Runs.PutChunk(ctx, runpkg.PutChunkOptions{
			RunID: run.ID,
			Phase: runpkg.PlanPhase,
			Data:  []byte("\x02hello world\x03"),
		})
		require.NoError(t, err)

		tests := []struct {
			name string
			opts runpkg.GetChunkOptions
			want runpkg.Chunk
		}{
			{
				name: "entire chunk",
				opts: runpkg.GetChunkOptions{
					RunID: run.ID,
					Phase: runpkg.PlanPhase,
				},
				want: runpkg.Chunk{
					RunID:  run.ID,
					Phase:  runpkg.PlanPhase,
					Data:   []byte("\x02hello world\x03"),
					Offset: 0,
				},
			},
			{
				name: "first chunk",
				opts: runpkg.GetChunkOptions{
					RunID: run.ID,
					Phase: runpkg.PlanPhase,
					Limit: 4,
				},
				want: runpkg.Chunk{
					RunID:  run.ID,
					Phase:  runpkg.PlanPhase,
					Data:   []byte("\x02hel"),
					Offset: 0,
				},
			},
			{
				name: "intermediate chunk",
				opts: runpkg.GetChunkOptions{
					RunID:  run.ID,
					Phase:  runpkg.PlanPhase,
					Offset: 4,
					Limit:  3,
				},
				want: runpkg.Chunk{
					RunID:  run.ID,
					Phase:  runpkg.PlanPhase,
					Data:   []byte("lo "),
					Offset: 4,
				},
			},
			{
				name: "last chunk",
				opts: runpkg.GetChunkOptions{
					RunID:  run.ID,
					Phase:  runpkg.PlanPhase,
					Offset: 7,
				},
				want: runpkg.Chunk{
					RunID:  run.ID,
					Phase:  runpkg.PlanPhase,
					Data:   []byte("world\x03"),
					Offset: 7,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := svc.Runs.GetChunk(ctx, tt.opts)
				require.NoError(t, err)

				assert.Equal(t, tt.want, got)
			})
		}
	})
}

// TestClusterLogs tests the relaying of logs across a cluster of otfd nodes.
func TestClusterLogs(t *testing.T) {
	integrationTest(t)

	// simulate a cluster of two otfd nodes, and don't start runs
	db := withDatabase(sql.NewTestDB(t))
	local, _, ctx := setup(t, db, disableScheduler())
	remote, _, _ := setup(t, db, disableScheduler())

	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(func() { cancel() })

	// create run on local node
	run := local.createRun(t, ctx, nil, nil, nil)

	// follow run's plan logs on remote node
	sub, err := remote.Runs.Tail(ctx, runpkg.TailOptions{
		RunID: run.ID,
		Phase: runpkg.PlanPhase,
	})
	require.NoError(t, err)

	// upload first chunk to local node
	err = local.Runs.PutChunk(ctx, runpkg.PutChunkOptions{
		RunID: run.ID,
		Phase: runpkg.PlanPhase,
		Data:  []byte("\x02hello"),
	})
	require.NoError(t, err)

	// upload second and last chunk to local node
	err = local.Runs.PutChunk(ctx, runpkg.PutChunkOptions{
		RunID:  run.ID,
		Phase:  runpkg.PlanPhase,
		Data:   []byte(" world\x03"),
		Offset: 6,
	})
	require.NoError(t, err)

	got := <-sub
	assert.Equal(t, run.ID, got.RunID)
	assert.Equal(t, runpkg.PlanPhase, got.Phase)
	assert.Equal(t, []byte("\x02hello"), []byte(got.Data))

	got = <-sub
	assert.Equal(t, run.ID, got.RunID)
	assert.Equal(t, runpkg.PlanPhase, got.Phase)
	assert.Equal(t, []byte(" world\x03"), []byte(got.Data))
	assert.Equal(t, 6, got.Offset)
}
