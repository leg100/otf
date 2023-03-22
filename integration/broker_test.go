package integration

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/pubsub"
	"github.com/leg100/otf/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBroker demonstrates publishing and subscribing of events via postgres.
func TestBroker(t *testing.T) {
	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &otf.Superuser{})
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(func() { cancel() })

	db := sql.NewTestDB(t)
	// local broker to which events are published
	local := pubsub.NewTestBroker(t, db)
	// remote broker from which events should be received
	remote := pubsub.NewTestBroker(t, db)

	// setup organization service to use local broker
	svc := organization.NewTestService(t, db, &organization.Options{
		Logger: logr.Discard(),
		DB:     db,
		Broker: local,
	})
	// register table with remote broker
	remote.Register("organization", svc)

	done := make(chan error)
	go func() {
		done <- local.Start(ctx)
	}()
	go func() {
		done <- remote.Start(ctx)
	}()

	localsub, err := local.Subscribe(ctx, "local-sub")
	require.NoError(t, err)
	remotesub, err := remote.Subscribe(ctx, "remote-sub")
	require.NoError(t, err)

	// sends event via local broker
	org, err := svc.CreateOrganization(ctx, organization.OrganizationCreateOptions{
		Name: otf.String(uuid.NewString()),
	})
	require.NoError(t, err)

	want := otf.Event{Type: otf.EventOrganizationCreated, Payload: org}
	// receive event on local broker
	assert.Equal(t, want, <-localsub)
	// receive event on remote broker (via postgres)
	assert.Equal(t, want, <-remotesub)

	cancel()
	assert.NoError(t, <-done)
	assert.NoError(t, <-done)
}
