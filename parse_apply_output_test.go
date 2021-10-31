package otf

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseApplyOutputChanges(t *testing.T) {
	want := Resources{
		ResourceAdditions:    2,
		ResourceChanges:      0,
		ResourceDestructions: 0,
	}

	output, err := os.ReadFile("testdata/apply.txt")
	require.NoError(t, err)

	apply, err := ParseApplyOutput(string(output))
	require.NoError(t, err)
	assert.Equal(t, want, apply)
}

func TestParseApplyOutputNoChanges(t *testing.T) {
	want := Resources{
		ResourceAdditions:    0,
		ResourceChanges:      0,
		ResourceDestructions: 0,
	}

	output, err := ioutil.ReadFile("testdata/apply_no_changes.txt")
	require.NoError(t, err)

	apply, err := ParseApplyOutput(string(output))
	require.NoError(t, err)
	assert.Equal(t, want, apply)
}
