package integration

import (
	"os"
	"os/exec"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const enableKubeTestsEnvVar = "OTF_TEST_KUBE"

func TestHelm(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// This test is deliberately not run in parallel with other tests. The test
	// can trigger a ERR_NETWORK_CHANGED error in chrome, disrupting the other
	// tests.

	if _, ok := os.LookupEnv(enableKubeTestsEnvVar); !ok {
		t.Skip(enableKubeTestsEnvVar, "not set")
	}

	if _, err := exec.LookPath("kind"); err != nil {
		t.Skip("kubernetes kind not installed")
	}

	kdeploy, debug, err := NewKubeDeploy(t.Context(), KubeDeployConfig{
		RepoDir: "../../",
		// Delete job and its secret 1 second after job finishes.
		JobTTL: 1,
	})
	require.NoError(t, err, debug(t.Context()))

	t.Cleanup(func() {
		// Don't delete namespace if test failed, to allow debugging.
		kdeploy.Close(!t.Failed())
	})

	org, err := kdeploy.Organizations.Create(t.Context(), tfe.OrganizationCreateOptions{
		Name:  new("acme"),
		Email: new("bollocks@morebollocks.bollocks"),
	})
	require.NoError(t, err)

	t.Run("create run", func(t *testing.T) {
		t.Parallel()

		ws, err := kdeploy.Workspaces.Create(t.Context(), org.Name, tfe.WorkspaceCreateOptions{
			Name: new("dev"),
		})
		require.NoError(t, err)

		cv, err := kdeploy.ConfigurationVersions.Create(t.Context(), ws.ID, tfe.ConfigurationVersionCreateOptions{})
		require.NoError(t, err)

		tarball, err := os.Open("./testdata/root.tar.gz")
		require.NoError(t, err)
		err = kdeploy.ConfigurationVersions.UploadTarGzip(t.Context(), cv.UploadURL, tarball)
		require.NoError(t, err)

		run, err := kdeploy.Runs.Create(t.Context(), tfe.RunCreateOptions{
			Workspace:            ws,
			ConfigurationVersion: cv,
		})
		require.NoError(t, err)

		// Pod should succeed and run should reach planned status
		debug, err = kdeploy.WaitPodSucceed(t.Context(), run.ID, time.Minute)
		require.NoError(t, err, debug(t.Context()))

		// Ensure k8s garbage collection works as configured with both job and
		// secret resources deleted.
		err = kdeploy.WaitJobAndSecretDeleted(t.Context(), run.ID)
		require.NoError(t, err)

		run, err = kdeploy.Runs.Read(t.Context(), run.ID)
		require.NoError(t, err)
		assert.Equal(t, runstatus.Planned, runstatus.Status(run.Status), debug(t.Context()))
	})

	t.Run("deploy agent", func(t *testing.T) {
		t.Parallel()

		// Create agent pool and agent token
		pool, err := kdeploy.AgentPools.Create(t.Context(), org.Name, tfe.AgentPoolCreateOptions{
			Name:               new("test-pool"),
			OrganizationScoped: new(true),
		})
		require.NoError(t, err)

		token, err := kdeploy.AgentTokens.Create(t.Context(), pool.ID, tfe.AgentTokenCreateOptions{
			Description: new("my fancy token"),
		})
		require.NoError(t, err)

		debug, err = kdeploy.InstallAgentChart(t.Context(), token.Token)
		require.NoError(t, err, debug(t.Context()))

		ws, err := kdeploy.Workspaces.Create(t.Context(), org.Name, tfe.WorkspaceCreateOptions{
			Name:          new("dev-agent"),
			ExecutionMode: new("agent"),
			AgentPoolID:   new(pool.ID),
		})
		require.NoError(t, err)

		cv, err := kdeploy.ConfigurationVersions.Create(t.Context(), ws.ID, tfe.ConfigurationVersionCreateOptions{})
		require.NoError(t, err)

		tarball, err := os.Open("./testdata/root.tar.gz")
		require.NoError(t, err)
		err = kdeploy.ConfigurationVersions.UploadTarGzip(t.Context(), cv.UploadURL, tarball)
		require.NoError(t, err)

		run, err := kdeploy.Runs.Create(t.Context(), tfe.RunCreateOptions{
			Workspace:            ws,
			ConfigurationVersion: cv,
		})
		require.NoError(t, err)

		// Pod should succeed and run should reach planned status
		debug, err = kdeploy.WaitPodSucceed(t.Context(), run.ID, time.Minute)
		require.NoError(t, err, debug(t.Context()))

		// Ensure k8s garbage collection works as configured with both job and
		// secret resources deleted.
		err = kdeploy.WaitJobAndSecretDeleted(t.Context(), run.ID)
		require.NoError(t, err)
	})
}
