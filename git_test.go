package otf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBranch(t *testing.T) {
	t.Run("master branch", func(t *testing.T) {
		got, found := ParseBranchRef("refs/heads/master")
		require.True(t, found)
		assert.Equal(t, "master", got)
	})

	t.Run("tag", func(t *testing.T) {
		_, found := ParseBranchRef("refs/tags/v0.0.1")
		require.False(t, found)
	})
}
