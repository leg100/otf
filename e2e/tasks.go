package e2e

import (
	"context"
	"strings"
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
)

func createWorkspaceTasks(t *testing.T, hostname, org, name string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate("https://" + hostname + "/organizations/" + org),
		screenshot(t),
		chromedp.Click("#menu-item-workspaces > a", chromedp.ByQuery),
		// sometimes get stuck on this one...
		chromedp.Click("#new-workspace-button", chromedp.NodeVisible, chromedp.ByQuery),
		screenshot(t),
		chromedp.Focus("input#name", chromedp.NodeVisible),
		input.InsertText(name),
		chromedp.Click("#create-workspace-button"),
		screenshot(t),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var got string
			err := chromedp.Run(ctx, chromedp.Text(".flash-success", &got, chromedp.NodeVisible))
			if err != nil {
				return err
			}
			assert.Equal(t, "created workspace: "+name, strings.TrimSpace(got))
			return nil
		}),
	}
}
