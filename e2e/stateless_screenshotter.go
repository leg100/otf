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

// screenshotRecord maps test name to a count of number of screenshots taken
var (
	screenshotRecord map[string]int
	screenshotMutex  sync.Mutex
)

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
