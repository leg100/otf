package logs

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProxy_Put checks put() results in identical data in both the db and
// the cache
func TestProxy_Put(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		// existing backend content
		backend *fakeBackend
		// existing cache content
		cache *fakeCache
		// chunk to be written
		chunk otf.Chunk
	}{
		{
			name:    "first chunk",
			chunk:   otf.Chunk{"run-123", otf.PlanPhase, 0, []byte("\x02hello")},
			backend: newFakeBackend(),
			cache:   newFakeCache(),
		},
		{
			name:    "second chunk",
			chunk:   otf.Chunk{"run-123", otf.PlanPhase, 0, []byte(" world")},
			backend: newFakeBackend("run-123.plan.log", "\x02hello"),
			cache:   newFakeCache("run-123.plan.log", "\x02hello"),
		},
		{
			name:    "third and final chunk",
			chunk:   otf.Chunk{"run-123", otf.PlanPhase, 0, []byte{0x03}},
			backend: newFakeBackend("run-123.plan.log", "\x02hello world"),
			cache:   newFakeCache("run-123.plan.log", "\x02hello world"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy := &proxy{cache: tt.cache, db: tt.backend}

			_, err := proxy.put(ctx, tt.chunk)
			require.NoError(t, err)

			assert.Equal(t, tt.backend.store, tt.cache.cache)
		})
	}
}

// TestProxy_Get checks get() with and without a cached entry
func TestProxy_Get(t *testing.T) {
	ctx := context.Background()

	opts := otf.GetChunkOptions{
		RunID:  "run-123",
		Phase:  otf.PlanPhase,
		Offset: 3,
		Limit:  4,
	}

	t.Run("cache hit", func(t *testing.T) {
		cache := newFakeCache("run-123.plan.log", "hello world")
		proxy := &proxy{cache: cache}

		got, err := proxy.get(ctx, opts)
		require.NoError(t, err)

		want := otf.Chunk{Offset: 3, Data: []byte("lo w")}
		assert.Equal(t, want, got)
	})

	t.Run("cache miss", func(t *testing.T) {
		db := newFakeBackend("run-123.plan.log", "hello world")
		cache := newFakeCache()
		proxy := &proxy{cache: cache, db: db}

		got, err := proxy.get(ctx, opts)
		require.NoError(t, err)

		want := otf.Chunk{Offset: 3, Data: []byte("lo w")}
		assert.Equal(t, want, got)

		// cache should be populated now
		assert.Equal(t, "hello world", string(cache.cache["run-123.plan.log"]))
	})
}
