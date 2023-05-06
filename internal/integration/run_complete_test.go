package integration

import (
	"context"
	"os"
	"testing"

	internal "github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/run"
	"github.com/stretchr/testify/require"
)

// TestCompleteRun tests a terraform run from start to finish.
func TestCompleteRun(t *testing.T) {
	t.Parallel()

	svc := setup(t, nil)

	ctx := internal.AddSubjectToContext(context.Background(), &auth.SiteAdmin)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sub, err := svc.Subscribe(ctx, "")
	require.NoError(t, err)

	ws := svc.createWorkspace(t, ctx, nil)
	cv := svc.createConfigurationVersion(t, ctx, ws)
	tarball, err := os.ReadFile("./testdata/root.tar.gz")
	require.NoError(t, err)
	svc.UploadConfig(ctx, cv.ID, tarball)

	_ = svc.createRun(t, ctx, ws, cv)

	for event := range sub {
		if r, ok := event.Payload.(*run.Run); ok {
			switch r.Status {
			case internal.RunErrored:
				t.Fatal("run unexpectedly errored")
			case internal.RunPlanned:
				err = svc.Apply(ctx, r.ID)
				require.NoError(t, err)
			case internal.RunApplied:
				return // success
			}
		}
	}
}
