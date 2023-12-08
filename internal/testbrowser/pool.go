package testbrowser

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/tokens"
	otfuser "github.com/leg100/otf/internal/user"
	"github.com/stretchr/testify/require"
)

const headlessEnvVar = "OTF_E2E_HEADLESS"

var poolSize = runtime.GOMAXPROCS(0)

// Pool of browsers
type Pool struct {
	pool chan *browser
	// service for creating new session in browser
	tokens *tokens.Service
	// allocator of browsers
	allocator context.Context
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

	allocator, cancelAllocator := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", headless),
			chromedp.Flag("hide-scrollbars", true),
			chromedp.Flag("mute-audio", true),
			chromedp.Flag("ignore-certificate-errors", true),
			chromedp.Flag("disable-gpu", true),
		)...)

	p := Pool{
		pool:      make(chan *browser, poolSize),
		tokens:    tokensService,
		allocator: allocator,
	}
	for i := 0; i < poolSize; i++ {
		p.pool <- nil
	}

	// Terminate all provisioned browsers and then terminate their allocator
	cleanup := func() {
		for i := 0; i < poolSize; i++ {
			if b := <-p.pool; b != nil {
				b.cancel()
			}
		}
		cancelAllocator()
	}

	return &p, cleanup, nil
}

func (p *Pool) Run(t *testing.T, user context.Context, actions ...chromedp.Action) {
	t.Helper()

	b := <-p.pool
	if b == nil {
		b = newBrowser(t, p.allocator)
	}
	// return browser back to pool after this method finishes
	defer func() { p.pool <- b }()

	// create a dedicated tab for test
	tab, cancel := chromedp.NewContext(b.ctx)
	defer cancel()

	// click OK on any browser javascript dialog boxes that pop up
	chromedp.ListenTarget(tab, func(ev any) {
		switch ev.(type) {
		case *page.EventJavascriptDialogOpening:
			go func() {
				err := chromedp.Run(tab, page.HandleJavaScriptDialog(true))
				require.NoError(t, err)
			}()
		}
	})

	// because browser is being re-used, cookies are cleared and a new session
	// is created for the calling user.
	resetAction := chromedp.ActionFunc(func(c context.Context) error {
		if err := network.ClearBrowserCookies().Do(c); err != nil {
			return err
		}
		if user != nil {
			user, err := otfuser.UserFromContext(user)
			if err != nil {
				return err
			}
			token, err := p.tokens.NewSessionToken(user.Username, internal.CurrentTimestamp(nil).Add(time.Hour))
			if err != nil {
				return err
			}
			err = network.SetCookie("session", token).WithDomain("127.0.0.1").Do(c)
			if err != nil {
				return err
			}
		}
		return nil
	})
	actions = append(chromedp.Tasks{resetAction}, actions...)
	err := chromedp.Run(tab, actions...)
	require.NoError(t, err)
}
