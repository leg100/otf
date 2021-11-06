package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCacheLogChunk checks that cacheLogChunk() does indeed ensure the cached
// content matches the store content after caching a chunk.
func TestCacheLogChunk(t *testing.T) {
	tests := []struct {
		name  string
		start bool
		// chunk being cached
		chunk string
		// existing store content
		store string
		// existing cache content
		cache string
	}{
		{
			name:  "first chunk",
			start: true,
			chunk: "first",
			store: "first",
		},
		{
			name:  "second chunk",
			chunk: "_second",
			store: "first_second",
			cache: "first",
		},
		{
			name:  "second chunk but empty cache",
			chunk: "_second",
			store: "first_second",
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

		err := cacheLogChunk(context.Background(), cache, store, "key", []byte(tt.chunk), tt.start)
		require.NoError(t, err)

		// expect cache to have identical content to store
		assert.Equal(t, string(store.store["key"]), string(cache.cache["key.log"]))
	}
}
