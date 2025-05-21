// Package testbrowser provides browsers for e2e tests
package testbrowser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/tokens"
	otfuser "github.com/leg100/otf/internal/user"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

const headlessEnvVar = "OTF_E2E_HEADLESS"

var poolSize = runtime.GOMAXPROCS(0)

// Pool of browsers
type Pool struct {
	// browser shared by pool of contexts
	browser playwright.Browser
	// pool of contexts, with isolated cookie store, data dir, etc
	pool chan playwright.BrowserContext
	// service for creating new session in browser
	tokens *tokens.Service
}

func NewPool(secret []byte) (*Pool, func(), error) {
	tokensService, err := tokens.NewService(tokens.Options{Secret: secret})
	if err != nil {
		return nil, nil, err
	}

	// Headless mode determines whether browser window is displayed (false) or
	// not (true).
	headless := true
	if v, ok := os.LookupEnv(headlessEnvVar); ok {
		var err error
		headless, err = strconv.ParseBool(v)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing %s: %w", headlessEnvVar, err)
		}
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, nil, fmt.Errorf("running playwright: %w", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: &headless,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("launching chromium: %w", err)
	}

	p := Pool{
		pool:    make(chan playwright.BrowserContext, poolSize),
		tokens:  tokensService,
		browser: browser,
	}
	for i := 0; i < poolSize; i++ {
		p.pool <- nil
	}

	// Terminate all provisioned browsers and then terminate their allocator
	cleanup := func() {
		for i := 0; i < poolSize; i++ {
			if b := <-p.pool; b != nil {
				// TODO: handle error
				_ = b.Close()
			}
		}
		// TODO: handle error
		_ = browser.Close()
	}

	return &p, cleanup, nil
}

// New requests a page from the browser pool.
func (p *Pool) New(t *testing.T, user context.Context, fn func(playwright.Page)) {
	t.Helper()

	// Wait for available context from pool
	<-p.pool

	// Construct new ctx
	browserCtx, err := p.browser.NewContext(playwright.BrowserNewContextOptions{
		IgnoreHttpsErrors: internal.Bool(true),
	})
	require.NoError(t, err)

	// When test has finished, close ctx and make available space in pool
	defer func() {
		p.pool <- nil
	}()
	defer browserCtx.Close()

	err = browserCtx.GrantPermissions([]string{
		"clipboard-read",
		"clipboard-write",
	})
	require.NoError(t, err)

	// Create a browser page (tab) for test
	page, err := browserCtx.NewPage()
	require.NoError(t, err)

	// Close page when done with page
	defer func() {
		err := page.Close()
		require.NoError(t, err)
	}()

	// In the event of a failure take a screenshot for debugging purposes.
	defer func() {
		if t.Failed() {
			fname := fmt.Sprintf("%s_failure.png", t.Name())
			path := filepath.Join("screenshots", fname)

			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Logf("failed to make screenshots directory: %s", err.Error())
				return
			}

			_, err = page.Screenshot(playwright.PageScreenshotOptions{Path: &path})
			if err != nil {
				t.Logf("failed to take screenshot: %s", err.Error())
			}
		}
	}()

	// Click OK on any browser javascript dialog boxes that pop up
	page.OnDialog(func(dialog playwright.Dialog) {
		dialog.Accept()
	})

	// Populate cookie jar with session token if user is specified.
	if user != nil {
		user, err := otfuser.UserFromContext(user)
		require.NoError(t, err)

		token, err := p.tokens.NewToken(user.ID, tokens.WithExpiry(
			internal.CurrentTimestamp(nil).Add(time.Hour),
		))
		require.NoError(t, err)

		err = browserCtx.AddCookies([]playwright.OptionalCookie{
			{
				Name:   "session",
				Value:  string(token),
				Domain: internal.String("127.0.0.1"),
				Path:   internal.String("/"),
			},
		})
		require.NoError(t, err)
	}

	fn(page)

}
