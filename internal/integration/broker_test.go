package integration

import (
	"testing"

	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/sql"
	"github.com/stretchr/testify/assert"
)

// TestBroker demonstrates publishing and subscribing of events via postgres,
// and tests that events are received on each node in a multiple node cluster.
func TestBroker(t *testing.T) {
	integrationTest(t)

	// simulate a cluster of two otfd nodes sharing a database
	db := withDatabase(sql.NewTestDB(t))
	local, org, ctx := setup(t, db)
	remote, _, _ := setup(t, db)

	// setup subscriptions
	localSub, localUnsub := local.Workspaces.Watch(ctx)
	defer localUnsub()
	remoteSub, remoteUnsub := remote.Workspaces.Watch(ctx)
	defer remoteUnsub()

	// create a workspace which should trigger an event
	ws := local.createWorkspace(t, ctx, org)

	// receive identical event from both local and remote brokers
	localEvent := <-localSub
	remoteEvent := <-remoteSub
	assert.Equal(t, localEvent, remoteEvent)

	assert.Equal(t, pubsub.CreatedEvent, localEvent.Type)
	assert.Equal(t, ws.ID, localEvent.Payload.ID)
}
