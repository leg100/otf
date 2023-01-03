package e2e

import (
	"context"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

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
