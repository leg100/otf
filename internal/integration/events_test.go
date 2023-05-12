package integration

import (
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_Events demonstrates events are triggered and successfully
// received by a subscriber.
func TestIntegration_Events(t *testing.T) {
	t.Parallel()

	daemon := setup(t, nil)
	sub, err := daemon.Subscribe(ctx, "")
	require.NoError(t, err)

	org := daemon.createOrganization(t, ctx)
	ws := daemon.createWorkspace(t, ctx, org)
	cv := daemon.createAndUploadConfigurationVersion(t, ctx, ws)
	run := daemon.createRun(t, ctx, ws, cv)

	assert.Equal(t, internal.NewCreatedEvent(org), <-sub)
	assert.Equal(t, internal.NewCreatedEvent(ws), <-sub)
	assert.Equal(t, internal.NewCreatedEvent(run), <-sub)
	assert.Equal(t, internal.NewUpdatedEvent(run), <-sub)
	assert.Equal(t, internal.NewUpdatedEvent(run), <-sub)
	assert.Equal(t, internal.NewUpdatedEvent(run), <-sub)
	assert.Equal(t, internal.NewUpdatedEvent(run), <-sub)
}
