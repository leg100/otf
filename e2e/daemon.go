package e2e

import (
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"testing"
	"time"

	"github.com/leg100/otf"
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
}

func (d *daemon) withFlags(flags ...string) {
	d.flags = append(d.flags, flags...)
}

func (d *daemon) withGithubUser(user *otf.User) {
	d.enableGithub = true
	d.githubOptions = append(d.githubOptions, github.WithUser(user))
}

func (d *daemon) withGithubRepo(repo *otf.Repo) {
	d.enableGithub = true
	d.githubOptions = append(d.githubOptions, github.WithRepo(repo))
}

func (d *daemon) withGithubTarball(tarball []byte) {
	d.enableGithub = true
	d.githubOptions = append(d.githubOptions, github.WithArchive(tarball))
}

// start an instance of the otfd daemon and return its hostname.
func (d *daemon) start(t *testing.T) string {
	database, ok := os.LookupEnv("OTF_TEST_DATABASE_URL")
	require.True(t, ok, "OTF_TEST_DATABASE_URL not set")

	flags := append(d.flags,
		"--address", ":0",
		"--ssl", "true",
		"--secret", "fe56cd2eae641f73687349ee32af43048805a9624eb3fcd0bdaf5d5dc8ffd5bc",
		"--cert-file", "./fixtures/cert.crt",
		"--key-file", "./fixtures/key.pem",
		"--dev-mode=false",
		"--database", database,
	)

	if d.enableGithub {
		githubServer := github.NewTestServer(t, d.githubOptions...)
		githubURL, err := url.Parse(githubServer.URL)
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
