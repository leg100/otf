package logs

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB(t *testing.T) {
	db := &pgdb{sql.NewTestDB(t)}
	ctx := context.Background()

	org := organization.CreateTestOrganization(t, db)
	ws := workspace.CreateTestWorkspace(t, db, org.Name)
	cv := configversion.CreateTestConfigurationVersion(t, db, ws, configversion.ConfigurationVersionCreateOptions{})

	t.Run("upload chunk", func(t *testing.T) {
		run := run.CreateTestRun(t, db, ws, cv, run.RunCreateOptions{})

		got, err := db.put(ctx, otf.Chunk{
			RunID: run.ID,
			Phase: otf.PlanPhase,
			Data:  []byte("\x02hello world\x03"),
		})
		require.NoError(t, err)

		want := otf.PersistedChunk{
			ChunkID: got.ChunkID,
			Chunk: otf.Chunk{
				RunID: run.ID,
				Phase: otf.PlanPhase,
				Data:  []byte("\x02hello world\x03"),
			},
		}
		assert.Equal(t, want, got)
		assert.Positive(t, got.ChunkID)
	})

	t.Run("reject empty chunk", func(t *testing.T) {
		run := run.CreateTestRun(t, db, ws, cv, run.RunCreateOptions{})

		_, err := db.put(ctx, otf.Chunk{
			RunID: run.ID,
			Phase: otf.PlanPhase,
		})
		assert.Error(t, err)
	})

	t.Run("get chunk", func(t *testing.T) {
		run := run.CreateTestRun(t, db, ws, cv, run.RunCreateOptions{})

		_, err := db.put(ctx, otf.Chunk{
			RunID: run.ID,
			Phase: otf.PlanPhase,
			Data:  []byte("\x02hello"),
		})
		require.NoError(t, err)

		_, err = db.put(ctx, otf.Chunk{
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
				got, err := db.get(ctx, tt.opts)
				require.NoError(t, err)

				assert.Equal(t, tt.want, got)
			})
		}
	})

	t.Run("get chunk by id", func(t *testing.T) {
		run := run.CreateTestRun(t, db, ws, cv, run.RunCreateOptions{})

		want, err := db.put(ctx, otf.Chunk{
			RunID:  run.ID,
			Phase:  otf.PlanPhase,
			Data:   []byte("\x02hello world\x03"),
			Offset: 0,
		})
		require.NoError(t, err)

		got, err := db.getByID(ctx, want.ChunkID)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})
}
