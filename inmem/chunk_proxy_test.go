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
	tests := []struct {
		name string
		// existing backend content
		backend map[string][]byte
		// existing cache content
		cache map[string][]byte
		// chunk to be written
		chunk otf.Chunk
	}{
		{
			name: "first chunk",
			chunk: otf.Chunk{
				RunID: "run-123",
				Phase: otf.PlanPhase,
				Data:  []byte("\x02hello"),
			},
			backend: map[string][]byte{},
			cache:   map[string][]byte{},
		},
		{
			name: "second chunk",
			chunk: otf.Chunk{
				RunID: "run-123",
				Phase: otf.PlanPhase,
				Data:  []byte(" world"),
			},
			backend: map[string][]byte{"run-123.plan.log": []byte("\x02hello")},
			cache:   map[string][]byte{"run-123.plan.log": []byte("\x02hello")},
		},
		{
			name: "third and final chunk",
			chunk: otf.Chunk{
				RunID: "run-123",
				Phase: otf.PlanPhase,
				Data:  []byte{0x03},
			},
			backend: map[string][]byte{"run-123.plan.log": []byte("\x02hello world")},
			cache:   map[string][]byte{"run-123.plan.log": []byte("\x02hello world")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := &fakeBackend{store: tt.backend}
			cache := &fakeCache{cache: tt.cache}

			proxy, err := NewChunkProxy(cache, backend)
			require.NoError(t, err)

			_, err = proxy.PutChunk(context.Background(), tt.chunk)
			require.NoError(t, err)

			assert.Equal(t, tt.backend, tt.cache)
		})
	}
}

// TestChunkProxy_GetChunk tests GetChunk() ensuring not only that we get back
// requested chunk but that the chunk is cached if not already.
func TestChunkProxy_GetChunk(t *testing.T) {
	tests := []struct {
		name string
		// existing backend content
		backend map[string][]byte
		// existing cache content
		cache map[string][]byte
		opts  otf.GetChunkOptions
		want  string
	}{
		{
			name: "retrieve from cache",
			backend: map[string][]byte{
				"run-123.plan.log": []byte("\x02hello world\x03"),
			},
			cache: map[string][]byte{
				"run-123.plan.log": []byte("\x02hello world\x03"),
			},
			opts: otf.GetChunkOptions{
				RunID: "run-123",
				Phase: otf.PlanPhase,
			},
			want: "\x02hello world\x03",
		},
		{
			name: "retrieve from backend",
			backend: map[string][]byte{
				"run-123.plan.log": []byte("\x02hello world\x03"),
			},
			cache: map[string][]byte{},
			opts: otf.GetChunkOptions{
				RunID: "run-123",
				Phase: otf.PlanPhase,
			},
			want: "\x02hello world\x03",
		},
		{
			name: "retrieve small chunk from cache",
			backend: map[string][]byte{
				"run-123.plan.log": []byte("\x02hello world\x03"),
			},
			cache: map[string][]byte{
				"run-123.plan.log": []byte("\x02hello world\x03"),
			},
			opts: otf.GetChunkOptions{
				RunID:  "run-123",
				Phase:  otf.PlanPhase,
				Offset: 3,
				Limit:  4,
			},
			want: "llo ",
		},
		{
			name: "retrieve small chunk from backend",
			backend: map[string][]byte{
				"run-123.plan.log": []byte("\x02hello world\x03"),
			},
			cache: map[string][]byte{},
			opts: otf.GetChunkOptions{
				RunID:  "run-123",
				Phase:  otf.PlanPhase,
				Offset: 3,
				Limit:  4,
			},
			want: "llo ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := &fakeBackend{store: tt.backend}
			cache := &fakeCache{cache: tt.cache}

			proxy, err := NewChunkProxy(cache, backend)
			require.NoError(t, err)

			// check we get wanted chunk
			chunk, err := proxy.GetChunk(context.Background(), tt.opts)
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(chunk.Data))

			// check cache always has identical content to backend
			assert.Equal(t, tt.backend, tt.cache)
		})
	}
}
