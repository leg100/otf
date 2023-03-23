package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogs(t *testing.T) {
	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &otf.Superuser{})

	t.Run("upload chunk", func(t *testing.T) {
		svc := setup(t, "")
		run := svc.createRun(t, ctx, nil, nil)

		err := svc.PutChunk(ctx, otf.Chunk{
			RunID: run.ID,
			Phase: otf.PlanPhase,
			Data:  []byte("\x02hello world\x03"),
		})
		require.NoError(t, err)
	})

	t.Run("reject empty chunk", func(t *testing.T) {
		svc := setup(t, "")
		run := svc.createRun(t, ctx, nil, nil)

		err := svc.PutChunk(ctx, otf.Chunk{
			RunID: run.ID,
			Phase: otf.PlanPhase,
		})
		assert.Error(t, err)
	})

	t.Run("get chunk", func(t *testing.T) {
		svc := setup(t, "")
		run := svc.createRun(t, ctx, nil, nil)

		err := svc.PutChunk(ctx, otf.Chunk{
			RunID: run.ID,
			Phase: otf.PlanPhase,
			Data:  []byte("\x02hello"),
		})
		require.NoError(t, err)

		err = svc.PutChunk(ctx, otf.Chunk{
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
