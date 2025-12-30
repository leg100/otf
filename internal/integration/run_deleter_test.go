package integration

import (
	"testing"
	"time"

	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/run"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunDeleter tests the run deleter subsystem, which deletes runs older than
// a user-specified time period.
func TestRunDeleter(t *testing.T) {
	integrationTest(t)

	// Delete runs older than 1 hour, and check runs every second.
	daemon, _, ctx := setup(t, withDeleteRunsAfter(time.Hour, time.Second))

	yesterday := time.Now().Add(-time.Hour * 24)

	// Create two old runs that should be deleted, and two new runs that should
	// not be deleted.
	old1 := daemon.createRun(t, ctx, nil, nil, &run.CreateOptions{CreatedAt: &yesterday})
	old2 := daemon.createRun(t, ctx, nil, nil, &run.CreateOptions{CreatedAt: &yesterday})
	_ = daemon.createRun(t, ctx, nil, nil, nil)
	_ = daemon.createRun(t, ctx, nil, nil, nil)

	var (
		old1Deleted bool
		old2Deleted bool
	)

	timeout := time.After(5 * time.Second)

	for {
		select {
		case <-timeout:
			t.Fatal("runs were not deleted within timeout")
		case event := <-daemon.runEvents:
			if event.Type == pubsub.DeletedEvent {
				switch event.Payload.ID {
				case old1.ID:
					old1Deleted = true
				case old2.ID:
					old2Deleted = true
				}
			}
		}
		if old1Deleted && old2Deleted {
			break
		}
	}

	// Listing runs site-wide requires site admin user
	runs, err := daemon.Runs.List(adminCtx, run.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, runs.Items, 2)
}
