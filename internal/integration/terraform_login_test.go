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
	"github.com/leg100/otf/internal"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTerraformLogin demonstrates using `terraform login` to retrieve
// credentials.
func TestTerraformLogin(t *testing.T) {
	integrationTest(t)

	tests := []struct {
		name string
		path string
	}{
		{"Terraform", terraformPath},
		{"OpenTofu", tofuPath},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, ctx := setup(t)

			out, err := os.CreateTemp(t.TempDir(), "terraform-login.out")
			require.NoError(t, err)

			// prevent engine from automatically opening a browser
			wd, err := os.Getwd()
			require.NoError(t, err)
			killBrowserPath := path.Join(wd, "./fixtures/kill-browser")

			e, tferr, err := goexpect.SpawnWithArgs(
				[]string{tt.path, "login", svc.System.Hostname()},
				time.Minute,
				goexpect.PartialMatch(true),
				goexpect.Verbose(testing.Verbose()),
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

			// capture url
			url, _, err := e.Expect(regexp.MustCompile(`https://.*\n.*`), -1)
			require.NoError(t, err)

			browser.New(t, ctx, func(page playwright.Page) {
				// navigate to auth url captured from engine login output
				_, err = page.Goto(strings.TrimSpace(url))
				require.NoError(t, err)
				screenshot(t, page, "engine_login_consent")

				// give consent
				err = page.Locator(`//button[text()='Accept']`).Click()
				require.NoError(t, err)

				// user is now redirected to the temp http server run by the
				// engine bin, which presents a short message indicating
				// success.
				//
				// NOTE: disabled on GHA because it is mysteriously flaky when invoked
				// on a GHA runner.
				if _, isGithubActions := os.LookupEnv("CI"); !isGithubActions {
					err = expect.Locator(page.Locator(`//body/p[1]`)).ToHaveText(
						fmt.Sprintf(
							`The login server has returned an authentication code to %s.`,
							tt.name,
						),
					)
					assert.NoError(t, err)
				}
			})

			err = <-tferr
			if !assert.NoError(t, err) || t.Failed() {
				logs, err := os.ReadFile(out.Name())
				require.NoError(t, err)
				t.Log("--- engine login output ---")
				t.Log(string(logs))
				return
			}

			// The engine binary now exits. Because it is no longer running we cannot
			// use expect. Instead check contents of output for a success message.
			logs, err := os.ReadFile(out.Name())
			require.NoError(t, err)
			unescapedLogs := internal.StripAnsi(string(logs))
			require.Contains(t, unescapedLogs, fmt.Sprintf(
				`Success! %s has obtained and saved an API token.`,
				tt.name,
			))

			// create some terraform config and run engine's init cmd to
			// demonstrate user has authenticated successfully.
			{
				org := svc.createOrganization(t, ctx)
				configPath := newRootModule(t, svc.System.Hostname(), org.Name, "my-workspace")
				cmd := exec.Command(tt.path, "init")
				cmd.Dir = configPath
				out, err := cmd.CombinedOutput()
				assert.NoError(t, err, string(out))
			}
		})
	}
}
