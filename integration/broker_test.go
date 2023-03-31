package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBroker demonstrates publishing and subscribing of events via postgres.
func TestBroker(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// simulate a cluster of two otfd nodes sharing a connstr
	connstr := sql.NewTestDB(t)
	local := setup(t, &config{connstr: &connstr})
	remote := setup(t, &config{connstr: &connstr})

	// wait 'til brokers are listening
	local.Broker.WaitUntilListening()
	remote.Broker.WaitUntilListening()

	localsub, err := local.Subscribe(ctx, "")
	require.NoError(t, err)
	remotesub, err := remote.Subscribe(ctx, "")
	require.NoError(t, err)

	// sends event via local broker
	org := local.createOrganization(t, ctx)

	want := otf.Event{Type: otf.EventOrganizationCreated, Payload: org}
	// receive event on local broker
	assert.Equal(t, want, <-localsub)
	// receive event on remote broker (via postgres)
	assert.Equal(t, want, <-remotesub)
}
