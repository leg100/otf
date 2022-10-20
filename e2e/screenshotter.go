package e2e

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"
	"testing"

	"github.com/chromedp/chromedp"
)

// screenshotter takes and persists screenshots
type screenshotter struct {
	// map of test name to counter
	m  map[string]int
	mu sync.Mutex
}

func (ss *screenshotter) screenshot(t *testing.T) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		ss.mu.Lock()
		defer ss.mu.Unlock()

		counter, ok := ss.m[t.Name()]
		if !ok {
			ss.m[t.Name()] = 0
		}
		ss.m[t.Name()]++

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
