package logs

import (
	"context"
	"errors"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/rbac"
)

type (
	fakeCache struct {
		cache map[string][]byte
	}

	fakeDB struct {
		data []byte
		proxydb
	}

	fakeTailProxy struct {
		// fake chunk to return
		chunk internal.Chunk
		chunkproxy
	}

	fakeAuthorizer struct{}
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

func (s *fakeDB) getLogs(ctx context.Context, runID string, phase internal.PhaseType) ([]byte, error) {
	return s.data, nil
}

func (f *fakeTailProxy) get(ctx context.Context, opts internal.GetChunkOptions) (internal.Chunk, error) {
	return f.chunk, nil
}

func (f *fakeAuthorizer) CanAccess(context.Context, rbac.Action, string) (authz.Subject, error) {
	return &authz.Superuser{}, nil
}

type fakeSubService struct {
	stream chan pubsub.Event[internal.Chunk]

	pubsub.SubscriptionService[internal.Chunk]
}

func (f *fakeSubService) Subscribe(ctx context.Context) (<-chan pubsub.Event[internal.Chunk], func()) {
	go func() {
		<-ctx.Done()
		close(f.stream)
	}()
	return f.stream, nil
}
