package e2e

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"github.com/mitchellh/iochan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var successfullyConnectedRegex = regexp.MustCompile("stream update.*successfully connected")

// setup dependencies for a test and return names for the org and workspace
func setup(t *testing.T) (org string, workspace string) {
	addBuildsToPath(t)

	// instruct terraform and otfd-agent to trust the self-signed cert
	wd, err := os.Getwd()
	require.NoError(t, err)
	t.Setenv("SSL_CERT_DIR", path.Join(wd, "./fixtures"))
	t.Logf("SSL_CERT_DIR=%s", os.Getenv("SSL_CERT_DIR"))

	// return unique name for org, and use the test name for the workspace name
	return uuid.NewString(), t.Name()
}

func startAgent(t *testing.T, token, address string, flags ...string) {
	args := []string{
		"--token", token,
		"--address", address,
	}
	args = append(args, flags...)

	cmd := exec.Command("otf-agent", args...)
	out, err := cmd.StdoutPipe()
	require.NoError(t, err)
	errout, err := cmd.StderrPipe()
	require.NoError(t, err)
	stdout := iochan.DelimReader(out, '\n')
	stderr := iochan.DelimReader(errout, '\n')
	// reset env
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "SSL_CERT_DIR=" + os.Getenv("SSL_CERT_DIR")}
	require.NoError(t, cmd.Start())

	// for capturing stdout
	loglines := []string{}

	t.Cleanup(func() {
		// kill otf-agent gracefully
		cmd.Process.Signal(os.Interrupt)
		assert.NoError(t, cmd.Wait())

		// upon failure dump stdout+stderr
		if t.Failed() {
			t.Log("test failed; here are the otf-agent logs:\n")
			for _, ll := range loglines {
				t.Logf(ll)
			}
		}
	})

	// wait for otf-agent to log that it has started successfully
	for {
		select {
		case <-time.After(time.Second * 5):
			t.Fatal("otf-agent failed to start correctly")
		case logline := <-stdout:
			loglines = append(loglines, logline)

			if successfullyConnectedRegex.MatchString(logline) {
				goto STARTED
			}
		case err := <-stderr:
			t.Fatalf(err)
		}
	}
STARTED:

	// capture remainder of stdout in background
	go func() {
		for logline := range stdout {
			loglines = append(loglines, logline)
		}
	}()
	// capture remainder of stderr in background
	go func() {
		for logline := range stderr {
			loglines = append(loglines, logline)
		}
	}()
}

// createAgentToken creates an agent token via the CLI
func createAgentToken(t *testing.T, organization, hostname, description string) string {
	cmd := exec.Command("otf", "agents", "tokens", "new", description, "--organization", organization, "--address", hostname)
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

func sendGithubPushEvent(t *testing.T, payload []byte, url, secret string) {
	// generate signature for push event
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	sig := mac.Sum(nil)

	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	require.NoError(t, err)
	req.Header.Add("Content-type", "application/json")
	req.Header.Add("X-GitHub-Event", "push")
	req.Header.Add("X-Hub-Signature-256", "sha256="+hex.EncodeToString(sig))

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	if !assert.Equal(t, http.StatusAccepted, res.StatusCode) {
		response, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		t.Fatal(string(response))
	}
}

// okDialog - Click OK on any browser javascript dialog boxes that pop up
func okDialog(t *testing.T, ctx context.Context) {
	chromedp.ListenTarget(ctx, func(ev any) {
		switch ev.(type) {
		case *page.EventJavascriptDialogOpening:
			go func() {
				err := chromedp.Run(ctx, page.HandleJavaScriptDialog(true))
				require.NoError(t, err)
			}()
		}
	})
}
