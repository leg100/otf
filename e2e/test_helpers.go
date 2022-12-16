package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	expect "github.com/google/goexpect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func startAgent(t *testing.T, token, address string) {
	out, err := os.CreateTemp(t.TempDir(), "agent.out")
	require.NoError(t, err)

	e, res, err := expect.Spawn(
		fmt.Sprintf("%s --token %s --address %s", "otf-agent", token, address),
		time.Minute,
		expect.PartialMatch(true),
		expect.Verbose(testing.Verbose()),
		expect.Tee(out),
	)
	require.NoError(t, err)

	_, err = e.ExpectBatch([]expect.Batcher{
		&expect.BExp{R: "successfully authenticated"},
		&expect.BExp{R: "stream update.*successfully connected"},
	}, time.Second*10)
	require.NoError(t, err)

	// terminate at end of parent test
	t.Cleanup(func() {
		e.SendSignal(os.Interrupt)
		if !assert.NoError(t, <-res) || t.Failed() {
			logs, err := os.ReadFile(out.Name())
			require.NoError(t, err)
			t.Log("--- agent logs ---")
			t.Log(string(logs))
		}
	})
}

// createAgentToken creates an agent token via the CLI
func createAgentToken(t *testing.T, organization, hostname string) string {
	cmd := exec.Command("otf", "agents", "tokens", "new", "testing", "--organization", organization, "--address", hostname)
	out, err := cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	re := regexp.MustCompile(`Successfully created agent token: (agent\.[a-zA-Z0-9\-_]+)`)
	matches := re.FindStringSubmatch(string(out))
	require.Equal(t, 2, len(matches))
	return matches[1]
}

// newRootModule creates a terraform root module, returning its directory path
func newRootModule(t *testing.T, hostname, organization, workspace string) string {
	config := []byte(fmt.Sprintf(`
terraform {
  backend "remote" {
	hostname = "%s"
	organization = "%s"

	workspaces {
	  name = "%s"
	}
  }
}
resource "null_resource" "e2e" {}
`, hostname, organization, workspace))

	root := t.TempDir()
	err := os.WriteFile(filepath.Join(root, "main.tf"), config, 0o600)
	require.NoError(t, err)

	return root
}

func addBuildsToPath(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	t.Setenv("PATH", path.Join(wd, "../_build")+":"+os.Getenv("PATH"))
}

// matchText is a custom chromedp Task that extracts text content using the
// selector and asserts that it matches the wanted string.
func matchText(t *testing.T, selector, want string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		var got string
		err := chromedp.Text(selector, &got, chromedp.NodeVisible).Do(ctx)
		require.NoError(t, err)
		require.Equal(t, want, strings.TrimSpace(got))
		return nil
	}
}
