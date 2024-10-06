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

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal/run"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

var (
	// map test name to a count of number of screenshots taken
	screenshotRecord map[string]int
	screenshotMutex  sync.Mutex

	// waitLoaded waits for a page to be fully loaded - this is important for
	// the OTF UI, because the execution alpine.js is deferred and once it
	// executes it meddles with the DOM, and only once that has finished should
	// actions such as chromedp.Click be invoked. Otherwise the dreaded "-32000
	// node not found" error arises.
	waitLoaded = chromedp.ActionFunc(func(ctx context.Context) error {
		ch := make(chan struct{})
		lctx, cancel := context.WithCancel(ctx)
		chromedp.ListenTarget(lctx, func(ev interface{}) {
			if _, ok := ev.(*page.EventLoadEventFired); ok {
				cancel()
				close(ch)
			}
		})
		select {
		case <-ch:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
)

// createWorkspace creates a workspace via the UI
func createWorkspace(t *testing.T, page playwright.Page, hostname, org, name string) {
	t.Helper()

	_, err := page.Goto(organizationURL(hostname, org))
	require.NoError(t, err)

	err = page.Locator("#menu-item-workspaces > a").Click()
	require.NoError(t, err)

	err = page.Locator("#new-workspace-button").Click()
	require.NoError(t, err)

	err = page.Locator("input#name").Fill(name)
	require.NoError(t, err)

	err = page.Locator("#create-workspace-button").Click()
	require.NoError(t, err)

	err = expect.Locator(page.Locator("//div[@role='alert']")).ToHaveText("created workspace: " + name)
	require.NoError(t, err)
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
func addWorkspacePermission(t *testing.T, page playwright.Page, hostname, org, workspaceName, teamID, role string) {
	t.Helper()

	// go to workspace
	_, err := page.Goto(workspaceURL(hostname, org, workspaceName))
	require.NoError(t, err)
	//screenshot(t),

	// go to workspace settings
	err = page.Locator(`//a[text()='settings']`).Click()
	require.NoError(t, err)

	// confirm builtin admin role for owners team
	err = expect.Locator(page.Locator("#permissions-owners td:first-child")).ToHaveText("owners")
	require.NoError(t, err)

	err = expect.Locator(page.Locator("#permissions-owners td:last-child")).ToHaveText("admin")
	require.NoError(t, err)

	// assign role to team
	chromedp.SetValue(`//select[@id="permissions-add-select-role"]`, role),
		chromedp.SetValue(`//select[@id="permissions-add-select-team"]`, teamID),
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
		//screenshot(t, "workspace_permissions"),
		err := page.Locator("#permissions-add-button").Click()
	require.NoError(t, err)

	//screenshot(t),
	err = expect.Locator(page.Locator("//div[@role='alert']")).ToHaveText("updated workspace permissions")
	require.NoError(t, err)
}

// startRunTasks starts a run via the UI
func startRunTasks(t *testing.T, page playwright.Page, hostname, organization, workspaceName string, op run.Operation) {
	t.Helper()

	// go to workspace page
	_, err := page.Goto(workspaceURL(hostname, organization, workspaceName))
	require.NoError(t, err)
	//screenshot(t, "connected_workspace_main_page"),

	// select operation for run
	err = page.Locator(`//select[@id="start-run-operation"]`).
		require.NoError(t, err)
	//screenshot(t, "run_page_started"),

	// confirm plan begins and ends
	err = expect.Locator(page.Locator(`//*[@id='tailed-plan-logs']//text()[contains(.,'Initializing the backend')]`)).ToBeAttached()
	require.NoError(t, err)
	//screenshot(t),

	err = expect.Locator(page.Locator(`//span[@id='plan-status' and text()='finished']`)).ToBeVisible()
	require.NoError(t, err)
	//screenshot(t),

	// wait for run to enter planned state
	err = expect.Locator(page.Locator(`//div[@class='widget']//a[text()='planned']`)).ToBeVisible()
	require.NoError(t, err)
	//screenshot(t),

	// run widget should show plan summary
	err = expect.Locator(page.Locator(`//div[@class='widget']//div[@id='resource-summary']`)).ToHaveText(regexp.MustCompile(`\+[0-9]+ \~[0-9]+ \-[0-9]+`))
	require.NoError(t, err)
	//screenshot(t, "run_page_planned_state"),

	// run widget should show discard button
	err = expect.Locator(page.Locator(`//button[@id='run-discard-button']`)).ToBeVisible()
	require.NoError(t, err)
	//screenshot(t),

	// click 'apply' button once it becomes visible
	err = page.Locator(`//button[text()='apply']`).Click()
	require.NoError(t, err)
	//screenshot(t),

	// confirm apply begins and ends
	expect.Locator(page.Locator(`//*[@id='tailed-apply-logs']//text()[contains(.,'Initializing the backend')]`)).ToBeAttached()
	err = expect.Locator(page.Locator(`//span[@id='apply-status' and text()='finished']`)).ToBeVisible()
	require.NoError(t, err)

	// confirm run ends in applied state
	err = expect.Locator(page.Locator(`//div[@class='widget']//a[text()='applied']`)).ToBeVisible()
	require.NoError(t, err)

	// run widget should show apply summary
	err = expect.Locator(page.Locator(`//div[@class='widget']//div[@id='resource-summary']`)).ToHaveText(regexp.MustCompile(`\+[0-9]+ \~[0-9]+ \-[0-9]+`))
	require.NoError(t, err)

	// because run was triggered from the UI, the UI icon should be visible.
	err = expect.Locator(page.Locator(`//div[@class='widget']//img[@id='run-trigger-ui']`)).ToBeVisible()
	require.NoError(t, err)

	// run should show elapsed time
	err = expect.Locator(page.Locator(`//div[@id='elapsed-time']/span`)).ToHaveText(regexp.MustCompile(`\d+(s|ms)`))
	require.NoError(t, err)

	// plan should show running time
	err = expect.Locator(page.Locator(`//span[@id='running-time-plan']`)).ToHaveText(regexp.MustCompile(`\d+(s|ms)`))
	require.NoError(t, err)

	// apply should show running time
	err = expect.Locator(page.Locator(`//span[@id='running-time-apply']`)).ToHaveText(regexp.MustCompile(`\d+(s|ms)`))
	require.NoError(t, err)
	//screenshot(t),
}

func connectWorkspaceTasks(t *testing.T, page playwright.Page, hostname, org, name, provider string) {
	t.Helper()

	// go to workspace
	_, err := page.Goto(workspaceURL(hostname, org, name))
	require.NoError(t, err)
	//screenshot(t, "workspace_main_page"),

	// navigate to workspace settings
	err = page.Locator(`//a[text()='settings']`).Click()
	require.NoError(t, err)
	//screenshot(t, "workspace_settings"),

	// click connect button
	err = page.Locator(`//button[@id='list-workspace-vcs-providers-button']`).Click()
	require.NoError(t, err)
	//screenshot(t, "workspace_vcs_providers_list"),

	// select provider
	err = page.Locator(`div.widget`).Click()
	require.NoError(t, err)
	//screenshot(t, "workspace_vcs_repo_list"),

	// connect to first repo in list (there should only be one)
	err = page.Locator(`//div[@id='content-list']//button[text()='connect']`).Click()
	require.NoError(t, err)
	//screenshot(t),

	// confirm connected
	err = expect.Locator(page.Locator("//div[@role='alert']")).ToHaveText("connected workspace to repo")
	require.NoError(t, err)
}

func disconnectWorkspaceTasks(t *testing.T, page playwright.Page, hostname, org, name string) {
	t.Helper()

	// go to workspace
	_, err := page.Goto(workspaceURL(hostname, org, name))
	require.NoError(t, err)
	//screenshot(t),

	// navigate to workspace settings
	err = page.Locator(`//a[text()='settings']`).Click()
	require.NoError(t, err)
	//screenshot(t),

	// click disconnect button
	err = page.Locator(`//button[@id='disconnect-workspace-repo-button']`).Click()
	require.NoError(t, err)
	//screenshot(t),

	// confirm disconnected
	err = expect.Locator(page.Locator("//div[@role='alert']")).ToHaveText("disconnected workspace from repo")
	require.NoError(t, err)
}

func reloadUntilVisible(t *testing.T, page playwright.Page, sel string) {
	for {
		visible, err := page.Locator(sel).IsVisible()
		require.NoError(t, err)

		if visible {
			return
		}

		time.Sleep(time.Second)

		_, err = page.Reload()
		require.NoError(t, err)
	}
}
