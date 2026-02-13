package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/runner"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDynamicCredentialsGCP tests dynamic provider credentials on GCP.
//
// NOTE: this test requires significant manual configuration and if any of the
// expected env vars are missing it'll be skipped. Consult the docs [1] on how
// to setup dynamic creds on GCP and specify the env vars accordingly.
//
// NOTE: dynamic credentials usually require otfd to publicly expose a couple of
// endpoints over HTTPS, via which the provider validates the JWKS. But that
// won't work for this test, not least because the daemon is run on a random
// port. Instead, you should manually upload the JWKS to GCP first [2] (run
// `otfd` with dynamic credentials enabled and then retrieve the JWKS from
// `https://localhost:8080/.well-known/jwks` and run the command in the linked
// document). GCP will then not attempt to connect to the daemon.
//
// [1]: https://docs.otf.ninja/dynamic_credentials/
// [2]: https://cloud.google.com/iam/docs/workload-identity-federation-with-other-providers#manage-oidc-keys
func TestDynamicCredentialsGCP(t *testing.T) {
	integrationTest(t)

	privateKeyPath, ok := os.LookupEnv("OTF_INTEGRATION_PRIVATE_KEY_PATH")
	if !ok {
		t.Skip("OTF_INTEGRATION_PRIVATE_KEY_PATH needed for dynamic credentials integration test")
	}
	publicKeyPath, ok := os.LookupEnv("OTF_INTEGRATION_PUBLIC_KEY_PATH")
	if !ok {
		t.Skip("OTF_INTEGRATION_PUBLIC_KEY_PATH needed for dynamic credentials integration test")
	}
	workload, ok := os.LookupEnv("OTF_INTEGRATION_TFC_GCP_WORKLOAD_PROVIDER_NAME")
	if !ok {
		t.Skip("OTF_INTEGRATION_TFC_GCP_WORKLOAD_PROVIDER_NAME needed for dynamic credentials integration test")
	}
	serviceAccount, ok := os.LookupEnv("OTF_INTEGRATION_TFC_GCP_RUN_SERVICE_ACCOUNT_EMAIL")
	if !ok {
		t.Skip("OTF_INTEGRATION_TFC_GCP_RUN_SERVICE_ACCOUNT_EMAIL needed for dynamic credentials integration test")
	}
	project, ok := os.LookupEnv("OTF_INTEGRATION_GCP_PROJECT")
	if !ok {
		t.Skip("OTF_INTEGRATION_GCP_PROJECT needed for dynamic credentials integration test")
	}
	// OTF_INTEGRATION_HOSTNAME should be set to something other than
	// "localhost" because GCP doesn't like localhost as an issuer
	issuer, ok := os.LookupEnv("OTF_INTEGRATION_HOSTNAME")
	if !ok {
		t.Skip("OTF_INTEGRATION_HOSTNAME needed for dynamic credentials integration test")
	}
	orgName, ok := os.LookupEnv("OTF_INTEGRATION_DYNAMIC_CREDENTIALS_ORGANIZATION")
	if !ok {
		t.Skip("OTF_INTEGRATION_DYNAMIC_CREDENTIALS_ORGANIZATION needed for dynamic credentials integration test")
	}

	daemon, _, ctx := setup(t,
		withKeyPairPaths(privateKeyPath, publicKeyPath),
		withHostname(issuer),
		skipDefaultOrganization(),
	)

	// check endpoints are exposed
	configResp := daemon.getLocalURL(t, "/.well-known/openid-configuration")
	assert.Equal(t, 200, configResp.StatusCode)

	jwksResp := daemon.getLocalURL(t, "/.well-known/jwks")
	assert.Equal(t, 200, jwksResp.StatusCode)

	// create an organization with a specific name that matches the assertion
	// condition in GCP, e.g. `attribute.terraform_organization_name="acme"`
	org, err := daemon.Organizations.Create(ctx, organization.CreateOptions{Name: new(orgName)})
	require.NoError(t, err)

	ws1 := daemon.createWorkspace(t, ctx, org)
	_ = daemon.createVariable(t, ctx, ws1, &variable.CreateVariableOptions{
		Key:      new("TFC_GCP_PROVIDER_AUTH"),
		Value:    new("true"),
		Category: internal.Ptr(variable.CategoryEnv),
	})
	_ = daemon.createVariable(t, ctx, ws1, &variable.CreateVariableOptions{
		Key:      new("TFC_GCP_RUN_SERVICE_ACCOUNT_EMAIL"),
		Value:    &serviceAccount,
		Category: internal.Ptr(variable.CategoryEnv),
	})
	_ = daemon.createVariable(t, ctx, ws1, &variable.CreateVariableOptions{
		Key:      new("TFC_GCP_WORKLOAD_PROVIDER_NAME"),
		Value:    &workload,
		Category: internal.Ptr(variable.CategoryEnv),
	})

	config := fmt.Sprintf(`
# this is the minimum configuration that leverages GCP permissions without
# spending any money.
provider "google" {
  project = "%s"
}

data "google_project" "my-project" {}
`, project)

	// create tarball of root module and upload
	root := t.TempDir()
	err = os.WriteFile(filepath.Join(root, "main.tf"), []byte(config), 0o777)
	require.NoError(t, err)
	tarball, err := internal.Pack(root)
	require.NoError(t, err)
	cv1 := daemon.createConfigurationVersion(t, ctx, ws1, &configversion.CreateOptions{})
	err = daemon.Configs.UploadConfig(ctx, cv1.ID, tarball)
	require.NoError(t, err)

	run := daemon.createRun(t, ctx, ws1, cv1, nil)
	daemon.waitRunStatus(t, ctx, run.ID, runstatus.PlannedAndFinished)

	// Now check dynamic creds work on an agent.
	t.Run("with agent", func(t *testing.T) {
		pool1, err := daemon.Runners.CreateAgentPool(ctx, runner.CreateAgentPoolOptions{
			Name:         "pool-1",
			Organization: org.Name,
		})
		require.NoError(t, err)

		_, err = daemon.Workspaces.Update(ctx, ws1.ID, workspace.UpdateOptions{
			ExecutionMode: internal.Ptr(workspace.AgentExecutionMode),
			AgentPoolID:   &pool1.ID,
		})
		require.NoError(t, err)

		_, shutdown := daemon.startAgent(
			t,
			ctx,
			org.Name,
			&pool1.ID,
			"",
		)
		defer shutdown()

		run := daemon.createRun(t, ctx, ws1, cv1, nil)
		daemon.waitRunStatus(t, ctx, run.ID, runstatus.PlannedAndFinished)
	})
}
