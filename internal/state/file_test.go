package state

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFile_Unmarshal(t *testing.T) {
	data, err := os.ReadFile("testdata/terraform.tfstate")
	require.NoError(t, err)

	var got File
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, 4, got.Version)
	assert.Equal(t, int64(9), got.Serial)
	assert.Equal(t, "f1d86b13-cf61-8c41-9cc9-bde8a04e94b4", got.Lineage)
	if assert.Equal(t, 3, len(got.Outputs)) {
		if assert.Contains(t, got.Outputs, "foo") {
			assert.True(t, got.Outputs["foo"].Sensitive)
		}
		assert.Contains(t, got.Outputs, "bar")
		assert.Contains(t, got.Outputs, "baz")
	}
	// skip testing output values because they're not unmarshaled
}
