package integration

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/testbrowser"
	"github.com/leg100/otf/internal/testcompose"
)

var (
	// a shared secret which signs the shared user session
	sharedSecret []byte

	// shared environment variables for individual tests to use
	sharedEnvs []string

	// Context conferring site admin privileges
	adminCtx = internal.AddSubjectToContext(context.Background(), &auth.SiteAdmin)

	// pool of web browsers
	browser *testbrowser.Pool
)

func TestMain(m *testing.M) {
	code, err := doMain(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup integration tests: %s\n", err.Error())
		os.Exit(1)
	}
	os.Exit(code)
}

func doMain(m *testing.M) (int, error) {
	for _, chromium := range []string{"chromium", "chromium-browser"} {
		if _, err := exec.LookPath(chromium); err == nil {
			return 0, fmt.Errorf("found %s executable in path; chromium has a bug that breaks the browser-based tests; see https://github.com/chromedp/chromedp/issues/1325", chromium)
		}
	}

	// Start external services
	if err := testcompose.Up(); err != nil {
		return 0, fmt.Errorf("starting external services: %w", err)
	}

	// get postgres host and set environment variable
	host, err := testcompose.GetHost(testcompose.Postgres)
	if err != nil {
		return 0, fmt.Errorf("getting postgres host: %w", err)
	}
	postgresURL := fmt.Sprintf("postgres://postgres:postgres@%s/postgres", host)
	unset, err := setenv("OTF_TEST_DATABASE_URL", postgresURL)
	if err != nil {
		return 0, err
	}
	defer unset()

	// get squid host and set environment variable
	squid, err := testcompose.GetHost(testcompose.Squid)
	if err != nil {
		return 0, fmt.Errorf("getting squid host: %w", err)
	}
	unset, err = setenv("HTTPS_PROXY", squid)
	if err != nil {
		return 0, err
	}
	defer unset()

	// get pubsub emulator host and set environment variable
	pubsub, err := testcompose.GetHost(testcompose.PubSub)
	if err != nil {
		return 0, fmt.Errorf("getting squid host: %w", err)
	}
	unset, err = setenv("PUBSUB_EMULATOR_HOST", pubsub)
	if err != nil {
		return 0, err
	}
	defer unset()

	// The otfd daemon spawned in an integration test uses a self-signed cert.
	// The following environment variable instructs any Go program spawned in a
	// test, e.g. the terraform CLI, the otf agent, etc, to trust the
	// self-signed cert.
	// * Assign the *absolute* path to the SSL cert because Go program's working
	// directory may differ from the integration test directory.
	wd, err := os.Getwd()
	if err != nil {
		return 0, fmt.Errorf("retrieving working directory: %w", err)
	}
	unset, err = setenv("SSL_CERT_FILE", filepath.Join(wd, "./fixtures/cert.pem"))
	if err != nil {
		return 0, err
	}
	defer unset()

	// Create dedicated home directory for duration of integration tests.
	// Terraform CLI and the `otf` CLI create various directories and dot files
	// in the home directory and we do not want to pollute the user's home
	// directory.
	homeDir, err := os.MkdirTemp("", "")
	if err != nil {
		return 0, fmt.Errorf("making dedicated home directory: %w", err)
	}
	defer os.RemoveAll(homeDir)
	unset, err = setenv("HOME", homeDir)
	if err != nil {
		return 0, err
	}
	defer unset()

	// Instruct terraform CLI to skip checks for new versions.
	unset, err = setenv("CHECKPOINT_DISABLE", "true")
	if err != nil {
		return 0, err
	}
	defer unset()

	// Ensure ~/.terraform.d exists - 'terraform login' has a bug whereby it tries to
	// persist the API token it receives to a temporary file in ~/.terraform.d but
	// fails if ~/.terraform.d doesn't exist yet. This only happens when
	// CHECKPOINT_DISABLE is set, because the checkpoint would otherwise handle
	// creating that directory first.
	os.MkdirAll(path.Join(os.Getenv("HOME"), ".terraform.d"), 0o755)

	// Create a secret with which to (1) create user session tokens and (2)
	// for assignment to daemons so that the token passes verification
	sharedSecret = make([]byte, 16)
	_, err = rand.Read(sharedSecret)
	if err != nil {
		return 0, err
	}

	// Setup pool of browsers
	pool, cleanup, err := testbrowser.NewPool(sharedSecret)
	if err != nil {
		return 0, err
	}
	defer cleanup()
	browser = pool

	return m.Run(), nil
}

// setenv sets an environment variable and returns a func to unset the variable.
// The environment variable is added to a shared slice, envs, for individual
// tests to use.
func setenv(name, value string) (func(), error) {
	err := os.Setenv(name, value)
	if err != nil {
		return nil, fmt.Errorf("setting %s=%s: %w", name, value, err)
	}
	sharedEnvs = append(sharedEnvs, name+"="+value)
	return func() {
		os.Unsetenv(name)
	}, nil
}
