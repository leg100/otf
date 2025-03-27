package testutils

import (
	"testing"

	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/require"
)

func ParseID(t *testing.T, s string) resource.TfeID {
	t.Helper()

	id, err := resource.ParseTfeID(s)
	require.NoError(t, err)
	return id
}
