package cache

import (
	"bytes"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLRU_PutGet(t *testing.T) {
	cache := NewLRUMemstore(100)

	require.NoError(t, cache.Put("foo", []byte("bar")))

	got, err := cache.Get("foo")
	require.NoError(t, err)

	assert.Equal(t, "bar", string(got))
}

// TestLRU_Evict tests the LRU having to evict an item because adding item will
// breach capacity.
func TestLRU_Evict(t *testing.T) {
	cache := NewLRUMemstore(100)

	require.NoError(t, cache.Put("evict_this", bytes.Repeat([]byte("a"), 99)))
	require.NoError(t, cache.Put("new_item", bytes.Repeat([]byte("b"), 99)))

	_, err := cache.Get("evict_this")
	assert.Error(t, err, otf.ErrResourceNotFound)
}

// TestLRU_ExceedCapacity tries to put an item into the cache that exceeds total
// capacity of cache.
func TestLRU_ExceedCapacity(t *testing.T) {
	cache := NewLRUMemstore(100)

	assert.Error(t, cache.Put("k1", bytes.Repeat([]byte("z"), 101)))
}
