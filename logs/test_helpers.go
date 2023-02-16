package logs

import (
	"context"
	"errors"

	"github.com/leg100/otf"
)

type fakeCache struct {
	cache map[string][]byte

	otf.Cache
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
	ChunkStore
}

func (s *fakeBackend) GetChunk(ctx context.Context, opts GetChunkOptions) (otf.Chunk, error) {
	if s.store == nil {
		// avoid panics
		s.store = make(map[string][]byte)
	}

	data, ok := s.store[opts.Key()]
	if !ok {
		return Chunk{}, errors.New("not found")
	}
	return Chunk{
		RunID: opts.RunID,
		Phase: opts.Phase,
		Data:  data,
	}, nil
}

func (s *fakeBackend) PutChunk(ctx context.Context, chunk Chunk) (PersistedChunk, error) {
	if existing, ok := s.store[chunk.Key()]; ok {
		s.store[chunk.Key()] = append(existing, chunk.Data...)
	} else {
		s.store[chunk.Key()] = chunk.Data
	}

	return PersistedChunk{
		ChunkID: 123,
		Chunk:   chunk,
	}, nil
}
