package logs

import (
	"context"
	"errors"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

type (
	fakeCache struct {
		cache map[string][]byte
	}

	fakeDB struct {
		data []byte
		db
	}

	fakeTailProxy struct {
		// fake chunk to return
		chunk otf.Chunk
		chunkproxy
	}

	fakeAuthorizer struct{}

	fakePubSubTailService struct {
		stream chan otf.Event
	}
)

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

func (s *fakeDB) GetLogs(ctx context.Context, runID string, phase otf.PhaseType) ([]byte, error) {
	return s.data, nil
}

func (f *fakeTailProxy) get(ctx context.Context, opts otf.GetChunkOptions) (otf.Chunk, error) {
	return f.chunk, nil
}

func newFakePubSubService() *fakePubSubTailService {
	return &fakePubSubTailService{stream: make(chan otf.Event)}
}

func (f *fakePubSubTailService) Subscribe(context.Context, string) (<-chan otf.Event, error) {
	return f.stream, nil
}

func (f *fakePubSubTailService) Publish(event otf.Event) {
	f.stream <- event
}

func (f *fakeAuthorizer) CanAccess(context.Context, rbac.Action, string) (otf.Subject, error) {
	return &otf.Superuser{}, nil
}
