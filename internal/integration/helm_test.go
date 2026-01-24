package integration

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/goccy/go-yaml"
	tfe "github.com/hashicorp/go-tfe"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/testutils"
	"github.com/mitchellh/iochan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8swait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func TestHelm(t *testing.T) {
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

	// Build and load all images into kind
	cmd = exec.CommandContext(t.Context(), "make", "load-all")
	cmd.Dir = "../.."
	out, err = cmd.CombinedOutput()
	loadOutput := string(out)
	require.NoError(t, err, loadOutput)
	// otfd image load logs something like:
	// Image: "leg100/otfd:v0.5.0-77-g2d796a09f-dirty" with ID "sha256:a74a9ffe92d774bc7dba20d1f70d232f22da2b7a7123a6a3494377d

	// Extract image tag loaded into k8s
	imageRe := regexp.MustCompile(`"leg100/otfd:([^ ]+)"`)
	matches := imageRe.FindStringSubmatch(loadOutput)
	require.NotNil(t, matches, string(out))
	imageTag := matches[1]

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

	ns := petname.Generate(2, "-")
	release := "otfd"
	testValuesPath := "../../charts/otfd/test-values.yaml"

	// Install helm charts
	cmd = exec.CommandContext(t.Context(),
		"helm",
		"install",
		release,
		"../../charts/otfd",
		"--kubeconfig", kubeconfigPath,
		"--create-namespace",
		"--namespace", ns,
		"--values", testValuesPath,
		"--set", "image.tag="+imageTag,
		"--set", "logging.http=true",
		"--set", "logging.verbosity=9",
		"--set", "runner.executor=kubernetes",
		"--set", "runner.kubernetesTTLAfterFinish=1s",
		"--wait",
	)
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	// Delete namespace at finish
	t.Cleanup(func() {
		err := clientset.CoreV1().Namespaces().Delete(context.Background(), ns, metav1.DeleteOptions{})
		require.NoError(t, err)
	})

	// Create tunnel to otfd service so that we can communicate with it.
	r, w := io.Pipe()
	go func() {
		cmd = exec.CommandContext(t.Context(),
			"kubectl",
			"-n", ns,
			"--kubeconfig", kubeconfigPath,
			"port-forward",
			"services/otfd",
			":80",
		)
		cmd.Dir = "../.."
		cmd.Stderr = &buf
		cmd.Stdout = w
		err := cmd.Run()
		require.NoError(t, err, buf.String())
	}()
	// Outputs something like:
	//Forwarding from 127.0.0.1:40009 -> 8080
	//Forwarding from [::1]:40009 -> 8080
	//
	// Grab random local listening port
	localPortRe := regexp.MustCompile(`Forwarding from 127.0.0.1:(\d+) -> 8080`)
	matches = localPortRe.FindStringSubmatch(<-iochan.DelimReader(r, '\n'))
	require.NotNil(t, matches, string(out))
	localPort := matches[1]

	// Extract site token from helm test values file
	rawTestValues := testutils.ReadFile(t, testValuesPath)
	var testValues struct {
		SiteToken string `json:"siteToken"`
	}
	err = yaml.Unmarshal(rawTestValues, &testValues)
	require.NoError(t, err)

	// Create TFE Client to talk to the remote otfd. We would use the OTF client
	// but not all endpoints are implemented.
	client, err := tfe.NewClient(&tfe.Config{
		Address: fmt.Sprintf("http://localhost:%s", localPort),
		Token:   testValues.SiteToken,
	})
	require.NoError(t, err)

	org, err := client.Organizations.Create(t.Context(), tfe.OrganizationCreateOptions{
		Name:  internal.Ptr("acme"),
		Email: internal.Ptr("bollocks@morebollocks.bollocks"),
	})
	require.NoError(t, err)

	ws, err := client.Workspaces.Create(t.Context(), org.Name, tfe.WorkspaceCreateOptions{
		Name: internal.Ptr("dev"),
	})
	require.NoError(t, err)

	cv, err := client.ConfigurationVersions.Create(t.Context(), ws.ID, tfe.ConfigurationVersionCreateOptions{})
	require.NoError(t, err)

	tarball, err := os.Open("./testdata/root.tar.gz")
	require.NoError(t, err)
	err = client.ConfigurationVersions.UploadTarGzip(t.Context(), cv.UploadURL, tarball)
	require.NoError(t, err)

	run, err := client.Runs.Create(t.Context(), tfe.RunCreateOptions{
		Workspace:            ws,
		ConfigurationVersion: cv,
	})
	require.NoError(t, err)

	// Pod should succeed and run should reach planned status
	opts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("otf.ninja/workspace-id=%s", ws.ID),
	}
	k8swait.PollUntilContextTimeout(t.Context(), time.Second, time.Second*30, true, func(ctx context.Context) (bool, error) {
		pods, err := clientset.CoreV1().Pods(ns).List(t.Context(), opts)
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

	// Ensure k8s garbage collection works as configured with both job and
	// secret resources deleted.
	err = k8swait.PollUntilContextTimeout(t.Context(), time.Second, time.Second*30, true, func(ctx context.Context) (bool, error) {
		secrets, err := clientset.CoreV1().Secrets(ns).List(t.Context(), opts)
		if err != nil {
			return false, err
		}
		return len(secrets.Items) == 0, nil
	})
	require.NoError(t, err)
	err = k8swait.PollUntilContextTimeout(t.Context(), time.Second, time.Second*30, true, func(ctx context.Context) (bool, error) {
		jobs, err := clientset.BatchV1().Jobs(ns).List(t.Context(), opts)
		if err != nil {
			return false, err
		}
		return len(jobs.Items) == 0, nil
	})
	require.NoError(t, err)

	run, err = client.Runs.Read(t.Context(), run.ID)
	require.NoError(t, err)
	assert.Equal(t, runstatus.Planned, runstatus.Status(run.Status))
}
