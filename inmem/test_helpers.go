package inmem

import (
	"context"
	"errors"

	"github.com/leg100/otf"
)

type testCache struct {
	cache map[string][]byte

	otf.Cache
}

func newTestCache() *testCache { return &testCache{cache: make(map[string][]byte)} }

func (c *testCache) Set(key string, val []byte) error {
	c.cache[key] = val

	return nil
}

func (c *testCache) Get(key string) ([]byte, error) {
	val, ok := c.cache[key]
	if !ok {
		return nil, errors.New("not found")
	}

	return val, nil
}

func (c *testCache) AppendChunk(key string, chunk []byte) error {
	c.cache[key] = append(c.cache[key], chunk...)

	return nil
}

var _ otf.ChunkStore = (*testChunkStore)(nil)

type testChunkStore struct {
	store map[string]otf.Chunk
}

func newTestChunkStore() *testChunkStore { return &testChunkStore{store: make(map[string]otf.Chunk)} }

func (s *testChunkStore) GetChunk(ctx context.Context, id string, phase otf.PhaseType, opts otf.GetChunkOptions) (otf.Chunk, error) {
	if opts.Limit == 0 {
		return s.store[id].Cut(opts)
	}
	return s.store[id].Cut(opts)
}

func (s *testChunkStore) PutChunk(ctx context.Context, id string, phase otf.PhaseType, chunk otf.Chunk) error {
	if val, ok := s.store[id]; ok {
		s.store[id] = val.Append(chunk)
	} else {
		s.store[id] = chunk
	}

	return nil
}
