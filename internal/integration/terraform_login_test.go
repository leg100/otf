package integration

import (
	"fmt"
	"os"
	"os/exec"
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
	integrationTest(t)

	svc, _, ctx := setup(t, nil)

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

	browser.Run(t, ctx, chromedp.Tasks{
		// navigate to auth url captured from terraform login output
		chromedp.Navigate(strings.TrimSpace(u)),
		screenshot(t, "terraform_login_consent"),
		// give consent
		chromedp.Click(`//button[text()='Accept']`),
		screenshot(t, "terraform_login_flow_complete"),
		matchText(t, `//body/p`, `The login server has returned an authentication code to Terraform.`),
	})

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
	cmd := exec.Command("terraform", "init")
	cmd.Dir = configPath
	assert.NoError(t, cmd.Run())
}
