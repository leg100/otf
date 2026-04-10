package policy

import (
	"testing"

	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadRepoPolicyBundleRejectsUnsupportedKind(t *testing.T) {
	service := &Service{}

	_, err := service.loadRepoPolicyBundle(t.Context(), OPAKind, resource.TfeID{}, vcs.Repo{}, "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not support VCS imports yet")
}

func TestOPAEvaluatorNotImplemented(t *testing.T) {
	evaluator := &opaEvaluator{}

	_, err := evaluator.Evaluate(t.Context(), &PolicySet{Kind: OPAKind, EngineVersion: LatestVersion()}, &MockBundle{}, nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}

func TestApplyOrganizationPolicySettingsOverridesSentinelVersion(t *testing.T) {
	version := Version{}
	require.NoError(t, version.UnmarshalText([]byte("0.39.0")))

	got, err := applyOrganizationPolicySettings(&PolicySet{
		Kind:          SentinelKind,
		EngineVersion: version,
	}, &organization.Organization{
		SentinelVersion: "0.40.0",
	})
	require.NoError(t, err)

	assert.Equal(t, "0.40.0", got.EngineVersion.String())
}

func TestApplyOrganizationPolicySettingsLeavesOPAUnchanged(t *testing.T) {
	version := Version{}
	require.NoError(t, version.UnmarshalText([]byte("1.0.0")))

	got, err := applyOrganizationPolicySettings(&PolicySet{
		Kind:          OPAKind,
		EngineVersion: version,
	}, &organization.Organization{
		SentinelVersion: "0.40.0",
	})
	require.NoError(t, err)

	assert.Equal(t, OPAKind, got.Kind)
	assert.Equal(t, "1.0.0", got.EngineVersion.String())
}

func TestApplyOrganizationPolicySettingsRejectsInvalidOrganizationSentinelVersion(t *testing.T) {
	_, err := applyOrganizationPolicySettings(&PolicySet{
		Kind: SentinelKind,
	}, &organization.Organization{
		SentinelVersion: "bogus",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing organization sentinel version")
}
