package otf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBranch(t *testing.T) {
	t.Run("master branch", func(t *testing.T) {
		got, found := ParseBranch("refs/heads/master")
		require.True(t, found)
		assert.Equal(t, "master", got)
	})

	t.Run("tag", func(t *testing.T) {
		_, found := ParseBranch("refs/tags/v0.0.1")
		require.False(t, found)
	})
}
