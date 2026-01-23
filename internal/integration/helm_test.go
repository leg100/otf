package integration

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/goccy/go-yaml"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/client"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/testutils"
	"github.com/mitchellh/iochan"
	"github.com/stretchr/testify/require"
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
	//restcfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	//require.NoError(t, err)
	//clientset, err := kubernetes.NewForConfig(restcfg)
	//require.NoError(t, err)

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
		"--wait",
	)
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	// Delete namespace at finish
	t.Cleanup(func() {
		//err := clientset.CoreV1().Namespaces().Delete(context.Background(), ns, metav1.DeleteOptions{})
		//require.NoError(t, err)
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

	client, err := client.New(otfhttp.ClientConfig{
		URL:   fmt.Sprintf("http://localhost:%s", localPort),
		Token: testValues.SiteToken,
	})

	org, err := client.Organizations.CreateOrganization(t.Context(), organization.CreateOptions{
		Name: internal.Ptr("acme"),
	})
	require.NoError(t, err)
	t.Logf("created organization: %s", org.Name)
}
