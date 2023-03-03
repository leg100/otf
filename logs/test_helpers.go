package logs

import (
	"context"
	"errors"

	"github.com/leg100/otf"
)

type fakeCache struct {
	cache map[string][]byte
}

func newFakeCache(keyvalues ...string) *fakeCache {
	cache := make(map[string][]byte, len(keyvalues)/2)
	for i := 0; i < len(keyvalues)/2; i += 2 {
		cache[keyvalues[i]] = []byte(keyvalues[i+1])
	}
	return &fakeCache{cache}
}

func (c *fakeCache) Set(key string, val []byte) error {
	c.cache[key] = val
	return nil
}

func (c *fakeCache) Get(key string) ([]byte, error) {
	val, ok := c.cache[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return val, nil
}

type fakeBackend struct {
	store map[string][]byte
}

func newFakeBackend(keyvalues ...string) *fakeBackend {
	db := make(map[string][]byte, len(keyvalues)/2)
	for i := 0; i < len(keyvalues)/2; i += 2 {
		db[keyvalues[i]] = []byte(keyvalues[i+1])
	}
	return &fakeBackend{db}
}

func (s *fakeBackend) get(ctx context.Context, opts otf.GetChunkOptions) (otf.Chunk, error) {
	key := cacheKey(opts.RunID, opts.Phase)
	data, ok := s.store[key]
	if !ok {
		return otf.Chunk{}, otf.ErrResourceNotFound
	}
	return otf.Chunk{Data: data}, nil
}

func (s *fakeBackend) put(ctx context.Context, chunk otf.Chunk) (otf.PersistedChunk, error) {
	key := cacheKey(chunk.RunID, chunk.Phase)

	if existing, ok := s.store[key]; ok {
		s.store[key] = append(existing, chunk.Data...)
	} else {
		s.store[key] = chunk.Data
	}

	return otf.PersistedChunk{
		ChunkID: 123,
		Chunk:   chunk,
	}, nil
}
