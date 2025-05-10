package integration

import (
	"testing"

	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/sql"
	"github.com/stretchr/testify/assert"
)

// TestBroker demonstrates publishing and subscribing of events via postgres.
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
	want := pubsub.NewCreatedEvent(ws)

	// receive event on local broker
	assert.Equal(t, want, <-localSub)
	// receive event on remote broker (via postgres)
	assert.Equal(t, want, <-remoteSub)
}
