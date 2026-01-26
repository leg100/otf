package integration

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/testbrowser"
	"github.com/leg100/otf/internal/testcompose"
	"github.com/leg100/otf/internal/user"
	"github.com/playwright-community/playwright-go"
)

const assertionTimeout = 10_000

var (
	// a shared secret which signs the shared user session
	sharedSecret []byte

	// shared environment variables for individual tests to use
	sharedEnvs []string

	// Context conferring site admin privileges
	adminCtx = authz.AddSubjectToContext(context.Background(), &user.SiteAdmin)

	// pool of web browsers
	browser *testbrowser.Pool

	// Setup playwright browser expectations with a timeout to wait for expected
	// condition.
	expect = playwright.NewPlaywrightAssertions(assertionTimeout)

	// Path to engine binaries
	terraformPath, tofuPath string
)

func TestMain(m *testing.M) {
	// must parse flags before calling testing.Short()
	flag.Parse()
	if testing.Short() {
		return
	}
	code, err := doMain(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup integration tests: %s\n", err.Error())
		os.Exit(1)
	}
	os.Exit(code)
}

func doMain(m *testing.M) (int, error) {
	// Configure ^C to terminate program
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-ctx.Done()
		// Stop handling ^C; another ^C will exit the program.
		cancel()
	}()

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
	//
	// NOTE: gcp pub sub emulator only runs on amd64
	if runtime.GOARCH == "amd64" {
		pubsub, err := testcompose.GetHost(testcompose.PubSub)
		if err != nil {
			return 0, fmt.Errorf("getting pub sub emulator host: %w", err)
		}
		unset, err = setenv("PUBSUB_EMULATOR_HOST", pubsub)
		if err != nil {
			return 0, err
		}
		defer unset()
	}

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

	oldHome := os.Getenv("HOME")
	unset, err = setenv("HOME", homeDir)
	if err != nil {
		return 0, err
	}
	defer unset()

	// Give assertions more time on Github Actions, because it can be awfully
	// slow.
	expect = playwright.NewPlaywrightAssertions(assertionTimeout * 2)

	// Playwright installs its drivers and browsers in
	// $HOME/.cache/ms-playwright, but we've set a new $HOME, so set the
	// relevant environment variable pointing at that directory in the original
	// home.
	err = os.Symlink(path.Join(oldHome, ".cache"), path.Join(os.Getenv("HOME"), ".cache"))
	if err != nil {
		return 0, err
	}

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

	// Make filesystem mirror available to tests. This is a local cache of
	// providers that speeds up tests. It shouldn't be necessary because the squid
	// proxy caches providers on first download, but in the case of OpenTofu,
	// providers are hosted on github, which sends back 302s and unique URLs
	// which are uncacheable.
	//
	// TODO: is squid needed anymore?
	{
		const mirrorPath = "../../mirror"
		if _, err := os.Stat(mirrorPath); err != nil {
			return 0, fmt.Errorf("integration tests require mirror to be setup with ./hacks/setup_mirror.sh: %w", err)
		}
		mirrorPathAbs, err := filepath.Abs(mirrorPath)
		if err != nil {
			return 0, fmt.Errorf("getting absolute path to provider mirror: %w", err)
		}
		err = os.Symlink(mirrorPathAbs, path.Join(os.Getenv("HOME"), ".terraform.d", "plugins"))
		if err != nil {
			return 0, fmt.Errorf("symlinking provider mirror: %w", err)
		}
	}

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
		return 0, fmt.Errorf("creating browser pool: %w", err)
	}
	defer cleanup()
	browser = pool

	// Download engines now rather than in individual tests because it would
	// otherwise make the latter flaky.
	{
		downloader, err := engine.NewDownloader(logr.Discard(), engine.Default, "")
		if err != nil {
			return 0, fmt.Errorf("creating downloader: %w", err)
		}
		terraformPath, err = downloader.Download(ctx, engine.Default.DefaultVersion(), os.Stdout)
		if err != nil {
			return 0, fmt.Errorf("downloading terraform: %w", err)
		}
	}
	{
		downloader, err := engine.NewDownloader(logr.Discard(), engine.Tofu, "")
		if err != nil {
			return 0, fmt.Errorf("creating downloader: %w", err)
		}
		tofuPath, err = downloader.Download(ctx, engine.Tofu.DefaultVersion(), os.Stdout)
		if err != nil {
			return 0, fmt.Errorf("downloading tofu: %w", err)
		}
	}

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
