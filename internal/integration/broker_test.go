package integration

import (
	"testing"

	"github.com/leg100/otf/internal/daemon"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/sql"
	"github.com/stretchr/testify/assert"
)

// TestBroker demonstrates publishing and subscribing of events via postgres.
func TestBroker(t *testing.T) {
	integrationTest(t)

	// simulate a cluster of two otfd nodes sharing a database
	cfg := config{
		Config: daemon.Config{Database: sql.NewTestDB(t)},
		// skip creating orgs which would otherwise send creation events
		skipDefaultOrganization: true,
	}
	local, _, ctx := setup(t, &cfg)
	remote, _, _ := setup(t, &cfg)

	// setup subscriptions
	localSub, localUnsub := local.WatchOrganizations()
	defer localUnsub()
	remoteSub, remoteUnsub := remote.WatchOrganizations()
	defer remoteUnsub()

	// create an org which should trigger an event
	org := local.createOrganization(t, ctx)
	want := pubsub.NewCreatedEvent(org)

	// receive event on local broker
	assert.Equal(t, want, <-localSub)
	// receive event on remote broker (via postgres)
	assert.Equal(t, want, <-remoteSub)
}
