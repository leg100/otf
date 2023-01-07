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
	expect "github.com/google/goexpect"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
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

// newOrg generates an organization name with a prefix and a random suffix.
func newOrg(prefix string) string {
	return prefix + "-" + otf.GenerateRandomString(6)
}

func newUser(username, org string, teams ...string) cloud.User {
	user := cloud.User{
		Name:          username,
		Organizations: []string{org},
	}
	for _, team := range teams {
		user.Teams = append(user.Teams, cloud.Team{
			Name:         team,
			Organization: org,
		})
	}
	return user
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
