package testutil

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/stretchr/testify/require"
)

func NewAuthService(t *testing.T, db otf.DB) *auth.Service {
	service, err := auth.NewService(context.Background(), auth.Options{
		Authorizer: NewAllowAllAuthorizer(),
		DB:         db,
		Logger:     logr.Discard(),
	})
	require.NoError(t, err)
	return service
}
