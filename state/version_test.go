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
	opts := otf.CreateStateVersionOptions{
		Serial:      otf.Int64(999),
		State:       state,
		WorkspaceID: otf.String("ws-123"),
	}
	got, err := newVersion(opts)
	require.NoError(t, err)
	assert.Equal(t, int64(999), got.serial)
	assert.Equal(t, "ws-123", got.workspaceID)
	assert.Equal(t, 3, len(got.outputs))

	assert.Equal(t, "foo", got.outputs["foo"].name)
	assert.Equal(t, "string", got.outputs["foo"].typ)
	assert.Equal(t, `"stringy"`, got.outputs["foo"].value)
	assert.True(t, got.outputs["foo"].sensitive)

	assert.Equal(t, "bar", got.outputs["bar"].name)
	assert.Equal(t, "tuple", got.outputs["bar"].typ)
	assert.Equal(t, `["item1","item2"]`, compactJSON(t, got.outputs["bar"].value))
	assert.False(t, got.outputs["bar"].sensitive)

	assert.Equal(t, "baz", got.outputs["baz"].name)
	assert.Equal(t, "object", got.outputs["baz"].typ)
	assert.Equal(t, `{"key1":"value1","key2":"value2"}`, compactJSON(t, got.outputs["baz"].value))
	assert.False(t, got.outputs["baz"].sensitive)
}

func compactJSON(t *testing.T, src string) string {
	var buf bytes.Buffer
	err := json.Compact(&buf, []byte(src))
	require.NoError(t, err)
	return buf.String()
}
