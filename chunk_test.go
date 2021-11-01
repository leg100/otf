package otf

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStream(t *testing.T) {
	store := testChunkStore{
		store: map[string][]byte{
			"test-123": []byte("\x02cat sat on the map\x03"),
		},
	}

	buf := new(bytes.Buffer)

	err := Stream(context.Background(), "test-123", &store, buf, time.Millisecond, 1)
	require.NoError(t, err)
}
