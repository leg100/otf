// Package browser provides interaction with web browsers for the purposes of
// testing
package browser

import (
	"context"
	"runtime"
	"testing"
)

// Pool of browsers
type Pool struct {
	pool chan context.Context
}

func NewPool(t *testing.T) Pool {
	p := Pool{
		pool: make(chan context.Context, runtime.GOMAXPROCS(0)),
	}
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		p.pool <- nil
	}
	return p
}
