package integration

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func TestHelm(t *testing.T) {
	integrationTest(t)

	if _, err := exec.LookPath("kind"); err != nil {
		t.Skip("kubernetes kind not installed")
	}

	deploy, err := NewKubeDeploy(t.Context(), "", "otfd", "../..")
	require.NoError(t, err)

	t.Cleanup(func() {
		// Don't delete namespace if test failed, to allow debugging.
		deploy.Close(!t.Failed())
	})

	org, err := deploy.Organizations.Create(t.Context(), tfe.OrganizationCreateOptions{
		Name:  internal.Ptr("acme"),
		Email: internal.Ptr("bollocks@morebollocks.bollocks"),
	})
	require.NoError(t, err)

	t.Run("create run", func(t *testing.T) {
		t.Parallel()

		ws, err := deploy.Workspaces.Create(t.Context(), org.Name, tfe.WorkspaceCreateOptions{
			Name: internal.Ptr("dev"),
		})
		require.NoError(t, err)

		cv, err := deploy.ConfigurationVersions.Create(t.Context(), ws.ID, tfe.ConfigurationVersionCreateOptions{})
		require.NoError(t, err)

		tarball, err := os.Open("./testdata/root.tar.gz")
		require.NoError(t, err)
		err = deploy.ConfigurationVersions.UploadTarGzip(t.Context(), cv.UploadURL, tarball)
		require.NoError(t, err)

		run, err := deploy.Runs.Create(t.Context(), tfe.RunCreateOptions{
			Workspace:            ws,
			ConfigurationVersion: cv,
		})
		require.NoError(t, err)

		// Pod should succeed and run should reach planned status
		err = deploy.WaitPodSucceed(t.Context(), run.ID)
		require.NoError(t, err)

		// Ensure k8s garbage collection works as configured with both job and
		// secret resources deleted.
		err = deploy.WaitJobAndSecretDeleted(t.Context(), run.ID)
		require.NoError(t, err)

		run, err = deploy.Runs.Read(t.Context(), run.ID)
		require.NoError(t, err)
		assert.Equal(t, runstatus.Planned, runstatus.Status(run.Status))
	})

	t.Run("deploy agent", func(t *testing.T) {
		t.Parallel()

		// Create agent pool and agent token
		pool, err := deploy.AgentPools.Create(t.Context(), org.Name, tfe.AgentPoolCreateOptions{
			Name:               internal.Ptr("test-pool"),
			OrganizationScoped: internal.Ptr(true),
		})
		require.NoError(t, err)

		token, err := deploy.AgentTokens.Create(t.Context(), pool.ID, tfe.AgentTokenCreateOptions{
			Description: internal.Ptr("my fancy token"),
		})
		require.NoError(t, err)

		err = deploy.InstallAgentChart(t.Context(), token.Token)
		require.NoError(t, err)

		ws, err := deploy.Workspaces.Create(t.Context(), org.Name, tfe.WorkspaceCreateOptions{
			Name:          internal.Ptr("dev-agent"),
			ExecutionMode: internal.Ptr("agent"),
			AgentPoolID:   internal.Ptr(pool.ID),
		})
		require.NoError(t, err)

		cv, err := deploy.ConfigurationVersions.Create(t.Context(), ws.ID, tfe.ConfigurationVersionCreateOptions{})
		require.NoError(t, err)

		tarball, err := os.Open("./testdata/root.tar.gz")
		require.NoError(t, err)
		err = deploy.ConfigurationVersions.UploadTarGzip(t.Context(), cv.UploadURL, tarball)
		require.NoError(t, err)

		run, err := deploy.Runs.Create(t.Context(), tfe.RunCreateOptions{
			Workspace:            ws,
			ConfigurationVersion: cv,
		})
		require.NoError(t, err)

		// Pod should succeed and run should reach planned status
		err = deploy.WaitPodSucceed(t.Context(), run.ID)
		require.NoError(t, err)
		// Ensure k8s garbage collection works as configured with both job and
		// secret resources deleted.
		err = deploy.WaitJobAndSecretDeleted(t.Context(), run.ID)
		require.NoError(t, err)
	})
}

func dumpPodLogs(t *testing.T, pod *corev1.Pod, clientset *kubernetes.Clientset) {
	t.Helper()

	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{})
	podLogs, err := req.Stream(t.Context())
	require.NoError(t, err)
	t.Cleanup(func() {
		podLogs.Close()
	})

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	require.NoError(t, err)

	t.Logf("---- pod logs -----\n%s", buf.String())
}
