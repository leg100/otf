package ots

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	data, err := os.ReadFile("testdata/terraform.tfstate")
	require.NoError(t, err)

	state, err := Parse(data)
	require.NoError(t, err)

	assert.Equal(t, state, &State{
		Version: 4,
		Serial:  2,
		Outputs: map[string]StateOutput{
			"test_output": {
				Value: "9023256633839603543",
				Type:  "string",
			},
		},
	})
}
