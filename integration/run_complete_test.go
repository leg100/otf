package integration

import (
	"context"
	"os"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/run"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

// TestCompleteRun tests a terraform run from start to finish.
func TestCompleteRun(t *testing.T) {
	t.Parallel()

	svc := setup(t, nil)

	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &otf.Superuser{})
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	agent, err := svc.NewAgent(logr.Discard())
	require.NoError(t, err)

	// Start necessary processes
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return svc.StartScheduler(ctx, logr.Discard(), svc.db) })
	g.Go(func() error { return agent.Start(ctx) })
	g.Go(func() error { return svc.Broker.Start(ctx) })
	terminated := make(chan error)
	go func() { terminated <- g.Wait() }()

	sub, err := svc.Subscribe(ctx, "")
	require.NoError(t, err)

	ws := svc.createWorkspace(t, ctx, nil)
	cv := svc.createConfigurationVersion(t, ctx, ws)
	tarball, err := os.ReadFile("./testdata/root.tar.gz")
	require.NoError(t, err)
	svc.UploadConfig(ctx, cv.ID, tarball)

	_ = svc.createRun(t, ctx, ws, cv)

	for {
		select {
		case event := <-sub:
			if r, ok := event.Payload.(*run.Run); ok {
				switch r.Status {
				case otf.RunErrored:
					t.Fatal("run unexpectedly errored")
				case otf.RunPlanned:
					err = svc.Apply(ctx, r.ID)
					require.NoError(t, err)
				case otf.RunApplied:
					return // success
				}
			}
		case err := <-terminated:
			t.Fatalf("process terminated with error: %s", err.Error())
		}
	}
}
