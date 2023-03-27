package logs

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProxy_Get tests get() with and without a cached entry
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

		want := otf.Chunk{RunID: "run-123", Phase: otf.PlanPhase, Offset: 3, Data: []byte("lo w")}
		assert.Equal(t, want, got)
	})

	t.Run("cache miss", func(t *testing.T) {
		db := &fakeDB{data: []byte("hello world")}
		cache := newFakeCache()
		proxy := &proxy{cache: cache, db: db}

		got, err := proxy.get(ctx, opts)
		require.NoError(t, err)

		want := otf.Chunk{RunID: "run-123", Phase: otf.PlanPhase, Offset: 3, Data: []byte("lo w")}
		assert.Equal(t, want, got)

		// cache should be populated now
		assert.Equal(t, "hello world", string(cache.cache["run-123.plan.log"]))
	})
}
