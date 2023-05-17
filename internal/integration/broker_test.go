package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/daemon"
	"github.com/leg100/otf/internal/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBroker demonstrates publishing and subscribing of events via postgres.
func TestBroker(t *testing.T) {
	t.Parallel()

	// perform all actions as superuser
	ctx := internal.AddSubjectToContext(context.Background(), &auth.SiteAdmin)

	// simulate a cluster of two otfd nodes sharing a database
	connstr := sql.NewTestDB(t)
	local := setup(t, &config{Config: daemon.Config{Database: connstr}})
	remote := setup(t, &config{Config: daemon.Config{Database: connstr}})

	localsub, err := local.Subscribe(ctx, "")
	require.NoError(t, err)
	remotesub, err := remote.Subscribe(ctx, "")
	require.NoError(t, err)

	// sends event via local broker
	org := local.createOrganization(t, ctx)

	want := internal.NewCreatedEvent(org)
	// receive event on local broker
	assert.Equal(t, want, <-localsub)
	// receive event on remote broker (via postgres)
	assert.Equal(t, want, <-remotesub)
}
