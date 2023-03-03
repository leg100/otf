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
	assert.Equal(t, int64(999), got.Serial)
	assert.Equal(t, "ws-123", got.WorkspaceID)
	assert.Equal(t, 3, len(got.Outputs))

	assert.Equal(t, "foo", got.Outputs["foo"].name)
	assert.Equal(t, "string", got.Outputs["foo"].typ)
	assert.Equal(t, `"stringy"`, got.Outputs["foo"].value)
	assert.True(t, got.Outputs["foo"].sensitive)

	assert.Equal(t, "bar", got.Outputs["bar"].name)
	assert.Equal(t, "tuple", got.Outputs["bar"].typ)
	assert.Equal(t, `["item1","item2"]`, compactJSON(t, got.Outputs["bar"].value))
	assert.False(t, got.Outputs["bar"].sensitive)

	assert.Equal(t, "baz", got.Outputs["baz"].name)
	assert.Equal(t, "object", got.Outputs["baz"].typ)
	assert.Equal(t, `{"key1":"value1","key2":"value2"}`, compactJSON(t, got.Outputs["baz"].value))
	assert.False(t, got.Outputs["baz"].sensitive)
}

func compactJSON(t *testing.T, src string) string {
	var buf bytes.Buffer
	err := json.Compact(&buf, []byte(src))
	require.NoError(t, err)
	return buf.String()
}
