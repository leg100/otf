package testutils

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func CompactJSON(t *testing.T, src string) string {
	var buf bytes.Buffer
	err := json.Compact(&buf, []byte(src))
	require.NoError(t, err)
	return buf.String()
}
