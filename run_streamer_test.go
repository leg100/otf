package otf

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunStreamer(t *testing.T) {
	store := testChunkStore{
		store: map[string][]byte{
			"plan-123": []byte("\x02cat sat on the map\x03"),
		},
	}

	run := Run{
		Plan:  &Plan{ID: "plan-123"},
		Apply: &Apply{ID: "plan-123"},
	}

	streamer := NewRunStreamer(&run, &store, &store, time.Millisecond)
	go streamer.Stream(context.Background())

	got, err := io.ReadAll(streamer)
	require.NoError(t, err)

	assert.Equal(t, "abc", string(got))
}
