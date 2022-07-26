package tail

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_New(t *testing.T) {
	srv := NewServer(&fakeChunkService{
		chunk: otf.Chunk{Data: []byte("cat sat on the mat")},
	})

	client, err := srv.Tail(context.Background(), otf.PhaseSpec{RunID: "run-123", Phase: otf.PlanPhase}, 0)
	require.NoError(t, err)

	buf := <-client.buffer
	assert.Equal(t, "cat sat on the mat", string(buf))
}

func TestServer_New_LastChunk(t *testing.T) {
	srv := NewServer(&fakeChunkService{
		chunk: otf.Chunk{
			End: true,
		},
	})

	client, err := srv.Tail(context.Background(), otf.PhaseSpec{RunID: "run-123", Phase: otf.PlanPhase}, 0)
	require.NoError(t, err)

	assert.Nil(t, <-client.buffer)
}

func TestServer_PutChunk(t *testing.T) {
	srv := NewServer(&fakeChunkService{
		chunk: otf.Chunk{
			Data: []byte("cat sat on the mat"),
		},
	})

	spec := otf.PhaseSpec{RunID: "run-123", Phase: otf.PlanPhase}
	client, err := srv.Tail(context.Background(), spec, 0)
	require.NoError(t, err)
	// There should be one client in the db
	assert.Equal(t, 1, len(srv.db))

	assert.Equal(t, "cat sat on the mat", string(<-client.Read()))

	srv.PutChunk(spec, otf.Chunk{Data: []byte(" and died the next day"), End: true})
	assert.Equal(t, " and died the next day", string(<-client.Read()))
	assert.Nil(t, <-client.Read())

	client.Close()
	assert.Equal(t, 0, len(srv.db))
}

type fakeChunkService struct {
	chunk otf.Chunk
	otf.ChunkService
}

func (f *fakeChunkService) GetChunk(context.Context, string, otf.PhaseType, otf.GetChunkOptions) (otf.Chunk, error) {
	return f.chunk, nil
}
