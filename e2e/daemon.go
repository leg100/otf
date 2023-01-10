package e2e

import (
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	gogithub "github.com/google/go-github/v41/github"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/github"
	"github.com/mitchellh/iochan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var startedServerRegex = regexp.MustCompile(`started server \| address=.*:(\d+)`)

// daemon builds and starts a daemon
type daemon struct {
	flags         []string
	enableGithub  bool
	githubOptions []github.TestServerOption
	githubServer  *github.TestServer
}

func (d *daemon) withFlags(flags ...string) {
	d.flags = append(d.flags, flags...)
}

func (d *daemon) withGithubUser(user *cloud.User) {
	d.enableGithub = true
	d.githubOptions = append(d.githubOptions, github.WithUser(user))
}

func (d *daemon) withGithubRepo(repo cloud.Repo) {
	d.enableGithub = true
	d.githubOptions = append(d.githubOptions, github.WithRepo(repo))
}

func (d *daemon) withGithubRefs(refs ...string) {
	d.enableGithub = true
	d.githubOptions = append(d.githubOptions, github.WithRefs(refs...))
}

func (d *daemon) withGithubTarball(tarball []byte) {
	d.enableGithub = true
	d.githubOptions = append(d.githubOptions, github.WithArchive(tarball))
}

func (d *daemon) registerStatusCallback(callback func(*gogithub.StatusEvent)) {
	d.enableGithub = true
	d.githubOptions = append(d.githubOptions, github.WithStatusCallback(callback))
}

// start an instance of the otfd daemon and return its hostname.
func (d *daemon) start(t *testing.T) string {
	database, ok := os.LookupEnv("OTF_TEST_DATABASE_URL")
	require.True(t, ok, "OTF_TEST_DATABASE_URL not set")

	hostname, ok := os.LookupEnv("OTF_E2E_HOSTNAME")
	require.True(t, ok, "OTF_E2E_HOSTNAME not set")
	if !strings.Contains(hostname, ".") {
		t.Fatalf("hostname %s is missing a dot (.) - terraform mandates that a module registry hostname must include a dot", hostname)
	}

	flags := append(d.flags,
		"--address", hostname+":0", // listen on random, available port
		"--ssl", "true",
		"--secret", "fe56cd2eae641f73687349ee32af43048805a9624eb3fcd0bdaf5d5dc8ffd5bc",
		"--cert-file", "./fixtures/cert.crt",
		"--key-file", "./fixtures/key.pem",
		"--dev-mode=false",
		"--database", database,
	)

	if d.enableGithub {
		d.githubServer = github.NewTestServer(t, d.githubOptions...)
		githubURL, err := url.Parse(d.githubServer.URL)
		require.NoError(t, err)

		flags = append(flags,
			"--github-client-id", "stub-client-id",
			"--github-client-secret", "stub-client-secret",
			"--github-skip-tls-verification",
			"--github-hostname", githubURL.Host,
		)
	}

	cmd := exec.Command("otfd", flags...)
	out, err := cmd.StdoutPipe()
	require.NoError(t, err)
	errout, err := cmd.StderrPipe()
	require.NoError(t, err)
	stdout := iochan.DelimReader(out, '\n')
	stderr := iochan.DelimReader(errout, '\n')
	cmd.Env = []string{"PATH=" + os.Getenv("PATH")}

	require.NoError(t, cmd.Start())

	// record daemon's URL
	var url string

	// for capturing stdout
	loglines := []string{}

	t.Cleanup(func() {
		// kill otfd gracefully
		cmd.Process.Signal(os.Interrupt)
		assert.NoError(t, cmd.Wait())

		// upon failure dump stdout+stderr
		if t.Failed() {
			t.Log("test failed; here are the otfd logs:\n")
			for _, ll := range loglines {
				t.Logf(ll)
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
				url = hostname + ":" + port
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
