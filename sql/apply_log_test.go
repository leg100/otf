package sql

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyLog_PutChunk(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws)
	run := createTestRun(t, db, ws, cv)

	logdb := NewApplyLogDB(db)

	err := logdb.PutChunk(context.Background(), run.Apply.ID, []byte("chunk1"), otf.PutChunkOptions{Start: true})
	require.NoError(t, err)
}

func TestApplyLog_GetChunk(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws)
	run := createTestRun(t, db, ws, cv)

	logdb := NewApplyLogDB(db)

	err := logdb.PutChunk(context.Background(), run.Apply.ID, []byte("chunk1"), otf.PutChunkOptions{Start: true})
	require.NoError(t, err)

	err = logdb.PutChunk(context.Background(), run.Apply.ID, []byte("chunk2"), otf.PutChunkOptions{})
	require.NoError(t, err)

	err = logdb.PutChunk(context.Background(), run.Apply.ID, []byte("chunk3"), otf.PutChunkOptions{End: true})
	require.NoError(t, err)

	tests := []struct {
		name string
		opts otf.GetChunkOptions
		want string
	}{
		{
			name: "all chunks",
			opts: otf.GetChunkOptions{},
			want: "\x02chunk1chunk2chunk3\x03",
		},
		{
			name: "first chunk",
			opts: otf.GetChunkOptions{Limit: 9},
			want: "\x02chunk1ch",
		},
		{
			name: "intermediate chunk",
			opts: otf.GetChunkOptions{Offset: 10, Limit: 9},
			want: "nk2chunk3",
		},
		{
			name: "last chunk",
			opts: otf.GetChunkOptions{Offset: 15},
			want: "unk3\x03",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := logdb.GetChunk(context.Background(), run.Apply.ID, tt.opts)
			require.NoError(t, err)

			assert.Equal(t, tt.want, string(got))
		})
	}
}
