package inmem

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChunkProxy_PutChunk ensures PutChunk() leaves both the backend and the
// cache with identical content, handling the case where the cache is empty and
// needs re-populating.
func TestChunkProxy_PutChunk(t *testing.T) {
	tests := []struct {
		name  string
		start bool
		// existing store content
		store string
		// existing cache content
		cache string
	}{
		{
			name:  "first chunk",
			start: true,
			store: "",
			cache: "",
		},
		{
			name:  "second chunk",
			store: "first",
			cache: "first",
		},
		{
			name:  "second chunk, empty cache",
			store: "first",
			cache: "",
		},
	}
	for _, tt := range tests {
		store := &testChunkStore{store: map[string][]byte{
			"key": []byte(tt.store),
		}}

		cache := &testCache{cache: make(map[string][]byte)}
		if tt.cache != "" {
			cache.cache["key.log"] = []byte(tt.cache)
		}

		proxy, err := NewChunkProxy(cache, store)
		require.NoError(t, err)

		err = proxy.PutChunk(context.Background(), "key", []byte("_new_chunk"), otf.PutChunkOptions{Start: tt.start})
		require.NoError(t, err)

		// expect cache to have identical content to store
		assert.Equal(t, string(store.store["key"]), string(cache.cache["key.log"]))
	}
}
