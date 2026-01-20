package integration

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/netip"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8swait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const kubeNamespace = "default"

func TestKubeExecutor(t *testing.T) {
	integrationTest(t)

	if _, err := exec.LookPath("kind"); err != nil {
		t.Skip("kubernetes kind not installed")
	}

	// Kind must have cluster running
	cmd := exec.CommandContext(t.Context(), "kind", "get", "clusters")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	if strings.Contains(string(out), "No kind clusters found") {
		t.Fatal("kind installed but no cluster found")
	}

	// Build leg100/otf-job image
	cmd = exec.CommandContext(t.Context(), "make", "image-job")
	cmd.Dir = "../.."
	out, err = cmd.CombinedOutput()
	imageBuildOutput := string(out)
	require.NoError(t, err, imageBuildOutput)

	// Extract image tag that was just built
	re := regexp.MustCompile(`leg100/otf-job:[^ ]+`)
	image := re.FindString(imageBuildOutput)
	require.NotEqual(t, "", image, imageBuildOutput)

	// Load image into kind
	cmd = exec.CommandContext(t.Context(), "kind", "load", "docker-image", image)
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	// Write kind's kubeconfig to a temp file
	cmd = exec.CommandContext(t.Context(), "kind", "get", "kubeconfig")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	err = cmd.Run()
	require.NoError(t, err)
	kubeconfigPath := testutils.TempFile(t, buf.Bytes())

	// Build k8s client
	restcfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	require.NoError(t, err)
	clientset, err := kubernetes.NewForConfig(restcfg)
	require.NoError(t, err)

	// Delete all jobs at test finish.
	t.Cleanup(func() {
		err := clientset.BatchV1().Jobs(kubeNamespace).DeleteCollection(
			context.Background(),
			metav1.DeleteOptions{
				// Delete pods too.
				PropagationPolicy: internal.Ptr(metav1.DeletePropagationBackground),
			},
			metav1.ListOptions{},
		)
		require.NoError(t, err)
	})

	// Because the daemon is running on the host, and the job is running in a
	// k3s cluster in a container, the job needs Job needs an endpoint with
	// which it can communicate with the daemon.
	//
	// The IP address of the host needs to be routable, so we get the IP address
	// that would be used for outbound traffic on the host, which is usually
	// routable from the other side of the docker bridge.
	//
	// Finding out the port of the daemon is trickier, because a random port is
	// assigned only once the server has started, and ordinarily we would have
	// to assign the port to the k8s executor *before* the server has started.
	// Hence we use an interface for the kubernetes job URL which permits the
	// test to use a callback, which is only called when the job is created.
	ip, err := internal.GetOutboundIP()
	require.NoError(t, err)
	serverURL := serverURLTestCallback{ip: ip}

	daemon, org, ctx := setup(t,
		// Delete job 1 second after it has finished.
		withKubernetesExecutor(kubeconfigPath, image, kubeNamespace, &serverURL, time.Second),
		withSSLDisabled(),
	)
	serverURL.port = daemon.ListenAddress.Port

	// create workspace, config, and run.
	ws := daemon.createWorkspace(t, ctx, org)
	cv := daemon.createAndUploadConfigurationVersion(t, ctx, ws, nil)
	run := daemon.createRun(t, ctx, ws, cv, nil)

	// Pod should succeed and run should reach planned status
	opts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("otf.ninja/workspace-id=%s", ws.ID.String()),
	}
	k8swait.PollUntilContextCancel(ctx, time.Second, true, func(ctx context.Context) (bool, error) {
		pods, err := clientset.CoreV1().Pods(kubeNamespace).List(t.Context(), opts)
		if err != nil {
			return false, err
		}
		if len(pods.Items) == 0 {
			return false, nil
		}
		switch pod := pods.Items[0]; pod.Status.Phase {
		case corev1.PodFailed:
			dumpPodLogs(t, &pod, clientset)
			t.FailNow()
		case corev1.PodSucceeded:
			return true, nil
		}
		return false, nil
	})
	run = daemon.getRun(t, ctx, run.ID)
	assert.Equal(t, runstatus.Planned, run.Status)

	// Ensure k8s garbage collection works as configured with both job and
	// secret resources deleted.
	k8swait.PollUntilContextCancel(ctx, time.Second, true, func(ctx context.Context) (bool, error) {
		secrets, err := clientset.CoreV1().Secrets(kubeNamespace).List(t.Context(), opts)
		if err != nil {
			return false, err
		}
		return len(secrets.Items) == 0, nil
	})
	k8swait.PollUntilContextCancel(ctx, time.Second, true, func(ctx context.Context) (bool, error) {
		jobs, err := clientset.BatchV1().Jobs(kubeNamespace).List(t.Context(), opts)
		if err != nil {
			return false, err
		}
		return len(jobs.Items) == 0, nil
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

type serverURLTestCallback struct {
	ip   netip.Addr
	port int
}

func (s *serverURLTestCallback) String() string {
	return (&url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(s.ip.String(), strconv.Itoa(s.port)),
	}).String()
}
