package integration

import (
	"bytes"
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

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func TestKubeExecutor(t *testing.T) {
	integrationTest(t)

	if _, err := exec.LookPath("kind"); err != nil {
		t.Skip("kubernetes kind not installed")
	}

	// Kind must have cluster running
	cmd := exec.CommandContext(t.Context(), "kind", "get", "cluster")
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
	re := regexp.MustCompile(`leg100/otf-job:v[^ ]+`)
	image := re.FindString(imageBuildOutput)
	require.NotEqual(t, "", image)

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
		withKubernetesExecutor(kubeconfigPath, image, &serverURL),
	)
	serverURL.port = daemon.ListenAddress.Port

	ws := daemon.createWorkspace(t, ctx, org)

	// upload config
	cv := daemon.createAndUploadConfigurationVersion(t, ctx, ws, nil)

	// let it run to completion
	//daemon.waitRunStatus(t, ctx, run.ID, runstatus.Planned)

	// create run
	_ = daemon.createRun(t, ctx, ws, cv, nil)

	waitPodSucceeded(t, clientset, ws.ID)

}

func waitPodSucceeded(t *testing.T, clientset *kubernetes.Clientset, workspaceID resource.TfeID) {
	t.Helper()

	watch, err := clientset.CoreV1().Pods("default").Watch(t.Context(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("workspace-id=%s", workspaceID.String()),
	})
	require.NoError(t, err)

	podList, err := clientset.CoreV1().Pods("default").List(t.Context(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("workspace-id=%s", workspaceID.String()),
	})
	require.NoError(t, err)

	if len(podList.Items) > 0 {
		switch pod := podList.Items[0]; pod.Status.Phase {
		case corev1.PodFailed:
			dumpPodLogs(t, &pod, clientset)
			t.FailNow()
		case corev1.PodSucceeded:
			return
		}
	}

	for event := range watch.ResultChan() {
		switch pod := event.Object.(*corev1.Pod); pod.Status.Phase {
		case corev1.PodFailed:
			dumpPodLogs(t, pod, clientset)
			t.FailNow()
		case corev1.PodSucceeded:
			return
		}
	}
}

func dumpPodLogs(t *testing.T, pod *corev1.Pod, clientset *kubernetes.Clientset) {
	t.Helper()

	req := clientset.CoreV1().Pods("default").GetLogs(pod.Name, &corev1.PodLogOptions{})
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
