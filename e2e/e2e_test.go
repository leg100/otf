package e2e

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// chromedp browser config
	allocator context.Context
	// for taking browser screenshots
	ss = &screenshotter{m: make(map[string]int)}
)

func TestMain(t *testing.M) {
	headless := true
	if v, ok := os.LookupEnv("OTF_E2E_HEADLESS"); ok {
		var err error
		headless, err = strconv.ParseBool(v)
		if err != nil {
			panic("cannot parse OTF_E2E_HEADLESS")
		}
	}

	var cancel context.CancelFunc
	allocator, cancel = chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", headless),
			chromedp.Flag("hide-scrollbars", true),
			chromedp.Flag("mute-audio", true),
			chromedp.Flag("ignore-certificate-errors", true),
			chromedp.Flag("disable-gpu", true),
		)...)
	defer cancel()

	os.Exit(t.Run())
}

func createWebWorkspace(t *testing.T, ctx context.Context, url string, org string) string {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	var gotFlashSuccess string
	workspaceName := "workspace-" + otf.GenerateRandomString(4)
	orgSelector := fmt.Sprintf("#item-organization-%s a", org)

	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Click(".login-button-github", chromedp.NodeVisible),
		chromedp.Click(orgSelector, chromedp.NodeVisible),
		ss.screenshot(t),
		chromedp.Click("#menu-item-workspaces > a", chromedp.ByQuery),
		// sometimes get stuck on this one...
		chromedp.Click("#new-workspace-button", chromedp.NodeVisible, chromedp.ByQuery),
		ss.screenshot(t),
		chromedp.Focus("input#name", chromedp.NodeVisible),
		input.InsertText(workspaceName),
		chromedp.Click("#create-workspace-button"),
		ss.screenshot(t),
		chromedp.Text(".flash-success", &gotFlashSuccess, chromedp.NodeVisible),
	})
	require.NoError(t, err)

	assert.Equal(t, "created workspace: "+workspaceName, strings.TrimSpace(gotFlashSuccess))

	return workspaceName
}

// addWorkspacePermission adds a workspace permission via the web app, assigning
// a role to a team on a workspace in an org.
func addWorkspacePermission(t *testing.T, allocater context.Context, url, org, workspace, team, role string) {
	ctx, cancel := chromedp.NewContext(allocater)
	defer cancel()

	var gotOwnersTeam string
	var gotOwnersRole string
	var gotFlashSuccess string

	orgSelector := fmt.Sprintf("#item-organization-%s a", org)
	workspaceSelector := fmt.Sprintf("#item-workspace-%s a", workspace)
	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(url),
		// login
		chromedp.Click(".login-button-github", chromedp.NodeVisible),
		chromedp.WaitReady(`body`),
		// select org
		chromedp.Click(orgSelector, chromedp.NodeVisible),
		chromedp.WaitReady(`body`),
		// list workspaces
		chromedp.Click("#menu-item-workspaces > a", chromedp.NodeVisible, chromedp.ByQuery),
		chromedp.WaitReady(`body`),
		// select workspace
		chromedp.Click(workspaceSelector, chromedp.NodeVisible),
		ss.screenshot(t),
		// confirm builtin admin permission for owners team
		chromedp.Text("#permissions-owners td:first-child", &gotOwnersTeam, chromedp.NodeVisible),
		chromedp.Text("#permissions-owners td:last-child", &gotOwnersRole, chromedp.NodeVisible),
		// assign role to team
		chromedp.SetValue(`//select[@id="permissions-add-select-role"]`, role, chromedp.BySearch),
		chromedp.SetValue(`//select[@id="permissions-add-select-team"]`, team, chromedp.BySearch),
		chromedp.Click("#permissions-add-button", chromedp.NodeVisible),
		ss.screenshot(t),
		chromedp.Text(".flash-success", &gotFlashSuccess, chromedp.NodeVisible),
	})
	require.NoError(t, err)

	assert.Equal(t, "owners", gotOwnersTeam)
	assert.Equal(t, "admin", gotOwnersRole)
	assert.Equal(t, "updated workspace permissions", gotFlashSuccess)
}
