package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/run"
	"github.com/stretchr/testify/require"
)

// TestCompleteRun tests a terraform run from start to finish.
func TestCompleteRun(t *testing.T) {
	integrationTest(t)

	daemon, _, ctx := setup(t, nil)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ws := daemon.createWorkspace(t, ctx, nil)
	cv := daemon.createAndUploadConfigurationVersion(t, ctx, ws, nil)

	_ = daemon.createRun(t, ctx, ws, cv)

	for event := range daemon.sub {
		if r, ok := event.Payload.(*run.Run); ok {
			switch r.Status {
			case internal.RunErrored:
				t.Fatal("run unexpectedly errored")
			case internal.RunPlanned:
				err := daemon.Apply(ctx, r.ID)
				require.NoError(t, err)
			case internal.RunApplied:
				return // success
			}
		}
	}
}
