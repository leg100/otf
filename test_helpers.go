package otf

import (
	"context"
	"fmt"
)

type testChunkStore struct {
	store map[string][]byte

	ChunkStore
}

func (s *testChunkStore) GetChunk(ctx context.Context, id string, opts GetChunkOptions) ([]byte, error) {
	data, ok := s.store[id]
	if !ok {
		return nil, fmt.Errorf("no object found with id: %s", id)
	}

	return GetChunk(data, opts)
}
