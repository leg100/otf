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
	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/tokens"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/stretchr/testify/require"
)

const headlessEnvVar = "OTF_E2E_HEADLESS"

var poolSize = runtime.GOMAXPROCS(0)

// Pool of browsers
type Pool struct {
	pool chan *browser
	// Key for generating session tokens
	key jwk.Key
	// allocator of browsers
	allocator context.Context
}

func NewPool(secret []byte) (*Pool, func(), error) {
	key, err := jwk.FromRaw(secret)
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
		key:       key,
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

	err := chromedp.Run(b.ctx, chromedp.ActionFunc(func(c context.Context) error {
		// Always clear cookies first in case a previous test has left some behind
		if err := network.ClearBrowserCookies().Do(c); err != nil {
			return err
		}
		if user != nil {
			// Seed a session cookie for the given user context
			user, err := auth.UserFromContext(user)
			if err != nil {
				return err
			}
			token, err := tokens.NewSessionToken(p.key, user.Username, internal.CurrentTimestamp().Add(time.Hour))
			if err != nil {
				return err
			}
			err = network.SetCookie("session", token).WithDomain("127.0.0.1").Do(c)
			if err != nil {
				return err
			}
		}
		return nil
	}))
	require.NoError(t, err)

	err = chromedp.Run(b.ctx, actions...)
	require.NoError(t, err)
}
