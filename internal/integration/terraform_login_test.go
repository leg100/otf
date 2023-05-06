package integration

import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	expect "github.com/google/goexpect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTerraformLogin demonstrates create a user token via the UI and passing it
// to `terraform login`
func TestTerraformLogin(t *testing.T) {
	t.Parallel()

	svc := setup(t, nil)
	user, ctx := svc.createUserCtx(t, ctx)

	var token string
	browser := createBrowserCtx(t)
	err := chromedp.Run(browser, chromedp.Tasks{
		newSession(t, ctx, svc.Hostname(), user.Username, svc.Secret),
		chromedp.Navigate("https://" + svc.Hostname()),
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
				[]string{"terraform", "login", svc.Hostname()},
				time.Minute,
				expect.PartialMatch(true),
				// expect.Verbose(testing.Verbose()),
				expect.Tee(out),
				expect.SetEnv(
					append(envs, fmt.Sprintf("PATH=%s:%s", killBrowserPath, os.Getenv("PATH"))),
				),
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
	})
	require.NoError(t, err)
}
