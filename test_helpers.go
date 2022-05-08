package otf

import (
	"context"
	"fmt"
)

type testChunkStore struct {
	store map[string]Chunk

	ChunkStore
}

func (s *testChunkStore) GetChunk(ctx context.Context, id string, opts GetChunkOptions) (Chunk, error) {
	chunk, ok := s.store[id]
	if !ok {
		return Chunk{}, fmt.Errorf("no object found with id: %s", id)
	}

	return chunk.Cut(opts)
}
