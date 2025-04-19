package integration

import (
	"testing"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/runner"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunner_Service(t *testing.T) {
	integrationTest(t)

	// Create database for multiple daemons to share.
	db := withDatabase(sql.NewTestDB(t))

	// Start first daemon, which creates a runner too.
	daemon, _, ctx := setup(t, db)

	// Create 3 more daemons, which creates 3 more runners
	for range 3 {
		_, _, _ = setup(t, db)
	}

	// listing site-wide runners requires site admin perms.
	ctx = authz.AddSubjectToContext(ctx, &user.SiteAdmin)

	results, err := daemon.Runners.List(ctx, runner.ListOptions{})
	require.NoError(t, err)

	assert.Equal(t, 4, len(results.Items))
	assert.Equal(t, 4, results.TotalCount)
}
