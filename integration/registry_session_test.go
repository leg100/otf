package integration

import (
	"context"
	"testing"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistrySession(t *testing.T) {
	t.Parallel()

	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &otf.Superuser{})

	t.Run("create", func(t *testing.T) {
		svc := setup(t, nil)
		org := svc.createOrganization(t, ctx)

		_, err := svc.CreateRegistrySession(ctx, auth.CreateRegistrySessionOptions{
			Organization: &org.Name,
		})
		require.NoError(t, err)
	})

	t.Run("get", func(t *testing.T) {
		svc := setup(t, nil)
		want := svc.createRegistrySession(t, ctx, nil, nil)

		got, err := svc.GetRegistrySession(ctx, want.Token)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("cleanup", func(t *testing.T) {
		svc := setup(t, nil)
		rs1 := svc.createRegistrySession(t, ctx, nil, otf.Time(time.Now()))
		rs2 := svc.createRegistrySession(t, ctx, nil, otf.Time(time.Now()))

		ctx, cancel := context.WithCancel(ctx)
		svc.StartExpirer(ctx)
		defer cancel()

		_, err := svc.GetRegistrySession(ctx, rs1.Token)
		assert.Equal(t, otf.ErrResourceNotFound, err)

		_, err = svc.GetRegistrySession(ctx, rs2.Token)
		assert.Equal(t, otf.ErrResourceNotFound, err)
	})
}
