package e2e

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	expect "github.com/google/goexpect"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
	"github.com/mitchellh/iochan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var startedServerRegex = regexp.MustCompile(`started server address=.*:(\d+) ssl=true`)

// startDaemon starts an instance of the otfd daemon along with a github stub
// seeded with the given user. The hostname of the otfd daemon is returned.
func startDaemon(t *testing.T, user *otf.User) string {
	githubServer := html.NewTestGithubServer(t, user)
	githubURL, err := url.Parse(githubServer.URL)
	require.NoError(t, err)

	cmd := exec.Command("otfd",
		"--address", ":0",
		"--cert-file", "./fixtures/cert.crt",
		"--key-file", "./fixtures/key.pem",
		"--dev-mode=false",
		"--github-client-id", "stub-client-id",
		"--github-client-secret", "stub-client-secret",
		"--github-skip-tls-verification",
		"--github-hostname", githubURL.Host,
	)
	out, err := cmd.StdoutPipe()
	require.NoError(t, err)
	errout, err := cmd.StderrPipe()
	require.NoError(t, err)
	stdout := iochan.DelimReader(out, '\n')
	stderr := iochan.DelimReader(errout, '\n')

	require.NoError(t, cmd.Start())

	// record daemon's URL
	var url string

	// for capturing stdout
	loglines := []string{}

	t.Cleanup(func() {
		// kill otfd gracefully
		cmd.Process.Signal(os.Interrupt)
		assert.NoError(t, cmd.Wait())

		// upon failure dump stdout
		if t.Failed() {
			for _, ll := range loglines {
				t.Log(ll)
			}
		}
	})

	// wait for otfd to log that it has started successfully
	for {
		select {
		case <-time.After(time.Second * 5):
			t.Fatal("otfd failed to start correctly")
		case logline := <-stdout:
			loglines = append(loglines, logline)

			matches := startedServerRegex.FindStringSubmatch(logline)
			switch len(matches) {
			case 2:
				port := matches[1]
				url = "localhost:" + port
				goto STARTED
			case 0:
				// keep waiting
				continue
			default:
				t.Fatalf("server returned malformed output: %s", logline)
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

	return url
}

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

func createOrganization(t *testing.T) string {
	organization := uuid.NewString()
	cmd := exec.Command("otf", "organizations", "new", organization)
	out, err := cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	return organization
}

// login invokes 'terraform login <hostname>', configuring credentials for the
// given hostname with the given token.
func login(t *testing.T, hostname, token string) {
	tfpath, err := exec.LookPath("terraform")
	require.NoErrorf(t, err, "terraform executable not found in path")

	// nullifying PATH temporarily to make `terraform login` skip opening a browser
	// window
	path := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", path)

	e, tferr, err := expect.SpawnWithArgs(
		[]string{tfpath, "login", hostname},
		time.Minute,
		expect.PartialMatch(true),
		expect.Verbose(testing.Verbose()))
	require.NoError(t, err)
	defer e.Close()

	e.ExpectBatch([]expect.Batcher{
		&expect.BExp{R: "Enter a value:"}, &expect.BSnd{S: "yes\n"},
		&expect.BExp{R: "Enter a value:"}, &expect.BSnd{S: token + "\n"},
		&expect.BExp{R: "Success! Logged in to Terraform Enterprise"},
	}, time.Minute)
	require.NoError(t, <-tferr)
}

func addBuildsToPath(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	t.Setenv("PATH", path.Join(wd, "../_build")+":"+os.Getenv("PATH"))
}

func newBrowserAllocater(t *testing.T) context.Context {
	headless := true
	if v, ok := os.LookupEnv("OTF_E2E_HEADLESS"); ok {
		var err error
		headless, err = strconv.ParseBool(v)
		require.NoError(t, err)
	}

	ctx, cancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", headless),
			chromedp.Flag("hide-scrollbars", true),
			chromedp.Flag("mute-audio", true),
			chromedp.Flag("ignore-certificate-errors", true),
			chromedp.Flag("disable-gpu", true),
		)...)
	t.Cleanup(cancel)

	return ctx
}

// createAPIToken creates an API token via the web app
func createAPIToken(t *testing.T, hostname string) string {
	allocater := newBrowserAllocater(t)

	ctx, cancel := chromedp.NewContext(allocater)
	defer cancel()

	var token string

	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate("https://" + hostname),
		chromedp.Click(".login-button-github", chromedp.NodeVisible),
		chromedp.WaitReady(`body`),
		chromedp.Click("#top-right-profile-link > a", chromedp.NodeVisible),
		chromedp.WaitReady(`body`),
		chromedp.Click("#user-tokens-link > a", chromedp.NodeVisible),
		chromedp.WaitReady(`body`),
		chromedp.Click("#new-user-token-button", chromedp.NodeVisible),
		chromedp.WaitReady(`body`),
		chromedp.Focus("#description", chromedp.NodeVisible),
		input.InsertText("e2e-test"),
		chromedp.Submit("#description"),
		chromedp.WaitReady(`body`),
		chromedp.Text(".flash-success > .data", &token, chromedp.NodeVisible),
	})
	require.NoError(t, err)

	assert.Regexp(t, `user\.(.+)`, token)

	return token
}
