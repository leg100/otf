package otf

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestState_UnmarshalState(t *testing.T) {
	data, err := os.ReadFile("testdata/terraform.tfstate")
	require.NoError(t, err)

	state, err := UnmarshalState(data)
	require.NoError(t, err)

	assert.Equal(t, state, &State{
		Version: 4,
		Serial:  2,
		Lineage: "b2b54b23-e7ea-5500-7b15-fcb68c1d92bb",
		Outputs: map[string]StateOutput{
			"test_output": {
				Value: "9023256633839603543",
				Type:  "string",
			},
		},
	})
}
