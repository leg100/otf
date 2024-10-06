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

	goexpect "github.com/google/goexpect"
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

	tfpath := svc.downloadTerraform(t, ctx, nil)

	e, tferr, err := goexpect.SpawnWithArgs(
		[]string{tfpath, "login", svc.System.Hostname()},
		time.Minute,
		goexpect.PartialMatch(true),
		// expect.Verbose(testing.Verbose()),
		goexpect.Tee(out),
		goexpect.SetEnv(
			append(sharedEnvs, fmt.Sprintf("PATH=%s:%s", killBrowserPath, os.Getenv("PATH"))),
		),
	)
	require.NoError(t, err)
	defer e.Close()

	_, _, err = e.Expect(regexp.MustCompile(`Enter a value:`), -1)
	require.NoError(t, err)

	err = e.Send("yes\n")
	require.NoError(t, err)

	_, _, err = e.Expect(regexp.MustCompile(`Open the following URL to access the login page for 127.0.0.1:[0-9]+:`), -1)
	require.NoError(t, err)

	url, _, err := e.Expect(regexp.MustCompile(`https://.*\n.*`), -1)
	require.NoError(t, err)

	page := browser.New(t, ctx)

	// navigate to auth url captured from terraform login output
	_, err = page.Goto(strings.TrimSpace(url))
	require.NoError(t, err)
	screenshot(t, page, "terraform_login_consent")

	// give consent
	err = page.Locator(`//button[text()='Accept']`).Click()
	require.NoError(t, err)

	screenshot(t, page, "terraform_login_flow_complete")
	err = expect.Locator(page.Locator(`//body/p[1]`)).ToHaveText(`The login server has returned an authentication code to Terraform.`)
	require.NoError(t, err)

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
	configPath := newRootModule(t, svc.System.Hostname(), org.Name, t.Name())
	cmd := exec.Command(tfpath, "init")
	cmd.Dir = configPath
	assert.NoError(t, cmd.Run())
}
