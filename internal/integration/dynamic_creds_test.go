package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/dynamiccreds"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/variable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDynamicCredentials tests enabling dynamic provider credentials.
func TestDynamicCredentials(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t, withKeyPairPaths(
		"./fixtures/private_key.pem",
		"./fixtures/public_key.pem",
	))
	ws1 := daemon.createWorkspace(t, ctx, org)
	_ = daemon.createVariable(t, ctx, ws1, &variable.CreateVariableOptions{
		Key:      internal.Ptr("TFC_GCP_PROVIDER_AUTH"),
		Value:    internal.Ptr("true"),
		Category: internal.Ptr(variable.CategoryEnv),
	})
	_ = daemon.createVariable(t, ctx, ws1, &variable.CreateVariableOptions{
		Key:      internal.Ptr("TFC_GCP_WORKLOAD_IDENTITY_AUDIENCE"),
		Value:    internal.Ptr("acme.audience"),
		Category: internal.Ptr(variable.CategoryEnv),
	})
	_ = daemon.createVariable(t, ctx, ws1, &variable.CreateVariableOptions{
		Key:      internal.Ptr("TFC_GCP_RUN_SERVICE_ACCOUNT_EMAIL"),
		Value:    internal.Ptr("terraform@iam.google.com"),
		Category: internal.Ptr(variable.CategoryEnv),
	})
	_ = daemon.createVariable(t, ctx, ws1, &variable.CreateVariableOptions{
		Key:      internal.Ptr("TFC_GCP_WORKLOAD_PROVIDER_NAME"),
		Value:    internal.Ptr("projects/123/locations/global/workloadIdentityPools/pool-123/providers/provider-123"),
		Category: internal.Ptr(variable.CategoryEnv),
	})
	cv1 := daemon.createAndUploadConfigurationVersion(t, ctx, ws1, nil)
	run := daemon.createRun(t, ctx, ws1, cv1, nil)
	daemon.waitRunStatus(t, ctx, run.ID, runstatus.Planned)

	resp, err := http.Get(daemon.System.URL("/.well-known/openid-configuration"))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	var cfg dynamiccreds.WellKnownConfig
	err = json.NewDecoder(resp.Body).Decode(&cfg)
	require.NoError(t, err)
	assert.Equal(t, []string{"RS256"}, cfg.Algorithms)

	resp, err = http.Get(daemon.System.URL("/.well-known/jwks"))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
}
