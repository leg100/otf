package logs

import (
	"context"
	"errors"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
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
		chunk Chunk
		chunkproxy
	}

	fakeAuthorizer struct {
		authz.Interface
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

func (s *fakeDB) getLogs(ctx context.Context, runID resource.ID, phase internal.PhaseType) ([]byte, error) {
	return s.data, nil
}

func (f *fakeTailProxy) get(ctx context.Context, opts GetChunkOptions) (Chunk, error) {
	return f.chunk, nil
}

func (f *fakeAuthorizer) Authorize(context.Context, authz.Action, *authz.AccessRequest, ...authz.CanAccessOption) (authz.Subject, error) {
	return &authz.Superuser{}, nil
}

type fakeSubService struct {
	stream chan pubsub.Event[Chunk]

	pubsub.SubscriptionService[Chunk]
}

func (f *fakeSubService) Subscribe(ctx context.Context) (<-chan pubsub.Event[Chunk], func()) {
	go func() {
		<-ctx.Done()
		close(f.stream)
	}()
	return f.stream, nil
}
