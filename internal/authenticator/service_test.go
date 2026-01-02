package authenticator

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/stretchr/testify/require"
)

func TestNewAuthenticatorService(t *testing.T) {
	opts := Options{
		Logger:          logr.Discard(),
		HostnameService: internal.NewHostnameService("fake-host.org"),
	}
	_, err := NewAuthenticatorService(context.Background(), opts)
	require.NoError(t, err)
}
