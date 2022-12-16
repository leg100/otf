package e2e

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/chromedp/chromedp"
)

// chromedp browser config
var allocator context.Context

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
