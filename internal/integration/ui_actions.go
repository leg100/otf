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
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal/tokens"
	"github.com/stretchr/testify/require"
)

var (
	// map test name to a count of number of screenshots taken
	screenshotRecord map[string]int
	screenshotMutex  sync.Mutex
)

// newSession adds a user session to the browser cookie jar
func newSession(t *testing.T, ctx context.Context, hostname, username string, secret []byte) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		token := tokens.NewTestSessionJWT(t, username, secret, time.Hour)
		return network.SetCookie("session", token).WithDomain(hostname).Do(ctx)
	})
}

// createWorkspace creates a workspace via the UI
func createWorkspace(t *testing.T, hostname, org, name string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(organizationURL(hostname, org)),
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
	t.Helper()

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
func screenshot(t *testing.T, docPath ...string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		// disable screenshots if headless mode is disabled - screenshots are
		// most likely unnecessary if the developer is using headless mode to
		// view the browser; and, depending on the developer's monitor, the
		// viewport in the screenshots is different to that when headless mode
		// is enabled, but we want the viewport to be consistent because
		// screenshots are also used in the documentation!
		if headless, ok := os.LookupEnv("OTF_E2E_HEADLESS"); ok {
			if headless == "false" {
				return nil
			}
		}

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

		//
		// additionally, save the screenshot image in the docs directory too,
		// but only if a path is specified AND the relevant env var is
		// specified.
		//
		if len(docPath) == 0 {
			return nil
		}
		if docScreenshots, ok := os.LookupEnv("OTF_DOC_SCREENSHOTS"); !ok {
			return nil
		} else if docScreenshots != "true" {
			return nil
		}

		fname = path.Join("..", "..", "docs", "images", docPath[0]+".png")
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
		chromedp.Navigate(workspaceURL(hostname, org, workspaceName)),
		screenshot(t),
		// go to workspace settings
		chromedp.Click(`//a[text()='settings']`, chromedp.NodeVisible),
		chromedp.WaitReady(`body`),
		// confirm builtin admin role for owners team
		matchText(t, "#permissions-owners td:first-child", "owners"),
		matchText(t, "#permissions-owners td:last-child", "admin"),
		// assign role to team
		chromedp.SetValue(`//select[@id="permissions-add-select-role"]`, role, chromedp.BySearch),
		chromedp.SetValue(`//select[@id="permissions-add-select-team"]`, team, chromedp.BySearch),
		// scroll to bottom so that permissions are visible in screenshot
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, exp, err := runtime.Evaluate(`window.scrollTo(0,document.body.scrollHeight);`).Do(ctx)
			if err != nil {
				return err
			}
			if exp != nil {
				return exp
			}
			return nil
		}),
		screenshot(t, "workspace_permissions"),
		chromedp.Click("#permissions-add-button", chromedp.NodeVisible),
		screenshot(t),
		matchText(t, ".flash-success", "updated workspace permissions"),
	}
}

func createGithubVCSProviderTasks(t *testing.T, hostname, org, name string) chromedp.Tasks {
	return chromedp.Tasks{
		// go to org
		chromedp.Navigate(organizationURL(hostname, org)),
		screenshot(t, "organization_main_menu"),
		// go to vcs providers
		chromedp.Click("#vcs_providers > a", chromedp.NodeVisible),
		screenshot(t, "vcs_providers_list"),
		// click 'New Github VCS Provider' button
		chromedp.Click(`//button[text()='New Github VCS Provider']`, chromedp.NodeVisible),
		screenshot(t, "new_github_vcs_provider_form"),
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
func startRunTasks(t *testing.T, hostname, organization, workspaceName, strategy string) chromedp.Tasks {
	return []chromedp.Action{
		// go to workspace page
		chromedp.Navigate(workspaceURL(hostname, organization, workspaceName)),
		screenshot(t, "connected_workspace_main_page"),
		// select strategy for run
		chromedp.SetValue(`//select[@id="start-run-strategy"]`, strategy, chromedp.BySearch),
		screenshot(t, "run_page_started"),
		// confirm plan begins and ends
		chromedp.WaitReady(`body`),
		chromedp.WaitReady(`//*[@id='tailed-plan-logs']//text()[contains(.,'Initializing the backend')]`, chromedp.BySearch),
		screenshot(t),
		chromedp.WaitReady(`#plan-status.phase-status-finished`),
		screenshot(t),
		// wait for run to enter planned state
		chromedp.WaitReady(`//*[@class='status status-planned']`, chromedp.BySearch),
		screenshot(t),
		// run widget should show plan summary
		matchRegex(t, `//div[@class='item']//div[@class='resource-summary']`, `\+[0-9]+ \~[0-9]+ \-[0-9]+`),
		screenshot(t, "run_page_planned_state"),
		// run widget should show discard button
		chromedp.WaitReady(`//button[@id='run-discard-button']`, chromedp.BySearch),
		screenshot(t),
		// click 'apply' button once it becomes visible
		chromedp.Click(`//button[text()='apply']`, chromedp.NodeVisible, chromedp.BySearch),
		screenshot(t),
		// confirm apply begins and ends
		chromedp.WaitReady(`//*[@id='tailed-apply-logs']//text()[contains(.,'Initializing the backend')]`, chromedp.BySearch),
		chromedp.WaitReady(`#apply-status.phase-status-finished`),
		// confirm run ends in applied state
		chromedp.WaitReady(`//*[@class='status status-applied']`, chromedp.BySearch),
		// run widget should show apply summary
		matchRegex(t, `//div[@class='item']//div[@class='resource-summary']`, `\+[0-9]+ \~[0-9]+ \-[0-9]+`),
		screenshot(t),
	}
}

func connectWorkspaceTasks(t *testing.T, hostname, org, name string) chromedp.Tasks {
	return chromedp.Tasks{
		// go to workspace
		chromedp.Navigate(workspaceURL(hostname, org, name)),
		screenshot(t, "workspace_main_page"),
		// navigate to workspace settings
		chromedp.Click(`//a[text()='settings']`, chromedp.NodeVisible),
		screenshot(t, "workspace_settings"),
		// click connect button
		chromedp.Click(`//button[@id='list-workspace-vcs-providers-button']`, chromedp.NodeVisible),
		screenshot(t, "workspace_vcs_providers_list"),
		// select provider
		chromedp.Click(`//a[normalize-space(text())='github']`, chromedp.NodeVisible),
		screenshot(t, "workspace_vcs_repo_list"),
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
		chromedp.Navigate(workspaceURL(hostname, org, name)),
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
