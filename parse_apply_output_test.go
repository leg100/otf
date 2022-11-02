package otf

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseApplyOutputChanges(t *testing.T) {
	want := ResourceReport{
		Additions:    1,
		Changes:      0,
		Destructions: 0,
	}

	output, err := os.ReadFile("testdata/apply.txt")
	require.NoError(t, err)

	apply, err := ParseApplyOutput(string(output))
	require.NoError(t, err)
	assert.Equal(t, want, apply)
}

func TestParseApplyOutputNoChanges(t *testing.T) {
	want := ResourceReport{
		Additions:    0,
		Changes:      0,
		Destructions: 0,
	}

	output, err := os.ReadFile("testdata/apply_no_changes.txt")
	require.NoError(t, err)

	apply, err := ParseApplyOutput(string(output))
	require.NoError(t, err)
	assert.Equal(t, want, apply)
}
