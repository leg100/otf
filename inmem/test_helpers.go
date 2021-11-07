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

type testChunkStore struct {
	store map[string][]byte

	otf.ChunkStore
}

func newTestChunkStore() *testChunkStore { return &testChunkStore{store: make(map[string][]byte)} }

func (s *testChunkStore) GetChunk(ctx context.Context, id string, opts otf.GetChunkOptions) ([]byte, error) {
	if opts.Limit == 0 {
		return s.store[id][opts.Offset:], nil
	}
	return s.store[id][opts.Offset : opts.Offset+opts.Limit], nil
}

func (s *testChunkStore) PutChunk(ctx context.Context, id string, chunk []byte, opts otf.PutChunkOptions) error {
	if val, ok := s.store[id]; ok {
		s.store[id] = append(val, chunk...)
	} else {
		s.store[id] = chunk
	}

	return nil
}
