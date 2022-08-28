package sql

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLog_PutChunk(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws)

	t.Run("first chunk", func(t *testing.T) {
		run := createTestRun(t, db, ws, cv)

		err := db.PutChunk(context.Background(), run.ID(), otf.PlanPhase, otf.Chunk{Data: []byte("chunk1"), Start: true})
		require.NoError(t, err)

		got, err := db.GetChunk(context.Background(), run.ID(), otf.PlanPhase, otf.GetChunkOptions{})
		require.NoError(t, err)

		assert.Equal(t, otf.Chunk{Start: true, Data: []byte("chunk1")}, got)
	})

	t.Run("middle chunk", func(t *testing.T) {
		run := createTestRun(t, db, ws, cv)

		err := db.PutChunk(context.Background(), run.ID(), otf.PlanPhase, otf.Chunk{Data: []byte("chunk1")})
		require.NoError(t, err)

		got, err := db.GetChunk(context.Background(), run.ID(), otf.PlanPhase, otf.GetChunkOptions{})
		require.NoError(t, err)

		assert.Equal(t, otf.Chunk{Data: []byte("chunk1")}, got)
	})

	t.Run("last chunk with no data", func(t *testing.T) {
		run := createTestRun(t, db, ws, cv)

		err := db.PutChunk(context.Background(), run.ID(), otf.PlanPhase, otf.Chunk{End: true})
		require.NoError(t, err)

		got, err := db.GetChunk(context.Background(), run.ID(), otf.PlanPhase, otf.GetChunkOptions{})
		require.NoError(t, err)

		assert.Equal(t, otf.Chunk{Data: []byte{}, End: true}, got)
	})
}

func TestLog_GetChunk(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws)
	run := createTestRun(t, db, ws, cv)

	err := db.PutChunk(context.Background(), run.ID(), otf.PlanPhase, otf.Chunk{Data: []byte("hello"), Start: true})
	require.NoError(t, err)

	err = db.PutChunk(context.Background(), run.ID(), otf.PlanPhase, otf.Chunk{Data: []byte(" world"), End: true})
	require.NoError(t, err)

	tests := []struct {
		name string
		opts otf.GetChunkOptions
		want otf.Chunk
	}{
		{
			name: "all chunks",
			want: otf.Chunk{
				Data:  []byte("hello world"),
				Start: true,
				End:   true,
			},
		},
		{
			name: "first chunk",
			opts: otf.GetChunkOptions{Limit: 4},
			want: otf.Chunk{
				Data:  []byte("hel"),
				Start: true,
			},
		},
		{
			name: "intermediate chunk",
			opts: otf.GetChunkOptions{Offset: 4, Limit: 3},
			want: otf.Chunk{
				Data: []byte("lo "),
			},
		},
		{
			name: "last chunk",
			opts: otf.GetChunkOptions{Offset: 7},
			want: otf.Chunk{
				Data: []byte("world"),
				End:  true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.GetChunk(context.Background(), run.ID(), otf.PlanPhase, tt.opts)
			require.NoError(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}
