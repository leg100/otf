package organization

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newTestOrg(t *testing.T) *Organization {
	name := uuid.NewString()
	org, err := NewOrganization(CreateOptions{Name: &name})
	require.NoError(t, err)
	return org
}
