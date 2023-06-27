package integration

import (
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/tokens"
	"github.com/stretchr/testify/require"
)

func TestRunToken(t *testing.T) {
	integrationTest(t)

	t.Run("create", func(t *testing.T) {
		svc, org, ctx := setup(t, nil)

		_, err := svc.CreateRunToken(ctx, tokens.CreateRunTokenOptions{
			Organization: &org.Name,
			RunID:        internal.String("run-123"),
		})
		require.NoError(t, err)
	})
}
