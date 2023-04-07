package integration

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	expect "github.com/google/goexpect"
	"github.com/leg100/otf/tokens"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// map test name to a count of number of screenshots taken
	screenshotRecord map[string]int
	screenshotMutex  sync.Mutex
)

// newSession adds a user session to the browser cookie jar
func newSession(t *testing.T, ctx context.Context, hostname, username, secret string) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		token := tokens.NewTestSessionJWT(t, username, secret, time.Hour)
		return network.SetCookie("session", token).WithDomain(hostname).Do(ctx)
	})
}

// createWorkspace creates a workspace via the UI
func createWorkspace(t *testing.T, hostname, org, name string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(organizationPath(hostname, org)),
		screenshot(t),
		chromedp.Click("#menu-item-workspaces > a", chromedp.ByQuery),
		screenshot(t),
		chromedp.Click("#new-workspace-button", chromedp.NodeVisible, chromedp.ByQuery),
		screenshot(t),
		chromedp.Focus("input#name", chromedp.NodeVisible),
		input.InsertText(name),
		chromedp.Click("#create-workspace-button"),
		screenshot(t),
		matchText(t, ".flash-success", "created workspace: "+name),
	}
}

// matchText is a custom chromedp Task that extracts text content using the
// selector and asserts that it matches the wanted string.
func matchText(t *testing.T, selector, want string) chromedp.ActionFunc {
	return matchRegex(t, selector, "^"+want+"$")
}

// matchRegex is a custom chromedp Task that extracts text content using the
// selector and asserts that it matches the regular expression.
func matchRegex(t *testing.T, selector, regex string) chromedp.ActionFunc {
	t.Helper()

	return func(ctx context.Context) error {
		var got string
		err := chromedp.Text(selector, &got, chromedp.NodeVisible).Do(ctx)
		require.NoError(t, err)
		require.Regexp(t, regex, strings.TrimSpace(got))
		return nil
	}
}

// terraformLoginTasks creates an API token via the UI before passing it to
// 'terraform login'
func terraformLoginTasks(t *testing.T, hostname string) chromedp.Tasks {
	var token string
	return []chromedp.Action{
		// go to profile
		chromedp.Click("#top-right-profile-link > a", chromedp.NodeVisible),
		screenshot(t),
		// go to tokens
		chromedp.Click("#user-tokens-link > a", chromedp.NodeVisible),
		screenshot(t),
		// create new token
		chromedp.Click("#new-user-token-button", chromedp.NodeVisible),
		screenshot(t),
		chromedp.Focus("#description", chromedp.NodeVisible),
		input.InsertText("e2e-test"),
		chromedp.Submit("#description"),
		screenshot(t),
		// capture token
		chromedp.Text(".flash-success > .data", &token, chromedp.NodeVisible),
		// pass token to terraform login
		chromedp.ActionFunc(func(ctx context.Context) error {
			out, err := os.CreateTemp(t.TempDir(), "terraform-login.out")
			require.NoError(t, err)

			// prevent terraform from automatically opening a browser
			wd, err := os.Getwd()
			require.NoError(t, err)
			killBrowserPath := path.Join(wd, "./fixtures/kill-browser")

			e, tferr, err := expect.SpawnWithArgs(
				[]string{"terraform", "login", hostname},
				time.Minute,
				expect.PartialMatch(true),
				expect.Verbose(testing.Verbose()),
				expect.Tee(out),
				expect.SetEnv([]string{
					fmt.Sprintf("PATH=%s:%s", killBrowserPath, os.Getenv("PATH")),
					"SSL_CERT_FILE=./fixtures/cert.pem",
				}),
			)
			require.NoError(t, err)
			defer e.Close()

			e.ExpectBatch([]expect.Batcher{
				&expect.BExp{R: "Enter a value:"}, &expect.BSnd{S: "yes\n"},
				&expect.BExp{R: "Enter a value:"}, &expect.BSnd{S: token + "\n"},
				&expect.BExp{R: "Success! Logged in to Terraform Enterprise"},
			}, time.Minute)
			err = <-tferr
			if !assert.NoError(t, err) || t.Failed() {
				logs, err := os.ReadFile(out.Name())
				require.NoError(t, err)
				t.Log("--- terraform login output ---")
				t.Log(string(logs))
			}
			return err
		}),
	}
}

// screenshot takes a screenshot of a browser and saves it to disk, using the
// test name and a counter to uniquely name the file.
func screenshot(t *testing.T) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		screenshotMutex.Lock()
		defer screenshotMutex.Unlock()

		// increment counter
		if screenshotRecord == nil {
			screenshotRecord = make(map[string]int)
		}
		counter, ok := screenshotRecord[t.Name()]
		if !ok {
			screenshotRecord[t.Name()] = 0
		}
		screenshotRecord[t.Name()]++

		// take screenshot
		var image []byte
		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.WaitReady(`body`),
			chromedp.CaptureScreenshot(&image),
		})
		if err != nil {
			return err
		}

		// save image to disk
		fname := path.Join("screenshots", fmt.Sprintf("%s_%02d.png", t.Name(), counter))
		err = os.MkdirAll(filepath.Dir(fname), 0o755)
		if err != nil {
			return err
		}
		err = os.WriteFile(fname, image, 0o644)
		if err != nil {
			return err
		}
		return nil
	}
}

// okDialog - Click OK on any browser javascript dialog boxes that pop up
func okDialog(t *testing.T, ctx context.Context) {
	t.Helper()

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
