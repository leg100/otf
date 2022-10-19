package e2e

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeb(t *testing.T) {
	addBuildsToPath(t)

	headless := true
	if v, ok := os.LookupEnv("OTF_E2E_HEADLESS"); ok {
		var err error
		headless, err = strconv.ParseBool(v)
		require.NoError(t, err)
	}

	org := otf.NewTestOrganization(t)
	team := otf.NewTestTeam(t, org)
	user := otf.NewTestUser(t, otf.WithOrganizationMemberships(org), otf.WithTeamMemberships(team))

	hostname := githubStub(t, user)
	t.Setenv("OTF_GITHUB_HOSTNAME", hostname)

	url := startDaemon(t)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", headless),
			chromedp.Flag("hide-scrollbars", true),
			chromedp.Flag("mute-audio", true),
			chromedp.Flag("ignore-certificate-errors", true),
			chromedp.Flag("disable-gpu", true),
		)...)
	defer cancel()

	t.Run("login", func(t *testing.T) {
		s := screenshotter(0)

		ctx, cancel := chromedp.NewContext(allocCtx)
		defer cancel()

		var gotLoginPrompt string
		var gotLocationOrganizations string

		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Navigate(url),
			s.screenshot(t),
			chromedp.Text(".center", &gotLoginPrompt, chromedp.NodeVisible),
			chromedp.Click(".login-button-github", chromedp.NodeVisible),
			s.screenshot(t),
			chromedp.Location(&gotLocationOrganizations),
		})
		require.NoError(t, err)

		assert.Equal(t, "Login with Github", strings.TrimSpace(gotLoginPrompt))
		assert.Equal(t, url+"/organizations", gotLocationOrganizations)
	})

	t.Run("new workspace", func(t *testing.T) {
		createWebWorkspace(t, allocCtx, url, org)
	})

	t.Run("add workspace permission", func(t *testing.T) {
		s := screenshotter(0)

		workspace := createWebWorkspace(t, allocCtx, url, org)

		ctx, cancel := chromedp.NewContext(allocCtx)
		defer cancel()

		var gotOwnersTeam string
		var gotOwnersRole string
		var gotFlashSuccess string

		orgSelector := fmt.Sprintf("#item-organization-%s a", org.Name())
		workspaceSelector := fmt.Sprintf("#item-workspace-%s a", workspace)
		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Navigate(url),
			// login
			chromedp.Click(".login-button-github", chromedp.NodeVisible),
			// select org
			chromedp.Click(orgSelector, chromedp.NodeVisible),
			// list workspaces
			chromedp.Click("#workspaces > a", chromedp.NodeVisible),
			// select workspace
			chromedp.Click(workspaceSelector, chromedp.NodeVisible),
			s.screenshot(t),
			// confirm builtin admin permission for owners team
			chromedp.Text("#permissions-owners td:first-child", &gotOwnersTeam, chromedp.NodeVisible),
			chromedp.Text("#permissions-owners td:last-child", &gotOwnersRole, chromedp.NodeVisible),
			// add write permission for the test team
			chromedp.SetValue(`//select[@id="permissions-add-select-role"]`, "write", chromedp.BySearch),
			chromedp.SetValue(`//select[@id="permissions-add-select-team"]`, team.Name(), chromedp.BySearch),
			chromedp.Click("#permissions-add-button", chromedp.NodeVisible),
			s.screenshot(t),
			chromedp.Text(".flash-success", &gotFlashSuccess, chromedp.NodeVisible),
		})
		require.NoError(t, err)

		assert.Equal(t, "owners", gotOwnersTeam)
		assert.Equal(t, "admin", gotOwnersRole)
		assert.Equal(t, "updated workspace permissions", gotFlashSuccess)
	})

	t.Run("list users", func(t *testing.T) {
		s := screenshotter(0)

		ctx, cancel := chromedp.NewContext(allocCtx)
		defer cancel()

		var gotUser string
		orgSelector := fmt.Sprintf("#item-organization-%s a", org.Name())
		userSelector := fmt.Sprintf("#item-user-%s .status", user.Username())
		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Navigate(url),
			// login
			chromedp.Click(".login-button-github", chromedp.NodeVisible),
			// select org
			chromedp.Click(orgSelector, chromedp.NodeVisible),
			s.screenshot(t),
			// list users
			chromedp.Click("#users > a", chromedp.NodeVisible),
			s.screenshot(t),
			chromedp.Text(userSelector, &gotUser, chromedp.NodeVisible),
		})
		require.NoError(t, err)

		assert.Equal(t, user.Username(), strings.TrimSpace(gotUser))
	})

	t.Run("list team members", func(t *testing.T) {
		s := screenshotter(0)

		ctx, cancel := chromedp.NewContext(allocCtx)
		defer cancel()

		var gotUser string
		orgSelector := fmt.Sprintf("#item-organization-%s a", org.Name())
		userSelector := fmt.Sprintf("#item-user-%s .status", user.Username())
		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Navigate(url),
			// login
			chromedp.Click(".login-button-github", chromedp.NodeVisible),
			// select org
			chromedp.Click(orgSelector, chromedp.NodeVisible),
			s.screenshot(t),
			// list teams
			chromedp.Click("#teams > a", chromedp.NodeVisible),
			s.screenshot(t),
			// select owners team
			chromedp.Click("#item-team-owners a", chromedp.NodeVisible),
			s.screenshot(t),
			chromedp.Text(userSelector, &gotUser, chromedp.NodeVisible),
		})
		require.NoError(t, err)

		assert.Equal(t, user.Username(), strings.TrimSpace(gotUser))
	})
}

func createWebWorkspace(t *testing.T, ctx context.Context, url string, org *otf.Organization) string {
	s := screenshotter(0)

	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	var gotFlashSuccess string
	workspaceName := "workspace-" + otf.GenerateRandomString(4)
	orgSelector := fmt.Sprintf("#item-organization-%s a", org.Name())

	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Click(".login-button-github", chromedp.NodeVisible),
		chromedp.Click(orgSelector, chromedp.NodeVisible),
		s.screenshot(t),
		chromedp.Click("#workspaces > a", chromedp.NodeVisible),
		chromedp.Click("#new-workspace-button", chromedp.NodeVisible),
		s.screenshot(t),
		chromedp.Focus("input#name", chromedp.NodeVisible),
		input.InsertText(workspaceName),
		chromedp.Click("#create-workspace-button"),
		s.screenshot(t),
		chromedp.Text(".flash-success", &gotFlashSuccess, chromedp.NodeVisible),
	})
	require.NoError(t, err)

	assert.Equal(t, "created workspace: "+workspaceName, strings.TrimSpace(gotFlashSuccess))

	return workspaceName
}

// screenshotter takes a series of screenshots for a test, incrementing a
// counter on each screenshot
type screenshotter int

func (ss *screenshotter) screenshot(t *testing.T) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		*ss++

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
		fname := path.Join("screenshots", fmt.Sprintf("%s_%02d.png", t.Name(), *ss))
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

func githubStub(t *testing.T, user *otf.User) string {
	srv := html.NewTestGithubServer(t, user)

	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	return u.Host
}
