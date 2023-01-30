package sql

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLog_PutChunk(t *testing.T) {
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	ws := CreateTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws, otf.ConfigurationVersionCreateOptions{})
	run := createTestRun(t, db, ws, cv)
	ctx := context.Background()

	t.Run("upload chunk", func(t *testing.T) {
		got, err := db.PutChunk(ctx, otf.Chunk{
			RunID: run.ID(),
			Phase: otf.PlanPhase,
			Data:  []byte("\x02hello world\x03"),
		})
		require.NoError(t, err)

		want := otf.PersistedChunk{
			ChunkID: got.ChunkID,
			Chunk: otf.Chunk{
				RunID: run.ID(),
				Phase: otf.PlanPhase,
				Data:  []byte("\x02hello world\x03"),
			},
		}
		assert.Equal(t, want, got)
		assert.Positive(t, got.ChunkID)
	})

	t.Run("reject empty chunk", func(t *testing.T) {
		_, err := db.PutChunk(ctx, otf.Chunk{
			RunID: run.ID(),
			Phase: otf.PlanPhase,
		})
		assert.Error(t, err)
	})
}

func TestLog_GetChunk(t *testing.T) {
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	ws := CreateTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws, otf.ConfigurationVersionCreateOptions{})
	run := createTestRun(t, db, ws, cv)
	ctx := context.Background()

	_, err := db.PutChunk(ctx, otf.Chunk{
		RunID: run.ID(),
		Phase: otf.PlanPhase,
		Data:  []byte("\x02hello"),
	})
	require.NoError(t, err)

	_, err = db.PutChunk(ctx, otf.Chunk{
		RunID: run.ID(),
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
				RunID: run.ID(),
				Phase: otf.PlanPhase,
			},
			want: otf.Chunk{
				RunID:  run.ID(),
				Phase:  otf.PlanPhase,
				Data:   []byte("\x02hello world\x03"),
				Offset: 0,
			},
		},
		{
			name: "first chunk",
			opts: otf.GetChunkOptions{
				RunID: run.ID(),
				Phase: otf.PlanPhase,
				Limit: 4,
			},
			want: otf.Chunk{
				RunID:  run.ID(),
				Phase:  otf.PlanPhase,
				Data:   []byte("\x02hel"),
				Offset: 0,
			},
		},
		{
			name: "intermediate chunk",
			opts: otf.GetChunkOptions{
				RunID:  run.ID(),
				Phase:  otf.PlanPhase,
				Offset: 4,
				Limit:  3,
			},
			want: otf.Chunk{
				RunID:  run.ID(),
				Phase:  otf.PlanPhase,
				Data:   []byte("lo "),
				Offset: 4,
			},
		},
		{
			name: "last chunk",
			opts: otf.GetChunkOptions{
				RunID:  run.ID(),
				Phase:  otf.PlanPhase,
				Offset: 7,
			},
			want: otf.Chunk{
				RunID:  run.ID(),
				Phase:  otf.PlanPhase,
				Data:   []byte("world\x03"),
				Offset: 7,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.GetChunk(ctx, tt.opts)
			require.NoError(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLog_GetChunkByID(t *testing.T) {
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	ws := CreateTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws, otf.ConfigurationVersionCreateOptions{})
	run := createTestRun(t, db, ws, cv)
	ctx := context.Background()

	want, err := db.PutChunk(ctx, otf.Chunk{
		RunID:  run.ID(),
		Phase:  otf.PlanPhase,
		Data:   []byte("\x02hello world\x03"),
		Offset: 0,
	})
	require.NoError(t, err)

	got, err := db.GetChunkByID(ctx, want.ChunkID)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}
