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

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/tokens"
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

// addWorkspacePermission adds a workspace permission via the UI, assigning
// a role to a team.
func addWorkspacePermission(t *testing.T, hostname, org, workspaceName, team, role string) chromedp.Tasks {
	return chromedp.Tasks{
		// go to workspace
		chromedp.Navigate(workspacePath(hostname, org, workspaceName)),
		screenshot(t),
		// go to workspace settings
		chromedp.Click(`//a[text()='settings']`, chromedp.NodeVisible),
		screenshot(t),
		// confirm builtin admin permission for owners team
		matchText(t, "#permissions-owners td:first-child", "owners"),
		matchText(t, "#permissions-owners td:last-child", "admin"),
		// assign role to team
		chromedp.SetValue(`//select[@id="permissions-add-select-role"]`, role, chromedp.BySearch),
		chromedp.SetValue(`//select[@id="permissions-add-select-team"]`, team, chromedp.BySearch),
		chromedp.Click("#permissions-add-button", chromedp.NodeVisible),
		screenshot(t),
		matchText(t, ".flash-success", "updated workspace permissions"),
	}
}

func createGithubVCSProviderTasks(t *testing.T, hostname, org, name string) chromedp.Tasks {
	return chromedp.Tasks{
		// go to org
		chromedp.Navigate(organizationPath(hostname, org)),
		// go to vcs providers
		chromedp.Click("#vcs_providers > a", chromedp.NodeVisible),
		screenshot(t),
		// click 'New Github VCS Provider' button
		chromedp.Click(`//button[text()='New Github VCS Provider']`, chromedp.NodeVisible),
		screenshot(t),
		// enter fake github token and name
		chromedp.Focus("input#token", chromedp.NodeVisible),
		input.InsertText("fake-github-personal-token"),
		chromedp.Focus("input#name"),
		input.InsertText(name),
		screenshot(t),
		// submit form to create provider
		chromedp.Submit("input#token"),
		screenshot(t),
		matchText(t, ".flash-success", "created provider: github"),
	}
}

// startRunTasks starts a run via the UI
func startRunTasks(t *testing.T, hostname, organization string, workspaceName string) chromedp.Tasks {
	return []chromedp.Action{
		// go to workspace page
		chromedp.Navigate(workspacePath(hostname, organization, workspaceName)),
		screenshot(t),
		// select strategy for run
		chromedp.SetValue(`//select[@id="start-run-strategy"]`, "plan-and-apply", chromedp.BySearch),
		screenshot(t),
		// confirm plan begins and ends
		chromedp.WaitReady(`body`),
		chromedp.WaitReady(`//*[@id='tailed-plan-logs']//text()[contains(.,'Initializing the backend')]`, chromedp.BySearch),
		screenshot(t),
		chromedp.WaitReady(`#plan-status.phase-status-finished`),
		screenshot(t),
		// wait for run to enter planned state
		chromedp.WaitReady(`//*[@id='run-status']//*[normalize-space(text())='planned']`, chromedp.BySearch),
		screenshot(t),
		// click 'confirm & apply' button once it becomes visible
		chromedp.Click(`//button[text()='Confirm & Apply']`, chromedp.NodeVisible, chromedp.BySearch),
		screenshot(t),
		// confirm apply begins and ends
		chromedp.WaitReady(`//*[@id='tailed-apply-logs']//text()[contains(.,'Initializing the backend')]`, chromedp.BySearch),
		chromedp.WaitReady(`#apply-status.phase-status-finished`),
		// confirm run ends in applied state
		chromedp.WaitReady(`//*[@id='run-status']//*[normalize-space(text())='applied']`, chromedp.BySearch),
		screenshot(t),
	}
}

func connectWorkspaceTasks(t *testing.T, hostname, org, name string) chromedp.Tasks {
	return chromedp.Tasks{
		// go to workspace
		chromedp.Navigate(workspacePath(hostname, org, name)),
		screenshot(t),
		// navigate to workspace settings
		chromedp.Click(`//a[text()='settings']`, chromedp.NodeVisible),
		screenshot(t),
		// click connect button
		chromedp.Click(`//button[text()='Connect to VCS']`, chromedp.NodeVisible),
		screenshot(t),
		// select provider
		chromedp.Click(`//a[normalize-space(text())='github']`, chromedp.NodeVisible),
		screenshot(t),
		// connect to first repo in list (there should only be one)
		chromedp.Click(`//div[@class='content-list']//button[text()='connect']`, chromedp.NodeVisible),
		screenshot(t),
		// confirm connected
		matchText(t, ".flash-success", "connected workspace to repo"),
	}
}

func disconnectWorkspaceTasks(t *testing.T, hostname, org, name string) chromedp.Tasks {
	return chromedp.Tasks{
		// go to workspace
		chromedp.Navigate(workspacePath(hostname, org, name)),
		screenshot(t),
		// navigate to workspace settings
		chromedp.Click(`//a[text()='settings']`, chromedp.NodeVisible),
		screenshot(t),
		// click disconnect button
		chromedp.Click(`//button[@id='disconnect-workspace-repo-button']`, chromedp.NodeVisible),
		screenshot(t),
		// confirm disconnected
		matchText(t, ".flash-success", "disconnected workspace from repo"),
	}
}

func reloadUntilVisible(sel string) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		var nodes []*cdp.Node
		for {
			err := chromedp.Run(ctx, chromedp.Tasks{
				chromedp.Nodes(sel, &nodes, chromedp.AtLeast(0)),
			})
			if err != nil {
				return err
			}
			if len(nodes) > 0 {
				return nil
			}
			err = chromedp.Run(ctx, chromedp.Tasks{
				chromedp.Sleep(time.Second),
				chromedp.Reload(),
			})
			if err != nil {
				return err
			}
		}
	})
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
