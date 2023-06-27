package integration

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal/run"
)

var (
	// map test name to a count of number of screenshots taken
	screenshotRecord map[string]int
	screenshotMutex  sync.Mutex
)

// createWorkspace creates a workspace via the UI
func createWorkspace(t *testing.T, hostname, org, name string) chromedp.Tasks {
	t.Helper()

	return chromedp.Tasks{
		chromedp.Navigate(organizationURL(hostname, org)),
		chromedp.Click("#menu-item-workspaces > a", chromedp.ByQuery),
		chromedp.Click("#new-workspace-button", chromedp.ByQuery),
		chromedp.Focus("input#name", chromedp.NodeVisible, chromedp.ByQuery),
		input.InsertText(name),
		chromedp.Click("#create-workspace-button", chromedp.ByQuery),
		matchText(t, ".flash-success", "created workspace: "+name, chromedp.ByQuery),
	}
}

// matchText is a custom chromedp Task that extracts text content using the
// selector and asserts that it matches the wanted string.
func matchText(t *testing.T, selector, want string, opts ...chromedp.QueryOption) chromedp.ActionFunc {
	t.Helper()

	return matchRegex(t, selector, "^"+want+"$")
}

// matchRegex is a custom chromedp Task that extracts text content using the
// selector and asserts that it matches the regular expression.
func matchRegex(t *testing.T, selector, regex string, opts ...chromedp.QueryOption) chromedp.ActionFunc {
	t.Helper()

	return func(ctx context.Context) error {
		var got string
		opts := append(opts, chromedp.NodeVisible)
		if err := chromedp.WaitReady(selector, opts...).Do(ctx); err != nil {
			return fmt.Errorf("matchRegex: waiting for %s: %w", selector, err)
		}
		if err := chromedp.Text(selector, &got, opts...).Do(ctx); err != nil {
			return fmt.Errorf("matching selector %s with regex %s: %w", selector, regex, err)
		}
		if !regexp.MustCompile(regex).MatchString(strings.TrimSpace(got)) {
			return fmt.Errorf("regex %s does not match %s", regex, got)
		}
		return nil
	}
}

// screenshot takes a screenshot of a browser and saves it to disk, using the
// test name and a counter to uniquely name the file.
func screenshot(t *testing.T, docPath ...string) chromedp.ActionFunc {
	t.Helper()

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
		err := chromedp.WaitVisible(`body`, chromedp.ByQuery).Do(ctx)
		if err != nil {
			return fmt.Errorf("waiting for body to be visible before capturing screenshot: %w", err)
		}
		err = chromedp.CaptureScreenshot(&image).Do(ctx)
		if err != nil {
			return fmt.Errorf("caputuring screenshot: %w", err)
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
	t.Helper()

	return chromedp.Tasks{
		// go to workspace
		chromedp.Navigate(workspaceURL(hostname, org, workspaceName)),
		screenshot(t),
		// go to workspace settings
		chromedp.Click(`//a[text()='settings']`),
		// confirm builtin admin role for owners team
		matchText(t, "#permissions-owners td:first-child", "owners", chromedp.ByQuery),
		matchText(t, "#permissions-owners td:last-child", "admin", chromedp.ByQuery),
		// assign role to team
		chromedp.SetValue(`//select[@id="permissions-add-select-role"]`, role),
		chromedp.SetValue(`//select[@id="permissions-add-select-team"]`, team),
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
		chromedp.Click("#permissions-add-button", chromedp.ByQuery),
		screenshot(t),
		matchText(t, ".flash-success", "updated workspace permissions", chromedp.ByQuery),
	}
}

func createGithubVCSProviderTasks(t *testing.T, hostname, org, name string) chromedp.Tasks {
	t.Helper()

	return chromedp.Tasks{
		// go to org
		chromedp.Navigate(organizationURL(hostname, org)),
		screenshot(t, "organization_main_menu"),
		// go to vcs providers
		chromedp.Click("#vcs_providers > a", chromedp.ByQuery),
		screenshot(t, "vcs_providers_list"),
		// click 'New Github VCS Provider' button
		chromedp.Click(`//button[text()='New Github VCS Provider']`),
		screenshot(t, "new_github_vcs_provider_form"),
		// enter fake github token and name
		chromedp.Focus("input#token", chromedp.NodeVisible, chromedp.ByQuery),
		input.InsertText("fake-github-personal-token"),
		chromedp.Focus("input#name", chromedp.ByQuery),
		input.InsertText(name),
		screenshot(t),
		// submit form to create provider
		chromedp.Submit("input#token", chromedp.ByQuery),
		screenshot(t),
		matchText(t, ".flash-success", "created provider: github", chromedp.ByQuery),
	}
}

// startRunTasks starts a run via the UI
func startRunTasks(t *testing.T, hostname, organization, workspaceName string, op run.Operation) chromedp.Tasks {
	t.Helper()

	return []chromedp.Action{
		// go to workspace page
		chromedp.Navigate(workspaceURL(hostname, organization, workspaceName)),
		screenshot(t, "connected_workspace_main_page"),
		// select operation for run
		chromedp.SetValue(`//select[@id="start-run-operation"]`, string(op)),
		screenshot(t, "run_page_started"),
		// confirm plan begins and ends
		chromedp.WaitReady(`//*[@id='tailed-plan-logs']//text()[contains(.,'Initializing the backend')]`),
		screenshot(t),
		chromedp.WaitVisible(`#plan-status.phase-status-finished`),
		screenshot(t),
		// wait for run to enter planned state
		chromedp.WaitVisible(`//*[@class='status status-planned']`),
		screenshot(t),
		// run widget should show plan summary
		matchRegex(t, `//div[@class='item']//div[@class='resource-summary']`, `\+[0-9]+ \~[0-9]+ \-[0-9]+`),
		screenshot(t, "run_page_planned_state"),
		// run widget should show discard button
		chromedp.WaitVisible(`//button[@id='run-discard-button']`),
		screenshot(t),
		// click 'apply' button once it becomes visible
		chromedp.Click(`//button[text()='apply']`),
		screenshot(t),
		// confirm apply begins and ends
		chromedp.WaitReady(`//*[@id='tailed-apply-logs']//text()[contains(.,'Initializing the backend')]`),
		chromedp.WaitVisible(`#apply-status.phase-status-finished`, chromedp.ByQuery),
		// confirm run ends in applied state
		chromedp.WaitVisible(`//*[@class='status status-applied']`),
		// run widget should show apply summary
		matchRegex(t, `//div[@class='item']//div[@class='resource-summary']`, `\+[0-9]+ \~[0-9]+ \-[0-9]+`),
		screenshot(t),
	}
}

func connectWorkspaceTasks(t *testing.T, hostname, org, name string) chromedp.Tasks {
	t.Helper()

	return chromedp.Tasks{
		// go to workspace
		chromedp.Navigate(workspaceURL(hostname, org, name)),
		screenshot(t, "workspace_main_page"),
		// navigate to workspace settings
		chromedp.Click(`//a[text()='settings']`),
		screenshot(t, "workspace_settings"),
		// click connect button
		chromedp.Click(`//button[@id='list-workspace-vcs-providers-button']`),
		screenshot(t, "workspace_vcs_providers_list"),
		// select provider
		chromedp.Click(`//a[normalize-space(text())='github']`),
		screenshot(t, "workspace_vcs_repo_list"),
		// connect to first repo in list (there should only be one)
		chromedp.Click(`//div[@class='content-list']//button[text()='connect']`),
		screenshot(t),
		// confirm connected
		matchText(t, ".flash-success", "connected workspace to repo", chromedp.ByQuery),
	}
}

func disconnectWorkspaceTasks(t *testing.T, hostname, org, name string) chromedp.Tasks {
	t.Helper()

	return chromedp.Tasks{
		// go to workspace
		chromedp.Navigate(workspaceURL(hostname, org, name)),
		screenshot(t),
		// navigate to workspace settings
		chromedp.Click(`//a[text()='settings']`),
		screenshot(t),
		// click disconnect button
		chromedp.Click(`//button[@id='disconnect-workspace-repo-button']`),
		screenshot(t),
		// confirm disconnected
		matchText(t, ".flash-success", "disconnected workspace from repo", chromedp.ByQuery),
	}
}

func reloadUntilVisible(sel string) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		var nodes []*cdp.Node
		for {
			err := chromedp.Nodes(sel, &nodes, chromedp.AtLeast(0)).Do(ctx)
			if err != nil {
				return err
			}
			if len(nodes) > 0 {
				return nil
			}
			err = chromedp.Sleep(time.Second).Do(ctx)
			if err != nil {
				return err
			}
			err = chromedp.Reload().Do(ctx)
			if err != nil {
				return err
			}
		}
	})
}
