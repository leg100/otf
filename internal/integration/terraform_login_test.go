package integration

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	expect "github.com/google/goexpect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTerraformLogin demonstrates using `terraform login` to retrieve
// credentials.
func TestTerraformLogin(t *testing.T) {
	t.Parallel()

	svc := setup(t, nil)
	user, ctx := svc.createUserCtx(t, ctx)

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

	e.Expect(regexp.MustCompile(`Enter a value:`), -1)
	e.Send("yes\n")
	e.Expect(regexp.MustCompile(`Open the following URL to access the login page for 127.0.0.1:[0-9]+:`), -1)
	u, _, _ := e.Expect(regexp.MustCompile(`https://.*\n.*`), -1)

	browser := createBrowserCtx(t)
	err = chromedp.Run(browser, chromedp.Tasks{
		newSession(t, ctx, svc.Hostname(), user.Username, svc.Secret),
		// navigate to auth url captured from terraform login output
		chromedp.Navigate(strings.TrimSpace(u)),
		screenshot(t),
		// give consent
		chromedp.Click(`//button[text()='Accept']`, chromedp.NodeVisible),
		screenshot(t),
		matchText(t, "//body/p", "The login server has returned an authentication code to Terraform."),
	})
	require.NoError(t, err)

	e.Expect(regexp.MustCompile(`Success! Terraform has obtained and saved an API token.`), -1)

	err = <-tferr
	if !assert.NoError(t, err) || t.Failed() {
		logs, err := os.ReadFile(out.Name())
		require.NoError(t, err)
		t.Log("--- terraform login output ---")
		t.Log(string(logs))
		return
	}

	// create some terraform config and run terraform init to demonstrate user
	// has authenticated successfully.
	org := svc.createOrganization(t, ctx)
	configPath := newRootModule(t, svc.Hostname(), org.Name, t.Name())
	svc.tfcli(t, ctx, "init", configPath)
}
