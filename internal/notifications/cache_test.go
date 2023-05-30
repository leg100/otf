package notifications

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_New(t *testing.T) {
	nc1 := newTestConfig(t, "", DestinationSlack, "http://example.com")
	nc2 := newTestConfig(t, "", DestinationSlack, "http://example.com")
	nc3 := newTestConfig(t, "", DestinationGCPPubSub, "gcppubsub://project1/topic1")

	cache := newTestCache(t, nil, nc1, nc2, nc3)

	assert.Equal(t, 3, len(cache.configs))
	assert.Equal(t, 2, len(cache.clients))
}

func TestCache_AddRemove(t *testing.T) {
	cache := newTestCache(t, nil)
	nc1 := newTestConfig(t, "", DestinationSlack, "http://example.com")
	nc2 := newTestConfig(t, "", DestinationSlack, "http://example.com")

	err := cache.add(nc1)
	require.NoError(t, err)

	err = cache.add(nc2)
	require.NoError(t, err)

	// both configs should share client
	assert.Equal(t, 2, len(cache.configs))
	assert.Equal(t, 1, len(cache.clients))

	err = cache.remove(nc1.ID)
	require.NoError(t, err)

	// client should not have been removed
	assert.Equal(t, 1, len(cache.configs))
	assert.Equal(t, 1, len(cache.clients))

	err = cache.remove(nc2.ID)
	require.NoError(t, err)

	// client should now have been removed
	assert.Equal(t, 0, len(cache.configs))
	assert.Equal(t, 0, len(cache.clients))
}
