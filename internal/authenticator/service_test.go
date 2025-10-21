package authenticator

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
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
