package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/daemon"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/runner"
	"github.com/leg100/otf/internal/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDaemonShutdown tests the shutdown of the daemon.
func TestDaemonShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	config := daemon.NewConfig()
	config.Database = sql.NewTestDB(t)
	config.Secret = sharedSecret
	config.DisableLatestChecker = true

	d, err := daemon.New(ctx, logr.Discard(), config)
	require.NoError(t, err)
	t.Cleanup(d.Close)

	started := make(chan struct{})
	exited := make(chan error)
	go func() {
		exited <- d.Start(ctx, logr.Discard(), started)
	}()
	// Don't proceed until daemon has started.
	<-started

	// Shutdown daemon and check it exits cleanly
	cancel()
	assert.NoError(t, <-exited)

	// Check runner set an exit status before shutting down.
	ctx = authz.AddSubjectToContext(t.Context(), &authz.Superuser{Username: "shutdown-test"})
	runners, err := d.Runners.ListRunners(ctx, runner.ListOptions{})
	require.NoError(t, err)
	assert.Equal(t, 1, len(runners))
	assert.Equal(t, runner.RunnerExited, runners[0].Status)
}
