package integration

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/chromedp/chromedp"
)

var (
	allocator context.Context
	browser   context.Context
)

func TestMain(m *testing.M) {
	// The otfd daemon spawned in an integration test uses a self-signed cert.
	// The following environment variable instructs any Go program spawned in a
	// test, e.g. the terraform CLI, the otf agent, etc, to trust the
	// self-signed cert.
	// * Assign the *absolute* path to the SSL cert because Go program's working
	// directory may differ from the integration test directory.
	wd, err := os.Getwd()
	panicIfError(err)
	unset := setenv("SSL_CERT_FILE", filepath.Join(wd, "./fixtures/cert.pem"))
	defer unset()

	// Create dedicated home directory for duration of integration tests.
	// Terraform CLI and the `otf` CLI create various directories and dot files
	// in the home directory and we do not want to pollute the user's home
	// directory.
	homeDir, err := os.MkdirTemp("", "")
	panicIfError(err)
	defer func() {
		os.RemoveAll(homeDir)
	}()
	unset = setenv("HOME", homeDir)
	defer unset()

	// If HTTPS_PROXY has been defined then add it to the authoritative list of
	// environment variables so that processes, particularly terraform, spawed
	// in tests use the proxy. This can be very useful for caching repeated
	// downloads of terraform providers during tests.
	if proxy, ok := os.LookupEnv("HTTPS_PROXY"); ok {
		envs = append(envs, "HTTPS_PROXY="+proxy)
	}

	// Instruct terraform CLI to skip checks for new versions.
	unset = setenv("CHECKPOINT_DISABLE", "true")
	defer unset()

	// Ensure ~/.terraform.d exists - 'terraform login' has a bug whereby it tries to
	// persist the API token it receives to a temporary file in ~/.terraform.d but
	// fails if ~/.terraform.d doesn't exist yet. This only happens when
	// CHECKPOINT_DISABLE is set, because the checkpoint would otherwise handle
	// creating that directory first.
	os.MkdirAll(path.Join(os.Getenv("HOME"), ".terraform.d"), 0o755)

	// Setup chromedp browser driver. Headless mode determines whether browser
	// window is displayed (false) or not (true).
	headless := true
	if v, ok := os.LookupEnv("OTF_E2E_HEADLESS"); ok {
		var err error
		headless, err = strconv.ParseBool(v)
		if err != nil {
			panic("cannot parse OTF_E2E_HEADLESS")
		}
	}

	// Must create an allocator before creating the browser
	var cancel context.CancelFunc
	allocator, cancel = chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", headless),
			chromedp.Flag("hide-scrollbars", true),
			chromedp.Flag("mute-audio", true),
			chromedp.Flag("ignore-certificate-errors", true),
			chromedp.Flag("disable-gpu", true),
		)...)
	defer cancel()

	// Create browser instance to be shared betwen tests. Each test creates
	// isolated tabs within the browser.
	browser, cancel = chromedp.NewContext(allocator)
	defer cancel()
	_ = chromedp.Run(browser)

	os.Exit(m.Run())
}

func panicIfError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

// setenv sets an environment variable and returns a func to unset the variable.
// The environment variable is added to a shared slice, envs, for individual
// tests to use.
func setenv(name, value string) func() {
	err := os.Setenv(name, value)
	if err != nil {
		panic(err.Error())
	}
	envs = append(envs, name+"="+value)
	return func() {
		os.Unsetenv(name)
	}
}
