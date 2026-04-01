package policy

import (
	"testing"

	"github.com/leg100/otf/internal/organization"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPolicySetDefaults(t *testing.T) {
	org := organization.NewTestName(t)

	set, err := newPolicySet(org, CreatePolicySetOptions{
		Name: new(t.Name()),
	})
	require.NoError(t, err)

	assert.Equal(t, SentinelKind, set.Kind)
	assert.True(t, set.EngineVersion.Latest)
	assert.Equal(t, "latest", set.EngineVersion.String())
}

func TestNewPolicySetExplicitKindAndVersion(t *testing.T) {
	org := organization.NewTestName(t)
	kind := OPAKind
	version := Version{}
	require.NoError(t, version.UnmarshalText([]byte("1.0.0")))

	set, err := newPolicySet(org, CreatePolicySetOptions{
		Name:          new(t.Name()),
		Kind:          &kind,
		EngineVersion: &version,
	})
	require.NoError(t, err)

	assert.Equal(t, OPAKind, set.Kind)
	assert.False(t, set.EngineVersion.Latest)
	assert.Equal(t, "1.0.0", set.EngineVersion.String())
}

func TestNewPolicySetRejectsInvalidKind(t *testing.T) {
	org := organization.NewTestName(t)
	kind := Kind("bogus")

	_, err := newPolicySet(org, CreatePolicySetOptions{
		Name: new(t.Name()),
		Kind: &kind,
	})
	require.Error(t, err)
}
