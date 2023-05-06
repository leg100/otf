package daemon

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	internal "github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func TestDaemon_MissingSecretError(t *testing.T) {
	var missing *internal.MissingParameterError
	_, err := New(context.Background(), logr.Discard(), Config{})
	require.True(t, errors.As(err, &missing))
}
