package integration

import (
	"testing"

	"github.com/leg100/otf/internal/run"
	"github.com/stretchr/testify/require"
)

func TestRunToken(t *testing.T) {
	integrationTest(t)

	t.Run("create", func(t *testing.T) {
		svc, org, ctx := setup(t, nil)

		_, err := svc.CreateRunToken(ctx, run.CreateRunTokenOptions{
			Organization: &org.Name,
		})
		require.NoError(t, err)
	})
}
