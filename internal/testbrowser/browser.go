// Package testbrowser provisions web browsers for tests
package testbrowser

import (
	"context"
	"testing"

	cdpbrowser "github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/require"
)

var (
	// permissions for chrome's clipboard
	clipboardReadPermission  = cdpbrowser.PermissionDescriptor{Name: "clipboard-read"}
	clipboardWritePermission = cdpbrowser.PermissionDescriptor{Name: "clipboard-write"}
)

type browser struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func newBrowser(t *testing.T, allocator context.Context) *browser {
	ctx, cancel := chromedp.NewContext(allocator)
	err := chromedp.Run(ctx, chromedp.Tasks{
		cdpbrowser.SetPermission(&clipboardReadPermission, cdpbrowser.PermissionSettingGranted).WithOrigin(""),
		cdpbrowser.SetPermission(&clipboardWritePermission, cdpbrowser.PermissionSettingGranted).WithOrigin(""),
	})
	require.NoError(t, err)

	return &browser{ctx: ctx, cancel: cancel}
}
