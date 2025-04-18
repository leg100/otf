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
	// skip creating orgs which would otherwise send creation events
	skipOrg := skipDefaultOrganization()
	local, _, ctx := setup(t, db, skipOrg)
	remote, _, _ := setup(t, db, skipOrg)

	// setup subscriptions
	localSub, localUnsub := local.Organizations.WatchOrganizations(ctx)
	defer localUnsub()
	remoteSub, remoteUnsub := remote.Organizations.WatchOrganizations(ctx)
	defer remoteUnsub()

	// create an org which should trigger an event
	org := local.createOrganization(t, ctx)
	want := pubsub.NewCreatedEvent(org)

	// receive event on local broker
	assert.Equal(t, want, <-localSub)
	// receive event on remote broker (via postgres)
	assert.Equal(t, want, <-remoteSub)
}
