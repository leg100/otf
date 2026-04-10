package organization

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrganizationDefaultsSentinelVersion(t *testing.T) {
	org, err := NewOrganization(CreateOptions{
		Name: new(t.Name()),
	})
	require.NoError(t, err)

	assert.Equal(t, DefaultSentinelVersion, org.SentinelVersion)
}

func TestNewOrganizationAllowsExplicitSentinelVersion(t *testing.T) {
	org, err := NewOrganization(CreateOptions{
		Name:            new(t.Name()),
		SentinelVersion: new("0.40.0"),
	})
	require.NoError(t, err)

	assert.Equal(t, "0.40.0", org.SentinelVersion)
}

func TestNewOrganizationRejectsInvalidSentinelVersion(t *testing.T) {
	_, err := NewOrganization(CreateOptions{
		Name:            new(t.Name()),
		SentinelVersion: new("not-a-version"),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sentinel version")
}

func TestOrganizationUpdateSentinelVersion(t *testing.T) {
	org, err := NewOrganization(CreateOptions{
		Name: new(t.Name()),
	})
	require.NoError(t, err)

	err = org.Update(UpdateOptions{
		SentinelVersion: new("0.40.0"),
	})
	require.NoError(t, err)

	assert.Equal(t, "0.40.0", org.SentinelVersion)
}
