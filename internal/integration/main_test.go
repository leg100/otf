package integration

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	cdpbrowser "github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/tokens"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

var (
	// shared browser allocator
	allocator context.Context

	// shared browser context
	sharedBrowser context.Context

	// a user with this username is created at the very beginning and the
	// shared browser is seeded with a session belonging to this user
	sharedUsername = "mr-tester"

	// a shared secret which signs the shared user session
	sharedSecret []byte

	clipboardReadPermission  = cdpbrowser.PermissionDescriptor{Name: "clipboard-read"}
	clipboardWritePermission = cdpbrowser.PermissionDescriptor{Name: "clipboard-write"}

	// shared environment variables for individual tests to use
	envs []string

	// Context conferring site admin privileges
	adminCtx = internal.AddSubjectToContext(context.Background(), &auth.SiteAdmin)
)

func TestMain(m *testing.M) {
	code, err := doMain(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup integration tests: %w\n", err)
	}
	os.Exit(code)
}

func doMain(m *testing.M) (int, error) {
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
	unset, err := setenv("SSL_CERT_FILE", filepath.Join(wd, "./fixtures/cert.pem"))
	if err != nil {
		return 0, fmt.Errorf("setting SSL_CERT_FILE: %w", err)
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
		return 0, fmt.Errorf("setting HOME: %w", err)
	}
	defer unset()

	// If HTTPS_PROXY has been defined then add it to the authoritative list of
	// environment variables so that processes, particularly terraform, spawed
	// in tests use the proxy. This can be very useful for caching repeated
	// downloads of terraform providers during tests.
	if proxy, ok := os.LookupEnv("HTTPS_PROXY"); ok {
		envs = append(envs, "HTTPS_PROXY="+proxy)
	}

	// Instruct terraform CLI to skip checks for new versions.
	unset, err = setenv("CHECKPOINT_DISABLE", "true")
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
			return 0, fmt.Errorf("parsing OTF_E2E_HEADLESS: %w", err)
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

	// Create a secret with which to (1) create a user session token and (2)
	// for assignment to daemons so that the token passes verification
	sharedSecret = make([]byte, 16)
	_, err = rand.Read(sharedSecret)
	if err != nil {
		return 0, err
	}

	// Create browser instance for sharing between tests, and seed with a
	// session cookie.
	sharedBrowser, cancel = chromedp.NewContext(allocator)
	defer cancel()
	err = chromedp.Run(sharedBrowser, chromedp.Tasks{
		cdpbrowser.SetPermission(&clipboardReadPermission, cdpbrowser.PermissionSettingGranted).WithOrigin(""),
		cdpbrowser.SetPermission(&clipboardWritePermission, cdpbrowser.PermissionSettingGranted).WithOrigin(""),
		chromedp.ActionFunc(func(ctx context.Context) error {
			key, err := jwk.FromRaw(sharedSecret)
			if err != nil {
				return err
			}
			token, err := tokens.NewSessionToken(key, sharedUsername, internal.CurrentTimestamp().Add(time.Hour))
			if err != nil {
				return err
			}
			return network.SetCookie("session", token).Do(ctx)
		}),
	})
	if err != nil {
		return 0, fmt.Errorf("creating shared browser: %w", err)
	}
	// Click OK on any browser javascript dialog boxes that pop up
	chromedp.ListenTarget(sharedBrowser, func(ev any) {
		switch ev.(type) {
		case *page.EventJavascriptDialogOpening:
			go func() {
				_ = chromedp.Run(sharedBrowser, page.HandleJavaScriptDialog(true))
			}()
		}
	})

	return m.Run(), nil
}

// setenv sets an environment variable and returns a func to unset the variable.
// The environment variable is added to a shared slice, envs, for individual
// tests to use.
func setenv(name, value string) (func(), error) {
	err := os.Setenv(name, value)
	if err != nil {
		return nil, err
	}
	envs = append(envs, name+"="+value)
	return func() {
		os.Unsetenv(name)
	}, nil
}
