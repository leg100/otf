package inmem

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChunkProxy_PutChunk ensures PutChunk() leaves both the backend and the
// cache with identical content.
func TestChunkProxy_PutChunk(t *testing.T) {
	id := "key"

	tests := []struct {
		name string
		// existing backend content
		backend *otf.Chunk
		// existing cache content
		cache []byte
		// chunk to be written
		chunk otf.Chunk
	}{
		{
			name:  "first chunk",
			chunk: otf.Chunk{Data: []byte("hello"), Start: true},
		},
		{
			name:    "second chunk",
			chunk:   otf.Chunk{Data: []byte(" world")},
			backend: &otf.Chunk{Data: []byte("hello "), Start: true},
			cache:   []byte("\x02hello"),
		},
		{
			name: "second chunk, empty cache",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup backend
			backend := newTestChunkStore()
			if tt.backend != nil {
				backend.store[id] = *tt.backend
			}

			// setup cache
			cache := newTestCache()
			if tt.cache != nil {
				cache.cache[id] = tt.cache
			}

			proxy, err := NewChunkProxy(cache, backend)
			require.NoError(t, err)

			err = proxy.PutChunk(context.Background(), id, otf.PlanPhase, tt.chunk)
			require.NoError(t, err)

			// expect cache to have identical content to store
			assert.Equal(t, string(backend.store[id].Marshal()), string(cache.cache[otf.LogCacheKey(id, otf.PlanPhase)]))
		})
	}
}

// TestChunkProxy_GetChunk_FromCache tests retrieving a chunk from the cache.
func TestChunkProxy_GetChunk_FromCache(t *testing.T) {
	id := "key"
	store := newTestChunkStore()
	cache := newTestCache()

	proxy, err := NewChunkProxy(cache, store)
	require.NoError(t, err)

	cache.cache[otf.LogCacheKey(id, otf.PlanPhase)] = []byte("\x02abcdefghijkl\x03")

	chunk, err := proxy.GetChunk(context.Background(), id, otf.PlanPhase, otf.GetChunkOptions{})
	require.NoError(t, err)

	assert.Equal(t, "\x02abcdefghijkl\x03", string(chunk.Marshal()))
}

// TestChunkProxy_GetChunk_FromStore tests retrieving a chunk from the backend,
// and that the cache is re-populated.
func TestChunkProxy_GetChunk_FromStore(t *testing.T) {
	store := newTestChunkStore()
	cache := newTestCache()

	proxy, err := NewChunkProxy(cache, store)
	require.NoError(t, err)

	store.store["key"] = otf.Chunk{Data: []byte("abcdefghijkl")}

	chunk, err := proxy.GetChunk(context.Background(), "key", otf.PlanPhase, otf.GetChunkOptions{})
	require.NoError(t, err)

	assert.Equal(t, "abcdefghijkl", string(chunk.Data))
	assert.Equal(t, "abcdefghijkl", string(cache.cache["key.plan.log"]))
}
