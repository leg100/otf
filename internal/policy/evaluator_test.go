package policy

import (
	"testing"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLIEvaluatorMissingBinaryProducesFailedCheck(t *testing.T) {
	evaluator := &cliEvaluator{
		binPath: "definitely-not-a-real-sentinel-binary",
		workDir: t.TempDir(),
	}

	checks, err := evaluator.Evaluate(t.Context(), &MockBundle{
		Files: map[string][]byte{
			"sentinel.hcl":              []byte(sentinelConfig()),
			"mock-tfrun.sentinel":       []byte(`workspace = {"name": "nexus"}`),
			"mock-tfplan.sentinel":      []byte(`resource_changes = {}`),
			"mock-tfplan-v2.sentinel":   []byte(`resource_changes = {}`),
			"mock-tfstate.sentinel":     []byte(`resources = {}`),
			"mock-tfstate-v2.sentinel":  []byte(`resources = {}`),
			"mock-tfconfig.sentinel":    []byte(`module = {}`),
			"mock-tfconfig-v2.sentinel": []byte(`module = {}`),
		},
	}, []*Policy{{
		Name:             "workspace-name",
		Source:           `import "tfrun" ; main = rule { tfrun.workspace.name is "nexus" }`,
		EnforcementLevel: MandatoryEnforcement,
	}}, nil)

	require.NoError(t, err)
	require.Len(t, checks, 1)
	assert.False(t, checks[0].Passed)
	assert.Contains(t, checks[0].Output, "sentinel execution error")
}

func TestGeneratedModuleFilenameUsesStableLocalName(t *testing.T) {
	got := generatedModuleFilename(&PolicyModule{
		PolicySetID: resource.NewTfeID(resource.PolicySetKind),
		Name:        "tfplan-functions",
		Path:        "common_functions/tfplan-functions.sentinel",
	})

	assert.Equal(t, "tfplan_functions.sentinel", got)
}

func TestMatchesPolicySetRef(t *testing.T) {
	event := vcs.Event{
		EventPayload: vcs.EventPayload{
			Branch:        "main",
			DefaultBranch: "main",
			CommitSHA:     "abc123",
		},
	}

	assert.True(t, matchesPolicySetRef("", event))
	assert.True(t, matchesPolicySetRef("main", event))
	assert.True(t, matchesPolicySetRef("refs/heads/main", event))
	assert.True(t, matchesPolicySetRef("abc123", event))
	assert.False(t, matchesPolicySetRef("develop", event))
}
