package testutils

import (
	"testing"

	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/require"
)

func ParseID(t *testing.T, s string) resource.ID {
	t.Helper()

	id, err := resource.ParseTfeID(s)
	require.NoError(t, err)
	return id
}
