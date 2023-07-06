package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/tokens"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/stretchr/testify/require"
)

func TestCreateWorkspace(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t, nil)

	allocator, cancelAllocator := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("hide-scrollbars", true),
			chromedp.Flag("mute-audio", true),
			chromedp.Flag("ignore-certificate-errors", true),
			chromedp.Flag("disable-gpu", true),
		)...)
	defer cancelAllocator()

	browser, cancel := chromedp.NewContext(allocator)
	defer cancel()

	key, err := jwk.FromRaw(daemon.Secret)
	require.NoError(t, err)

	err = chromedp.Run(browser,
		chromedp.ActionFunc(func(c context.Context) error {
			// Always clear cookies first in case a previous test has left some behind
			if err := network.ClearBrowserCookies().Do(c); err != nil {
				return err
			}
			// Seed a session cookie for the given user context
			user, err := auth.UserFromContext(ctx)
			if err != nil {
				return err
			}
			token, err := tokens.NewSessionToken(key, user.Username, internal.CurrentTimestamp().Add(time.Hour))
			if err != nil {
				return err
			}
			err = network.SetCookie("session", token).WithDomain("127.0.0.1").Do(c)
			if err != nil {
				return err
			}
			return nil
		}),
	)
	require.NoError(t, err)

	for i := 0; i < 100; i++ {
		workspace := fmt.Sprintf("ws-%d", i)
		t.Run(workspace, func(t *testing.T) {

			ctx, cancel = chromedp.NewContext(browser)
			defer cancel()

			err := chromedp.Run(ctx,
				chromedp.ActionFunc(func(c context.Context) error {
					err := chromedp.Navigate(workspacesURL(daemon.Hostname(), org.Name)).Do(c)
					if err != nil {
						return fmt.Errorf("navigating to workspaces url: %w", err)
					}
					return nil
				}),
				chromedp.ActionFunc(func(c context.Context) error {
					err := chromedp.Click("//a[text()='organizations']", chromedp.BySearch).Do(c)
					if err != nil {
						return fmt.Errorf("clicking organizations link: %w", err)
					}
					return nil
				}),
			)
			require.NoError(t, err)
		})
	}
}
