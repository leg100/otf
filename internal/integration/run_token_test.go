package integration

import (
	"context"
	"testing"

	internal "github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/tokens"
	"github.com/stretchr/testify/require"
)

func TestRunToken(t *testing.T) {
	t.Parallel()

	// perform all actions as superuser
	ctx := internal.AddSubjectToContext(context.Background(), &auth.SiteAdmin)

	t.Run("create", func(t *testing.T) {
		svc := setup(t, nil)
		org := svc.createOrganization(t, ctx)

		_, err := svc.CreateRunToken(ctx, tokens.CreateRunTokenOptions{
			Organization: &org.Name,
			RunID:        internal.String("run-123"),
		})
		require.NoError(t, err)
	})
}
