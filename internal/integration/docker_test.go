package integration

import (
	"os"
	"os/exec"
	"testing"

	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
)

func TestDocker(t *testing.T) {
	integrationTest(t)

	var (
		kubeConfigPath   string
		k3sContainer     *k3s.K3sContainer
		imageBuildOutput []byte
		err              error
	)

	// Standup k8s cluster and build otf-job image concurrently - they can both
	// take a little while (> 10 seconds).
	{
		g, ctx := errgroup.WithContext(t.Context())
		g.Go(func() error {
			k3sContainer, err = k3s.Run(ctx, "rancher/k3s:v1.35.0-k3s1")
			testcontainers.CleanupContainer(t, k3sContainer)
			t.Log("finished standing up k8s")
			return err
		})
		g.Go(func() error {
			cmd := exec.CommandContext(ctx, "make", "image-job")
			cmd.Dir = "../.."
			//out, err := cmd.CombinedOutput()
			t.Log("finished making image")
			//imageBuildOutput = out
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		})

		require.NoError(t, g.Wait(), string(imageBuildOutput))
	}

	kubeConfigYaml, err := k3sContainer.GetKubeConfig(t.Context())
	require.NoError(t, err)

	restcfg, err := clientcmd.RESTConfigFromKubeConfig(kubeConfigYaml)
	require.NoError(t, err)

	kubeConfigPath = testutils.TempFile(t, kubeConfigYaml)

	_, err = kubernetes.NewForConfig(restcfg)
	require.NoError(t, err)

	t.Log("loading image")
	err = k3sContainer.LoadImages(t.Context(), "leg100/otf-job:latest")
	require.NoError(t, err)

	t.Log("starting daemon")
	daemon, org, ctx := setup(t, withKubernetesExecutor(kubeConfigPath))
	ws := daemon.createWorkspace(t, ctx, org)

	t.Log("upload config")

	// upload config
	cv := daemon.createAndUploadConfigurationVersion(t, ctx, ws, nil)

	t.Log("create run")
	// create run
	run := daemon.createRun(t, ctx, ws, cv, nil)

	t.Log("wait")

	// let it run to completion
	daemon.waitRunStatus(t, ctx, run.ID, runstatus.Planned)
}
