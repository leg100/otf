package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/require"
)

func TestSync(t *testing.T) {
	t.Parallel()

	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &auth.SiteAdmin)

	svc := setup(t, nil)
	err := svc.Sync(ctx, cloud.User{Name: "bobby"})
	require.NoError(t, err)
}
