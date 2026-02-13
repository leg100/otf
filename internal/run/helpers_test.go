package run

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/require"
)

func newTestRun(t *testing.T, ctx context.Context, opts CreateOptions) *Run {
	org, err := organization.NewOrganization(organization.CreateOptions{Name: new("acme-corp")})
	require.NoError(t, err)

	ws := workspace.NewTestWorkspace(t, nil)
	cv := configversion.NewConfigurationVersion(ws.ID, configversion.CreateOptions{})

	factory := newTestFactory(org, ws, cv)

	run, err := factory.NewRun(ctx, ws.ID, opts)
	require.NoError(t, err)

	return run
}
