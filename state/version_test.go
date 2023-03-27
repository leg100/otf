package state

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersion_new(t *testing.T) {
	state, err := os.ReadFile("testdata/terraform.tfstate")
	require.NoError(t, err)
	opts := CreateStateVersionOptions{
		Serial:      otf.Int64(999),
		State:       state,
		WorkspaceID: otf.String("ws-123"),
	}
	got, err := newVersion(opts)
	require.NoError(t, err)
	assert.Equal(t, int64(999), got.Serial)
	assert.Equal(t, "ws-123", got.WorkspaceID)
	assert.Equal(t, 3, len(got.Outputs))

	assert.Equal(t, "foo", got.Outputs["foo"].Name)
	assert.Equal(t, "string", got.Outputs["foo"].Type)
	assert.Equal(t, `"stringy"`, got.Outputs["foo"].Value)
	assert.True(t, got.Outputs["foo"].Sensitive)

	assert.Equal(t, "bar", got.Outputs["bar"].Name)
	assert.Equal(t, "tuple", got.Outputs["bar"].Type)
	assert.Equal(t, `["item1","item2"]`, compactJSON(t, got.Outputs["bar"].Value))
	assert.False(t, got.Outputs["bar"].Sensitive)

	assert.Equal(t, "baz", got.Outputs["baz"].Name)
	assert.Equal(t, "object", got.Outputs["baz"].Type)
	assert.Equal(t, `{"key1":"value1","key2":"value2"}`, compactJSON(t, got.Outputs["baz"].Value))
	assert.False(t, got.Outputs["baz"].Sensitive)
}

func compactJSON(t *testing.T, src string) string {
	var buf bytes.Buffer
	err := json.Compact(&buf, []byte(src))
	require.NoError(t, err)
	return buf.String()
}
