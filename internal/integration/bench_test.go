package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/daemon"
	"github.com/leg100/otf/internal/logr"
	"golang.org/x/sync/errgroup"
)

func Benchmark(b *testing.B) {
	ctx := context.Background()

	logger, _ := logr.New(&logr.Config{Verbosity: 0, Format: "default"})
	cfg := daemon.Config{
		Database: "postgres:///otf",
		Secret:   sharedSecret,
	}
	daemon.ApplyDefaults(&cfg)
	d, err := daemon.New(ctx, logger, cfg)
	panicIfErr(err)

	g, ctx := errgroup.WithContext(ctx)
	started := make(chan struct{})
	g.Go(func() error {
		return d.Start(ctx, started)
	})
	// don't proceed until daemon has started.
	<-started

	if err := g.Wait(); err != nil {
		b.Logf("daemon exited with error: %s", err.Error())
	}
}

func panicIfErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}
